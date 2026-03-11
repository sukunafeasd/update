param()

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$pidFile = Join-Path $root ".runtime\backup-watch-process.json"

if (-not (Test-Path $pidFile)) {
  Write-Output "guardiao nao estava rodando"
  exit 0
}

try {
  $data = Get-Content $pidFile -Raw | ConvertFrom-Json
  if ($data.pid) {
    Stop-Process -Id ([int]$data.pid) -Force -ErrorAction SilentlyContinue
  }
} finally {
  Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
}

Write-Output "guardiao parado"
