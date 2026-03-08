param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [string]$RenderApiKey = $env:RENDER_API_KEY,
  [string]$RenderServiceId = $env:RENDER_SERVICE_ID,
  [int]$WaitTimeoutSec = 900
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($OpsToken)) {
  throw "UNIVERSALD_OPS_TOKEN obrigatorio."
}
if ([string]::IsNullOrWhiteSpace($RenderApiKey)) {
  throw "RENDER_API_KEY obrigatorio."
}
if ([string]::IsNullOrWhiteSpace($RenderServiceId)) {
  throw "RENDER_SERVICE_ID obrigatorio."
}

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$backupScript = Join-Path $root "scripts\\backup-remote-panel.ps1"
$smokeScript = Join-Path $root "scripts\\smoke-production.ps1"

function Wait-ProductionReady {
  param(
    [string]$HealthUrl,
    [string]$ReadyUrl,
    [int]$TimeoutSec
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSec)
  do {
    try {
      $health = Invoke-RestMethod -Uri $HealthUrl -UseBasicParsing -TimeoutSec 30
      $ready = Invoke-RestMethod -Uri $ReadyUrl -UseBasicParsing -TimeoutSec 30
      if ($health.status -eq "ok" -and $ready.ready -eq $true) {
        return
      }
    } catch {
    }
    Start-Sleep -Seconds 5
  } while ((Get-Date) -lt $deadline)

  throw "Produção não ficou pronta dentro do tempo limite."
}

Write-Host "[1/5] backup remoto"
$backupPath = & powershell -ExecutionPolicy Bypass -File $backupScript -BaseUrl $BaseUrl -OpsToken $OpsToken
if (-not (Test-Path $backupPath)) {
  throw "Backup remoto não foi gerado."
}

Write-Host "[2/5] trigger deploy no Render"
$headers = @{
  Authorization = "Bearer $RenderApiKey"
  Accept = "application/json"
  "Content-Type" = "application/json"
}
$deployBody = @{ clearCache = "do_not_clear" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri ("https://api.render.com/v1/services/" + $RenderServiceId + "/deploys") -Headers $headers -Body $deployBody | Out-Null

Write-Host "[3/5] aguardando produção voltar"
Wait-ProductionReady -HealthUrl ($BaseUrl.TrimEnd("/") + "/api/health") -ReadyUrl ($BaseUrl.TrimEnd("/") + "/api/ready") -TimeoutSec $WaitTimeoutSec

Write-Host "[4/5] restaurando snapshot após deploy"
$opsHeaders = @{
  Authorization = "Bearer $OpsToken"
}
Invoke-RestMethod -Method Post -Uri ($BaseUrl.TrimEnd("/") + "/api/ops/import") -Headers $opsHeaders -InFile $backupPath -ContentType "application/zip" | Out-Null

Write-Host "[5/5] smoke final"
& powershell -ExecutionPolicy Bypass -File $smokeScript -BaseUrl $BaseUrl -OpsToken $OpsToken -MutatingChecks
