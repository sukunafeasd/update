param(
  [string]$DataDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) "."),
  [string]$OutputDir = (Join-Path (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)) "backups"),
  [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"

$resolvedDataDir = Resolve-Path $DataDir
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$zipPath = Join-Path $OutputDir ("painel-dief-backup-" + $stamp + ".zip")
$stagingDir = Join-Path $env:TEMP ("painel-dief-backup-stage-" + $stamp)
New-Item -ItemType Directory -Path $stagingDir -Force | Out-Null

try {
  $dbPath = Join-Path $resolvedDataDir "universald.db"
  $uploadsPath = Join-Path $resolvedDataDir "panel_uploads"
  $items = @()
  $manifestFiles = @()

  if (Test-Path $dbPath) {
    $snapshotPath = Join-Path $stagingDir "universald.snapshot.db"
    $helperPath = Join-Path $stagingDir "sqlite_backup.go"
@'
package main
import (
  "database/sql"
  "fmt"
  "os"
  "strings"
  _ "modernc.org/sqlite"
)
func main() {
  if len(os.Args) != 3 {
    panic("usage: sqlite_backup <source> <dest>")
  }
  src := strings.ReplaceAll(os.Args[1], "\\", "/")
  dst := strings.ReplaceAll(os.Args[2], "\\", "/")
  db, err := sql.Open("sqlite", "file:"+src)
  if err != nil { panic(err) }
  defer db.Close()
  escaped := strings.ReplaceAll(dst, "'", "''")
  if _, err := db.Exec("VACUUM INTO '" + escaped + "'"); err != nil {
    panic(err)
  }
  fmt.Println(dst)
}
'@ | Set-Content $helperPath -Encoding UTF8
    go run $helperPath $dbPath $snapshotPath | Out-Null
    Remove-Item $helperPath -Force -ErrorAction SilentlyContinue
    $items += $snapshotPath
    $manifestFiles += [ordered]@{
      name = "universald.snapshot.db"
      type = "sqlite-snapshot"
      bytes = (Get-Item $snapshotPath).Length
    }
  }

  if (Test-Path $uploadsPath) {
    $uploadsSnapshot = Join-Path $stagingDir "panel_uploads"
    Copy-Item $uploadsPath $uploadsSnapshot -Recurse -Force
    $items += $uploadsSnapshot
    $uploadsFiles = @(Get-ChildItem $uploadsSnapshot -Recurse -File -ErrorAction SilentlyContinue)
    $manifestFiles += [ordered]@{
      name = "panel_uploads"
      type = "directory"
      fileCount = $uploadsFiles.Count
      bytes = ($uploadsFiles | Measure-Object -Property Length -Sum).Sum
    }
  }

  if (-not $items.Count) {
    throw "Nenhum arquivo de dados do Painel Dief foi encontrado em $resolvedDataDir"
  }

  $manifest = [ordered]@{
    service = "painel-dief"
    version = "1.4.4"
    createdAt = (Get-Date).ToString("o")
    sourceDataDir = [string]$resolvedDataDir
    retentionDays = $RetentionDays
    appEnv = [string]$env:UNIVERSALD_APP_ENV
    files = $manifestFiles
  }
  $manifest | ConvertTo-Json -Depth 6 | Set-Content (Join-Path $stagingDir "backup-manifest.json") -Encoding UTF8

  Compress-Archive -Path (Join-Path $stagingDir "*") -DestinationPath $zipPath -CompressionLevel Optimal -Force
  if ($RetentionDays -gt 0) {
    $cutoff = (Get-Date).AddDays(-1 * $RetentionDays)
    Get-ChildItem $OutputDir -Filter "painel-dief-backup-*.zip" -File -ErrorAction SilentlyContinue |
      Where-Object { $_.LastWriteTime -lt $cutoff } |
      Remove-Item -Force -ErrorAction SilentlyContinue
  }
  Write-Output $zipPath
} finally {
  if (Test-Path $stagingDir) {
    cmd /c "rmdir /s /q $stagingDir" | Out-Null
  }
}
