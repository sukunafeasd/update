param(
  [string]$GoExe = "C:\Program Files\Go\bin\go.exe"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $GoExe)) {
  throw "Go binary not found: $GoExe"
}

& $GoExe mod tidy
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

& $GoExe fmt ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

& $GoExe test ./...
exit $LASTEXITCODE
