param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [string]$OutputDir = "",
  [string]$MirrorDir = "",
  [string]$StateFile = "",
  [string]$LogFile = "",
  [int]$IntervalSec = 30,
  [int]$RetentionDays = 14,
  [switch]$Once
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($OpsToken)) {
  throw "UNIVERSALD_OPS_TOKEN obrigatorio para observar backups da producao."
}
if ($IntervalSec -lt 10) {
  $IntervalSec = 10
}

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if ([string]::IsNullOrWhiteSpace($OutputDir)) {
  $OutputDir = Join-Path $root "backups\remote"
}
if ([string]::IsNullOrWhiteSpace($MirrorDir)) {
  $MirrorDir = Join-Path $root ".runtime\production-mirror"
}
if ([string]::IsNullOrWhiteSpace($StateFile)) {
  $StateFile = Join-Path $root ".runtime\backup-watch-state.json"
}
if ([string]::IsNullOrWhiteSpace($LogFile)) {
  $LogFile = Join-Path $root ".runtime\backup-watch.log"
}

$runtimeDir = Split-Path -Parent $StateFile
New-Item -ItemType Directory -Path $runtimeDir -Force | Out-Null
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
New-Item -ItemType Directory -Path $MirrorDir -Force | Out-Null

$syncScript = Join-Path $root "scripts\sync-production-mirror.ps1"
$pruneScript = Join-Path $root "scripts\prune-backups.ps1"

function Write-BackupWatchLog {
  param([string]$Message)
  $stamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
  $line = "[${stamp}] $Message"
  Add-Content -Path $LogFile -Value $line
  Write-Host $line
}

function Load-BackupWatchState {
  if (-not (Test-Path $StateFile)) {
    return @{}
  }
  try {
    return Get-Content $StateFile -Raw | ConvertFrom-Json -AsHashtable
  } catch {
    Write-BackupWatchLog "estado anterior invalido; recriando arquivo de estado"
    return @{}
  }
}

function Save-BackupWatchState {
  param([hashtable]$State)
  ($State | ConvertTo-Json -Depth 6) | Set-Content -Path $StateFile -Encoding UTF8
}

function Get-RemoteSummary {
  $headers = @{ Authorization = "Bearer $OpsToken" }
  return Invoke-RestMethod -Method Get -Uri ($BaseUrl.TrimEnd("/") + "/api/ops/summary") -Headers $headers -UseBasicParsing -TimeoutSec 30
}

function Sync-ProductionBackup {
  $raw = & powershell -ExecutionPolicy Bypass -File $syncScript -BaseUrl $BaseUrl -OpsToken $OpsToken -OutputDir $OutputDir -MirrorDir $MirrorDir -RetentionDays $RetentionDays
  $lines = @($raw | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($lines.Count -lt 2) {
    throw "sync-production-mirror nao retornou archive e mirror."
  }
  return @{
    ArchivePath = $lines[0]
    MirrorDir   = $lines[1]
  }
}

function Protect-BackupRetention {
  & powershell -ExecutionPolicy Bypass -File $pruneScript -Directory $OutputDir -Keep 8 | Out-Null
}

$state = Load-BackupWatchState
Write-BackupWatchLog "guardiao de backup iniciado para $BaseUrl"

while ($true) {
  try {
    $summary = Get-RemoteSummary
    $fingerprint = [string]$summary.summary.dataFingerprint
    $users = [int]$summary.summary.users
    $messages = [int]$summary.summary.messages
    $uploads = [int]$summary.summary.uploadFiles
    $now = (Get-Date).ToUniversalTime().ToString("o")

    $needsSync = $false
    if ([string]::IsNullOrWhiteSpace($fingerprint)) {
      throw "ops summary nao trouxe dataFingerprint"
    }
    if (-not $state.ContainsKey("lastFingerprint")) {
      $needsSync = $true
      Write-BackupWatchLog "primeira observacao detectada; criando espelho inicial"
    } elseif ([string]$state.lastFingerprint -ne $fingerprint) {
      $needsSync = $true
      Write-BackupWatchLog "mudanca detectada na assinatura dos dados; criando backup incremental"
    }

    if ($needsSync) {
      $result = Sync-ProductionBackup
      $state.lastFingerprint = $fingerprint
      $state.lastBackupPath = $result.ArchivePath
      $state.lastMirrorDir = $result.MirrorDir
      $state.lastBackupAt = $now
      $state.lastUsers = $users
      $state.lastMessages = $messages
      $state.lastUploads = $uploads
      Save-BackupWatchState $state
      Protect-BackupRetention
      Write-BackupWatchLog ("backup salvo: users={0} messages={1} uploads={2}" -f $users, $messages, $uploads)
    } else {
      $state.lastSeenAt = $now
      $state.lastUsers = $users
      $state.lastMessages = $messages
      $state.lastUploads = $uploads
      Save-BackupWatchState $state
    }
  } catch {
    $state.lastErrorAt = (Get-Date).ToUniversalTime().ToString("o")
    $state.lastError = $_.Exception.Message
    Save-BackupWatchState $state
    Write-BackupWatchLog ("falha no guardiao: " + $_.Exception.Message)
  }

  if ($Once) {
    break
  }
  Start-Sleep -Seconds $IntervalSec
}
