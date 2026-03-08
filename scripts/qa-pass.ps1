param(
  [string]$GoExe = "C:\Program Files\Go\bin\go.exe",
  [string]$ProductionUrl = "https://update-wrl6.onrender.com",
  [string]$DesktopProject = "C:\Users\cafe\Desktop\universalD",
  [string]$DesktopExe = "C:\Users\cafe\Desktop\universalD.exe",
  [string]$OwnerPassword = $env:PAINEL_DIEF_OWNER_PASSWORD,
  [switch]$IncludeDesktop
)

$ErrorActionPreference = "Stop"

$siteRoot = Split-Path -Parent $MyInvocation.MyCommand.Path | Split-Path -Parent
$siteTmpDir = Join-Path $siteRoot ".runtime\tmp"
$siteExe = Join-Path $siteTmpDir "universald.exe"
Set-Location $siteRoot

Write-Host "[1/5] go test ./..."
& $GoExe test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[2/5] go build ./cmd/universald (tmp)"
New-Item -ItemType Directory -Force -Path $siteTmpDir | Out-Null
if (Test-Path $siteExe) {
  Remove-Item $siteExe -Force -ErrorAction SilentlyContinue
}
& $GoExe build -o $siteExe ./cmd/universald
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[3/5] smoke production"
& powershell -ExecutionPolicy Bypass -File (Join-Path $siteRoot "scripts\smoke-production.ps1") -BaseUrl $ProductionUrl
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if ($IncludeDesktop -and (Test-Path $DesktopProject)) {
  Write-Host "[4/5] build desktop"
  Set-Location $DesktopProject
  & $GoExe build ./cmd/universald-desktop
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

if ($IncludeDesktop -and (Test-Path $DesktopExe)) {
  Write-Host "[5/5] smoke desktop"
  $desktopSmokeArgs = @(
    "-ExecutionPolicy", "Bypass",
    "-File", (Join-Path $DesktopProject "scripts\smoke-desktop.ps1"),
    "-ExePath", $DesktopExe
  )
  if (-not [string]::IsNullOrWhiteSpace($OwnerPassword)) {
    $desktopSmokeArgs += @("-Password", $OwnerPassword)
  }
  & powershell @desktopSmokeArgs
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} elseif (-not $IncludeDesktop) {
  Write-Host "[4/5] desktop QA pulado (site-only)"
}

if (Test-Path $siteExe) {
  Remove-Item $siteExe -Force -ErrorAction SilentlyContinue
}
