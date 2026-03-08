$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $repoRoot "deploy/docker-compose.yml"

function Invoke-Checked {
  param(
    [scriptblock]$Command,
    [string]$ErrorMessage
  )

  & $Command
  if ($LASTEXITCODE -ne 0) {
    throw $ErrorMessage
  }
}

function Wait-ForPostgres {
  param(
    [string]$ComposeFile
  )

  for ($attempt = 1; $attempt -le 30; $attempt++) {
    & docker compose -f $ComposeFile exec -T postgres pg_isready -U charon -d charon | Out-Null
    if ($LASTEXITCODE -eq 0) {
      return
    }

    Start-Sleep -Seconds 2
  }

  throw "Postgres did not become ready in time."
}

Push-Location $repoRoot
try {
  Invoke-Checked { docker compose -f $composeFile up -d postgres redis rabbitmq } "Failed to start infrastructure containers."
  Wait-ForPostgres -ComposeFile $composeFile

  Invoke-Checked { docker compose -f $composeFile run --rm migrate up } "Failed to apply database migrations."

  Invoke-Checked { docker compose -f $composeFile up -d api worker } "Failed to start API and worker containers."

  Write-Host "Charon local services are up."
  Write-Host "API health: http://localhost:8080/healthz"
  Write-Host "RabbitMQ UI: http://localhost:15672"
  Write-Host "To start the admin app manually: cd apps/admin_app; npm run dev"
}
finally {
  Pop-Location
}
