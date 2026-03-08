$ErrorActionPreference = "SilentlyContinue"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$runtimeDir = Join-Path $root ".runtime"
$pidFile = Join-Path $runtimeDir "pids.json"

if (Test-Path $pidFile) {
  $meta = Get-Content $pidFile | ConvertFrom-Json
  if ($meta.serverPid) {
    Stop-Process -Id ([int]$meta.serverPid) -Force
  }
  if ($meta.tunnelPid) {
    Stop-Process -Id ([int]$meta.tunnelPid) -Force
  }
  Remove-Item $pidFile -Force
}

$portOwner = Get-NetTCPConnection -LocalPort 7788 -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($portOwner) {
  Stop-Process -Id $portOwner -Force
}
