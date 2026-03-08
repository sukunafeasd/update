param(
  [string]$GoExe = "C:\Program Files\Go\bin\go.exe"
)

if (-not (Test-Path $GoExe)) {
  throw "Go not found at $GoExe"
}

& $GoExe mod tidy
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

& $GoExe run .\cmd\universald
exit $LASTEXITCODE
