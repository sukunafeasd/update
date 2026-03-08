param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [string]$RenderApiKey = $env:RENDER_API_KEY,
  [string]$RenderServiceId = $env:RENDER_SERVICE_ID,
  [string]$OwnerPassword = $env:PAINEL_DIEF_OWNER_PASSWORD,
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

  throw "Producao nao ficou pronta dentro do tempo limite."
}

function Wait-RenderDeployLive {
  param(
    [string]$ServiceId,
    [string]$DeployId,
    [hashtable]$Headers,
    [int]$TimeoutSec
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSec)
  do {
    $deploy = Invoke-RestMethod -Method Get -Uri ("https://api.render.com/v1/services/" + $ServiceId + "/deploys/" + $DeployId) -Headers $Headers
    $status = [string]$deploy.status
    if ($status -eq "live") {
      return
    }
    if ($status -in @("build_failed", "update_failed", "canceled")) {
      throw ("Deploy do Render falhou com status: " + $status)
    }
    Start-Sleep -Seconds 5
  } while ((Get-Date) -lt $deadline)

  throw "Deploy do Render nao ficou live dentro do tempo limite."
}

Write-Host "[1/6] backup remoto"
$backupPath = & powershell -ExecutionPolicy Bypass -File $backupScript -BaseUrl $BaseUrl -OpsToken $OpsToken
if (-not (Test-Path $backupPath)) {
  throw "Backup remoto nao foi gerado."
}

Write-Host "[2/6] trigger deploy no Render"
$headers = @{
  Authorization = "Bearer $RenderApiKey"
  Accept = "application/json"
  "Content-Type" = "application/json"
}
$deployBody = @{ clearCache = "do_not_clear" } | ConvertTo-Json
$deploy = Invoke-RestMethod -Method Post -Uri ("https://api.render.com/v1/services/" + $RenderServiceId + "/deploys") -Headers $headers -Body $deployBody
if ([string]::IsNullOrWhiteSpace([string]$deploy.id)) {
  throw "Render nao devolveu id do deploy."
}

Write-Host "[3/6] aguardando deploy ficar live"
Wait-RenderDeployLive -ServiceId $RenderServiceId -DeployId ([string]$deploy.id) -Headers $headers -TimeoutSec $WaitTimeoutSec

Write-Host "[4/6] aguardando producao responder"
Wait-ProductionReady -HealthUrl ($BaseUrl.TrimEnd("/") + "/api/health") -ReadyUrl ($BaseUrl.TrimEnd("/") + "/api/ready") -TimeoutSec $WaitTimeoutSec

Write-Host "[5/6] restaurando snapshot apos deploy"
$opsHeaders = @{
  Authorization = "Bearer $OpsToken"
}
Invoke-RestMethod -Method Post -Uri ($BaseUrl.TrimEnd("/") + "/api/ops/import") -Headers $opsHeaders -InFile $backupPath -ContentType "application/zip" | Out-Null

Write-Host "[6/6] smoke final"
if (-not [string]::IsNullOrWhiteSpace($OwnerPassword)) {
  $env:PAINEL_DIEF_OWNER_PASSWORD = $OwnerPassword
  & powershell -ExecutionPolicy Bypass -File $smokeScript -BaseUrl $BaseUrl -Password $OwnerPassword -OpsToken $OpsToken -MutatingChecks
} else {
  & powershell -ExecutionPolicy Bypass -File $smokeScript -BaseUrl $BaseUrl -OpsToken $OpsToken -MutatingChecks
}
