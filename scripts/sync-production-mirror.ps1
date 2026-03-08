param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [string]$OutputDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) "backups\\remote"),
  [string]$MirrorDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) ".runtime\\production-mirror"),
  [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$backupScript = Join-Path $root "scripts\\backup-remote-panel.ps1"
$restoreScript = Join-Path $root "scripts\\restore-panel.ps1"

$archivePath = & powershell -ExecutionPolicy Bypass -File $backupScript -BaseUrl $BaseUrl -OpsToken $OpsToken -OutputDir $OutputDir -RetentionDays $RetentionDays
$archivePath = @($archivePath | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) | Select-Object -Last 1
if ([string]::IsNullOrWhiteSpace($archivePath) -or -not (Test-Path $archivePath)) {
  throw "Nao consegui baixar o backup remoto da producao."
}

& powershell -ExecutionPolicy Bypass -File $restoreScript -ArchivePath $archivePath -DataDir $MirrorDir -Force | Out-Null

Write-Output $archivePath
Write-Output $MirrorDir
