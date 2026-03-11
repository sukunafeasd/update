param()

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$runtimeDir = Join-Path $root ".runtime"
$pidFile = Join-Path $runtimeDir "backup-watch-process.json"
$stateFile = Join-Path $runtimeDir "backup-watch-state.json"

$status = [ordered]@{
  running = $false
  pid = $null
  startedAt = $null
  baseUrl = $null
  intervalSec = $null
  lastFingerprint = $null
  lastBackupAt = $null
  lastBackupPath = $null
  lastUsers = $null
  lastMessages = $null
  lastUploads = $null
  lastSeenAt = $null
  lastErrorAt = $null
  lastError = $null
}

if (Test-Path $pidFile) {
  try {
    $pidData = Get-Content $pidFile -Raw | ConvertFrom-Json
    $status.pid = $pidData.pid
    $status.startedAt = $pidData.startedAt
    $status.baseUrl = $pidData.baseUrl
    $status.intervalSec = $pidData.intervalSec
    if ($pidData.pid) {
      $proc = Get-Process -Id ([int]$pidData.pid) -ErrorAction SilentlyContinue
      $status.running = $null -ne $proc
    }
  } catch {
  }
}

if (Test-Path $stateFile) {
  try {
    $stateData = Get-Content $stateFile -Raw | ConvertFrom-Json
    $status.lastFingerprint = $stateData.lastFingerprint
    $status.lastBackupAt = $stateData.lastBackupAt
    $status.lastBackupPath = $stateData.lastBackupPath
    $status.lastUsers = $stateData.lastUsers
    $status.lastMessages = $stateData.lastMessages
    $status.lastUploads = $stateData.lastUploads
    $status.lastSeenAt = $stateData.lastSeenAt
    $status.lastErrorAt = $stateData.lastErrorAt
    $status.lastError = $stateData.lastError
  } catch {
  }
}

$status | ConvertTo-Json -Depth 5
