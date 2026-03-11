param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [int]$IntervalSec = 30,
  [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($env:UNIVERSALD_OPS_TOKEN)) {
  throw "UNIVERSALD_OPS_TOKEN obrigatorio para iniciar o guardiao."
}

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$runtimeDir = Join-Path $root ".runtime"
$stateFile = Join-Path $runtimeDir "backup-watch-state.json"
$pidFile = Join-Path $runtimeDir "backup-watch-process.json"
$watchScript = Join-Path $root "scripts\watch-production-backup.ps1"

New-Item -ItemType Directory -Path $runtimeDir -Force | Out-Null

if (Test-Path $pidFile) {
  try {
    $existing = Get-Content $pidFile -Raw | ConvertFrom-Json
    if ($existing.pid) {
      $proc = Get-Process -Id ([int]$existing.pid) -ErrorAction SilentlyContinue
      if ($null -ne $proc) {
        Write-Output ("guardiao ja esta rodando no pid " + $existing.pid)
        exit 0
      }
    }
  } catch {
  }
  Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
}

$args = @(
  "-ExecutionPolicy", "Bypass",
  "-File", $watchScript,
  "-BaseUrl", $BaseUrl,
  "-IntervalSec", [string]$IntervalSec,
  "-RetentionDays", [string]$RetentionDays
)

$proc = Start-Process -FilePath "powershell" -ArgumentList $args -WindowStyle Hidden -PassThru
$payload = @{
  pid = $proc.Id
  startedAt = (Get-Date).ToUniversalTime().ToString("o")
  baseUrl = $BaseUrl
  intervalSec = $IntervalSec
  stateFile = $stateFile
} | ConvertTo-Json -Depth 4
$payload | Set-Content -Path $pidFile -Encoding UTF8

Write-Output ("guardiao iniciado no pid " + $proc.Id)
