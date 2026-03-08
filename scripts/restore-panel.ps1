param(
  [Parameter(Mandatory = $true)]
  [string]$ArchivePath,
  [string]$DataDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) "."),
  [switch]$Force
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $ArchivePath)) {
  throw "Arquivo de backup nao encontrado: $ArchivePath"
}

$resolvedArchive = Resolve-Path $ArchivePath
$resolvedDataDir = Convert-Path (New-Item -ItemType Directory -Path $DataDir -Force)
$dbPath = Join-Path $resolvedDataDir "universald.db"
$uploadsPath = Join-Path $resolvedDataDir "panel_uploads"
$extractDir = Join-Path $env:TEMP ("painel-dief-restore-" + (Get-Date -Format "yyyyMMdd-HHmmss"))

New-Item -ItemType Directory -Path $extractDir -Force | Out-Null

try {
  Expand-Archive -Path $resolvedArchive -DestinationPath $extractDir -Force

  $snapshotPath = Join-Path $extractDir "universald.snapshot.db"
  if (-not (Test-Path $snapshotPath)) {
    throw "Backup invalido: snapshot do banco nao encontrado."
  }

  if ((Test-Path $dbPath) -and (-not $Force)) {
    throw "Ja existe um banco em $dbPath. Use -Force para sobrescrever."
  }

  if (Test-Path $dbPath) {
    Remove-Item $dbPath -Force -ErrorAction SilentlyContinue
  }
  Remove-Item ($dbPath + "-wal") -Force -ErrorAction SilentlyContinue
  Remove-Item ($dbPath + "-shm") -Force -ErrorAction SilentlyContinue
  Copy-Item $snapshotPath $dbPath -Force

  $snapshotUploads = Join-Path $extractDir "panel_uploads"
  if (Test-Path $snapshotUploads) {
    if (Test-Path $uploadsPath) {
      if (-not $Force) {
        throw "Ja existe diretório de uploads em $uploadsPath. Use -Force para sobrescrever."
      }
      Remove-Item $uploadsPath -Recurse -Force
    }
    Copy-Item $snapshotUploads $uploadsPath -Recurse -Force
  }

  Write-Output "restore-ok"
  Write-Output $dbPath
  if (Test-Path $uploadsPath) {
    Write-Output $uploadsPath
  }
} finally {
  if (Test-Path $extractDir) {
    Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}
