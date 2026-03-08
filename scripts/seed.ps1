param(
  [string]$SeedSet = ""
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "deploy/docker-compose.yml"

Push-Location $repoRoot
try {
  if ([string]::IsNullOrWhiteSpace($SeedSet)) {
    docker compose -f $composeFile run --rm seed
  }
  else {
    docker compose -f $composeFile run --rm seed $SeedSet
  }

  if ($LASTEXITCODE -ne 0) {
    throw "Seed command failed."
  }
}
finally {
  Pop-Location
}
