param(
  [string]$BaseUrl = "http://127.0.0.1:7788",
  [string]$Token = ""
)

$ErrorActionPreference = "Stop"

$headers = @{}
if ($Token) {
  $headers["Authorization"] = "Bearer $Token"
}

$payload = Invoke-RestMethod -Uri ($BaseUrl.TrimEnd("/") + "/api/ops/summary") -Headers $headers -UseBasicParsing
$payload | ConvertTo-Json -Depth 6
