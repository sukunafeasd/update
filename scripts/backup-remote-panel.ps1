param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [string]$OutputDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) "backups\\remote"),
  [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($OpsToken)) {
  throw "UNIVERSALD_OPS_TOKEN obrigatorio para baixar o backup remoto."
}

New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$archivePath = Join-Path $OutputDir ("painel-dief-remote-" + $stamp + ".zip")
$headers = @{ Authorization = "Bearer $OpsToken" }

Invoke-WebRequest -Uri ($BaseUrl.TrimEnd("/") + "/api/ops/export") -Headers $headers -OutFile $archivePath -UseBasicParsing

if ($RetentionDays -gt 0) {
  $cutoff = (Get-Date).AddDays(-1 * $RetentionDays)
  Get-ChildItem $OutputDir -Filter "painel-dief-remote-*.zip" -File -ErrorAction SilentlyContinue |
    Where-Object { $_.LastWriteTime -lt $cutoff } |
    Remove-Item -Force -ErrorAction SilentlyContinue
}

Write-Output $archivePath
