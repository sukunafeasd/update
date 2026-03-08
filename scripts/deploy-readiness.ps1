param(
  [string]$Root = (Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)),
  [string]$PublicOrigin = "",
  [string]$Domain = "",
  [string]$Provider = "render"
)

$ErrorActionPreference = "Stop"

function Resolve-ToolPath {
  param([string]$Name)

  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd.Source
  }

  switch ($Name.ToLowerInvariant()) {
    "git" {
      foreach ($candidate in @(
        "C:\Program Files\Git\cmd\git.exe",
        "C:\Program Files\Git\bin\git.exe",
        "C:\Program Files (x86)\Git\cmd\git.exe",
        "C:\Program Files (x86)\Git\bin\git.exe"
      )) {
        if (Test-Path $candidate) {
          return $candidate
        }
      }
    }
  }

  return $null
}

function Add-CheckResult {
  param(
    [System.Collections.Generic.List[object]]$List,
    [string]$Name,
    [bool]$Ok,
    [string]$Detail
  )
  $List.Add([pscustomobject]@{
    name = $Name
    ok = $Ok
    detail = $Detail
  }) | Out-Null
}

$results = [System.Collections.Generic.List[object]]::new()
$missing = [System.Collections.Generic.List[string]]::new()

$dockerfile = Join-Path $Root "Dockerfile"
$renderFile = Join-Path $Root "render.yaml"
$stagingFile = Join-Path $Root "render.staging.yaml"
$productionDoc = Join-Path $Root "docs\PRODUCTION.md"
$preflightOrigin = $PublicOrigin
$renderBlueprintText = ""
if (Test-Path $renderFile) {
  $renderBlueprintText = Get-Content $renderFile -Raw
}

Add-CheckResult -List $results -Name "projectRoot" -Ok (Test-Path $Root) -Detail $Root
Add-CheckResult -List $results -Name "dockerfile" -Ok (Test-Path $dockerfile) -Detail $dockerfile
Add-CheckResult -List $results -Name "renderBlueprint" -Ok (Test-Path $renderFile) -Detail $renderFile
Add-CheckResult -List $results -Name "stagingBlueprint" -Ok (Test-Path $stagingFile) -Detail $stagingFile
Add-CheckResult -List $results -Name "productionDoc" -Ok (Test-Path $productionDoc) -Detail $productionDoc

$gitPath = Resolve-ToolPath -Name "git"
$gitAvailable = -not [string]::IsNullOrWhiteSpace($gitPath)
$gitDetail = "git nao encontrado no PATH"
if ($gitAvailable) {
  $gitDetail = $gitPath
}
Add-CheckResult -List $results -Name "gitAvailable" -Ok $gitAvailable -Detail $gitDetail
if (-not $gitAvailable) {
  $missing.Add("Instalar Git ou disponibilizar git no PATH para versionamento e deploy via provedor.") | Out-Null
}

$gitDir = Join-Path $Root ".git"
$gitRepo = Test-Path $gitDir
$gitRepoDetail = "repositorio Git ainda nao inicializado"
if ($gitRepo) {
  $gitRepoDetail = $gitDir
}
Add-CheckResult -List $results -Name "gitRepository" -Ok $gitRepo -Detail $gitRepoDetail
if (-not $gitRepo) {
  $missing.Add("Inicializar um repositorio Git em $Root e enviar o codigo para GitHub/GitLab/Bitbucket.") | Out-Null
}

$gitRemoteOk = $false
$gitRemoteDetail = "remote Git ainda nao configurado"
if ($gitAvailable -and $gitRepo) {
  $remoteOutput = & $gitPath -C $Root remote -v 2>$null
  if ($remoteOutput) {
    $gitRemoteOk = $true
    $gitRemoteDetail = (($remoteOutput | Select-Object -Unique) -join "; ")
  }
}
Add-CheckResult -List $results -Name "gitRemote" -Ok $gitRemoteOk -Detail $gitRemoteDetail
if (-not $gitRemoteOk) {
  $missing.Add("Configurar um remote Git hospedado para o deploy continuo do provedor.") | Out-Null
}

$opsTokenSet = -not [string]::IsNullOrWhiteSpace($env:UNIVERSALD_OPS_TOKEN)
$opsTokenBlueprint = $false
if ($Provider -eq "render" -and $renderBlueprintText -match 'UNIVERSALD_OPS_TOKEN' -and $renderBlueprintText -match 'generateValue:\s*true') {
  $opsTokenBlueprint = $true
}
$opsTokenOk = $opsTokenSet -or $opsTokenBlueprint
$opsDetail = "token nao definido no ambiente atual"
if ($opsTokenSet) {
  $opsDetail = "token carregado por ambiente"
} elseif ($opsTokenBlueprint) {
  $opsDetail = "token sera gerado automaticamente pela blueprint do Render"
}
Add-CheckResult -List $results -Name "opsToken" -Ok $opsTokenOk -Detail $opsDetail
if (-not $opsTokenOk) {
  $missing.Add("Definir um UNIVERSALD_OPS_TOKEN forte no provedor de hospedagem.") | Out-Null
}

if ([string]::IsNullOrWhiteSpace($preflightOrigin) -and -not [string]::IsNullOrWhiteSpace($env:UNIVERSALD_PUBLIC_ORIGIN)) {
  $preflightOrigin = $env:UNIVERSALD_PUBLIC_ORIGIN
}

$originOk = $false
if (-not [string]::IsNullOrWhiteSpace($preflightOrigin)) {
  $originOk = $preflightOrigin -match '^https://'
}
$originDetail = "origem publica nao informada"
if ($preflightOrigin) {
  $originDetail = $preflightOrigin
}
Add-CheckResult -List $results -Name "publicOrigin" -Ok $originOk -Detail $originDetail
if (-not $originOk) {
  $missing.Add("Definir a URL final em UNIVERSALD_PUBLIC_ORIGIN usando HTTPS.") | Out-Null
}

$domainOk = -not [string]::IsNullOrWhiteSpace($Domain)
$domainDetail = "dominio/subdominio final ainda nao informado"
if ($domainOk) {
  $domainDetail = $Domain
}
Add-CheckResult -List $results -Name "domain" -Ok $domainOk -Detail $domainDetail
if (-not $domainOk) {
  $missing.Add("Escolher o dominio ou subdominio fixo do Painel Dief e ter acesso ao DNS.") | Out-Null
}

$providerOk = -not [string]::IsNullOrWhiteSpace($Provider)
Add-CheckResult -List $results -Name "provider" -Ok $providerOk -Detail $Provider
if (-not $providerOk) {
  $missing.Add("Escolher um provedor de hospedagem com disco persistente.") | Out-Null
}

$providerAccessOk = $true
$providerAccessDetail = "acesso ao provedor nao verificado"
if ($Provider -eq "render") {
  $providerAccessOk = -not [string]::IsNullOrWhiteSpace($env:RENDER_API_KEY)
  if ($providerAccessOk) {
    $providerAccessDetail = "RENDER_API_KEY carregado no ambiente"
  } else {
    $providerAccessDetail = "RENDER_API_KEY nao encontrado; deploy automatico daqui ainda bloqueado"
    $missing.Add("Fornecer acesso ao Render por login no navegador ou RENDER_API_KEY para concluir a URL fixa daqui.") | Out-Null
  }
}
Add-CheckResult -List $results -Name "providerAccess" -Ok $providerAccessOk -Detail $providerAccessDetail

$goPath = Resolve-ToolPath -Name "go"
$goAvailable = -not [string]::IsNullOrWhiteSpace($goPath)
$goDetail = "go nao encontrado no PATH"
if ($goAvailable) {
  $goDetail = $goPath
}
Add-CheckResult -List $results -Name "goAvailable" -Ok $goAvailable -Detail $goDetail
if (-not $goAvailable) {
  $missing.Add("Instalar Go para validar build/testes locais antes do deploy.") | Out-Null
}

$summary = [pscustomobject]@{
  service = "Painel Dief"
  provider = $Provider
  root = $Root
  readyForFixedUrl = ($missing.Count -eq 0)
  missing = @($missing)
  checks = @($results)
  generatedAt = (Get-Date).ToString("o")
}

$summary | ConvertTo-Json -Depth 6
