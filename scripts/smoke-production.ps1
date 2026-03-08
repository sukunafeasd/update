param(
  [string]$BaseUrl = "https://update-wrl6.onrender.com",
  [string]$OpsToken = $env:UNIVERSALD_OPS_TOKEN,
  [switch]$MutatingChecks
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$smokeScript = Join-Path $root "scripts\\smoke-panel.ps1"

$args = @(
  "-ExecutionPolicy", "Bypass",
  "-File", $smokeScript,
  "-BaseUrl", $BaseUrl
)
if (-not [string]::IsNullOrWhiteSpace($OpsToken)) {
  $args += @("-OpsToken", $OpsToken)
}
if ($MutatingChecks) {
  $args += "-MutatingChecks"
}

powershell @args
