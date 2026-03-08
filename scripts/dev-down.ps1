$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "deploy/docker-compose.yml"

Push-Location $repoRoot
try {
  docker compose -f $composeFile down
}
finally {
  Pop-Location
}
