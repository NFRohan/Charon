param(
  [ValidateSet("up", "down", "reset", "status")]
  [string]$Command = "up"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "deploy/docker-compose.yml"

Push-Location $repoRoot
try {
  docker compose -f $composeFile run --rm --build migrate $Command
  if ($LASTEXITCODE -ne 0) {
    throw "Migration command failed."
  }
}
finally {
  Pop-Location
}
