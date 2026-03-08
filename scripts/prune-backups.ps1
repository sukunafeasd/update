param(
  [string]$Directory = "",
  [int]$Keep = 4
)

$ErrorActionPreference = "Stop"

if ($Keep -lt 1) {
  throw "Keep precisa ser pelo menos 1."
}

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if ([string]::IsNullOrWhiteSpace($Directory)) {
  $Directory = Join-Path $root "backups\remote"
}

if (-not (Test-Path $Directory)) {
  Write-Host "Diretorio de backup nao existe: $Directory"
  exit 0
}

$files = Get-ChildItem $Directory -File | Sort-Object LastWriteTime -Descending
if ($files.Count -le $Keep) {
  Write-Host "Nada para limpar. Total atual: $($files.Count)"
  exit 0
}

$removed = @()
$files | Select-Object -Skip $Keep | ForEach-Object {
  Remove-Item -Force $_.FullName
  $removed += $_.Name
}

Write-Host ("Removidos: " + ($removed -join ", "))
Write-Host ("Mantidos: " + $Keep)
