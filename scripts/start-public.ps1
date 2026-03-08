param(
  [string]$AppEnv = "development",
  [string]$PublicOrigin = "",
  [string]$OpsToken = "",
  [int]$MaintenanceIntervalSec = 60
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$runtimeDir = Join-Path $root ".runtime"
$serverLog = Join-Path $runtimeDir "server.log"
$serverErr = Join-Path $runtimeDir "server.err.log"
$tunnelLog = Join-Path $runtimeDir "tunnel.log"
$tunnelErr = Join-Path $runtimeDir "tunnel.err.log"
$pidFile = Join-Path $runtimeDir "pids.json"
$dbPath = Join-Path $root "universald.db"
$uploadsDir = Join-Path $root "panel_uploads"
$bindUrl = "http://127.0.0.1:7788"
$cloudflared = Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages\Cloudflare.cloudflared_Microsoft.Winget.Source_8wekyb3d8bbwe\cloudflared.exe"

New-Item -ItemType Directory -Path $runtimeDir -Force | Out-Null
New-Item -ItemType Directory -Path $uploadsDir -Force | Out-Null

$env:UNIVERSALD_APP_ENV = $AppEnv
$env:UNIVERSALD_PUBLIC_ORIGIN = $PublicOrigin
$env:UNIVERSALD_OPS_TOKEN = $OpsToken
$env:UNIVERSALD_MAINTENANCE_INTERVAL_SEC = [string]$MaintenanceIntervalSec

if (-not (Test-Path $cloudflared)) {
  throw "cloudflared.exe nao encontrado em $cloudflared"
}

if (Test-Path $pidFile) {
  $meta = Get-Content $pidFile | ConvertFrom-Json
  if ($meta.serverPid) {
    Stop-Process -Id ([int]$meta.serverPid) -Force -ErrorAction SilentlyContinue
  }
  if ($meta.tunnelPid) {
    Stop-Process -Id ([int]$meta.tunnelPid) -Force -ErrorAction SilentlyContinue
  }
  Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
}

$portOwner = Get-NetTCPConnection -LocalPort 7788 -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($portOwner) {
  Stop-Process -Id $portOwner -Force -ErrorAction SilentlyContinue
}

Get-Process | Where-Object { $_.Path -like "$root*" } | Stop-Process -Force -ErrorAction SilentlyContinue

Set-Content -Path $serverLog -Value "" -Encoding UTF8
Set-Content -Path $serverErr -Value "" -Encoding UTF8
Set-Content -Path $tunnelLog -Value "" -Encoding UTF8
Set-Content -Path $tunnelErr -Value "" -Encoding UTF8

$server = Start-Process go `
  -ArgumentList @("run", ".\cmd\universald", "-open=false", "-bind", "127.0.0.1:7788", "-db", $dbPath, "-uploads", $uploadsDir, "-web", (Join-Path $root "web")) `
  -WorkingDirectory $root `
  -RedirectStandardOutput $serverLog `
  -RedirectStandardError $serverErr `
  -PassThru `
  -WindowStyle Hidden

for ($i = 0; $i -lt 25; $i++) {
  try {
    $ping = Invoke-WebRequest -Uri $bindUrl -UseBasicParsing -TimeoutSec 2
    if ($ping.StatusCode -ge 200) {
      break
    }
  } catch {}
  Start-Sleep -Seconds 1
}

$tunnel = Start-Process $cloudflared `
  -ArgumentList @("tunnel", "--url", $bindUrl, "--no-autoupdate") `
  -WorkingDirectory $root `
  -RedirectStandardOutput $tunnelLog `
  -RedirectStandardError $tunnelErr `
  -PassThru `
  -WindowStyle Hidden

$meta = [ordered]@{
  serverPid = $server.Id
  tunnelPid = $tunnel.Id
  startedAt = (Get-Date).ToString("o")
  version = "1.4.4"
  appEnv = $AppEnv
  publicOrigin = $PublicOrigin
  opsSecured = [bool](-not [string]::IsNullOrWhiteSpace($OpsToken))
  maintenanceIntervalSec = $MaintenanceIntervalSec
  bindUrl = $bindUrl
  dbPath = $dbPath
  uploadsDir = $uploadsDir
}
$meta | ConvertTo-Json | Set-Content $pidFile -Encoding UTF8

Start-Sleep -Seconds 8

$url = $null
for ($i = 0; $i -lt 30; $i++) {
  foreach ($candidate in @($tunnelErr, $tunnelLog)) {
    if (Test-Path $candidate) {
      $match = Select-String -Path $candidate -Pattern "https://[-a-z0-9]+\.trycloudflare\.com" -AllMatches | Select-Object -Last 1
      if ($match) {
        $url = $match.Matches.Value | Select-Object -Last 1
        break
      }
    }
  }
  if ($url) { break }
  Start-Sleep -Seconds 1
}

if (-not $url) {
  throw "Nao consegui capturar a URL publica. Veja $tunnelErr"
}

$meta["publicUrl"] = $url
$meta | ConvertTo-Json | Set-Content $pidFile -Encoding UTF8

Write-Output $url
