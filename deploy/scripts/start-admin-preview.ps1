param(
  [string]$RuntimeDirectory = (Join-Path $env:USERPROFILE 'Documents/HackwaysLocalPreview'),
  [string]$PostgresBin = 'C:/Program Files/PostgreSQL/18/bin'
)

$ErrorActionPreference = 'Stop'
$previewRoot = [System.IO.Path]::GetFullPath($RuntimeDirectory)
$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$sourcePath = Join-Path $previewRoot 'preview-runtime.json'
if (-not (Test-Path -LiteralPath $sourcePath)) { throw 'Protected local preview runtime is missing.' }
$settings = Get-Content -LiteralPath $sourcePath -Raw | ConvertFrom-Json
$databaseAddress = $null
if (-not [System.Uri]::TryCreate($settings.database_url, [System.UriKind]::Absolute, [ref]$databaseAddress)) {
  throw 'Invalid local preview configuration.'
}
if ($databaseAddress.Host -ne '127.0.0.1' -or $databaseAddress.Port -ne 55440 -or $databaseAddress.AbsolutePath -ne '/hackways_admin_preview') {
  throw 'Refusing any target other than the isolated local preview database.'
}

if (-not (Get-NetTCPConnection -LocalPort 55440 -State Listen -ErrorAction SilentlyContinue)) {
  $dataDirectory = Join-Path $previewRoot 'pgdata'
  if (-not (Test-Path -LiteralPath (Join-Path $dataDirectory 'PG_VERSION'))) { throw 'Restored local preview cluster is missing.' }
  & (Join-Path $PostgresBin 'pg_ctl.exe') -D $dataDirectory -l (Join-Path $previewRoot 'postgres.log') -o '-h 127.0.0.1 -p 55440 -c max_connections=24 -c shared_buffers=64MB' -w start
  if ($LASTEXITCODE -ne 0) { throw 'Local preview PostgreSQL could not start.' }
}

$environmentNames = @(
  'APP_ENV', 'ADMIN_PREVIEW_ONLY', 'DATABASE_URL', 'BIND_HOST', 'DB_MAX_CONNS', 'DB_MIN_CONNS',
  'CORS_ALLOWED_ORIGINS', 'PORT', 'SESSION_SECRET', 'GOOGLE_CLIENT_ID', 'GOOGLE_CLIENT_SECRET',
  'GOOGLE_REDIRECT_URL', 'GOOGLE_ALLOWED_RETURN_ORIGINS', 'ACS_ENDPOINT', 'ACS_ACCESS_KEY',
  'ACS_SENDER_ADDRESS', 'AZURE_ACS_CONNECTION_STRING', 'AZURE_ACS_SENDER_ADDRESS',
  'AZURE_STORAGE_ACCOUNT', 'AZURE_STORAGE_ACCOUNT_KEY', 'AZURE_CDN_DOMAIN', 'ADMIN_PREVIEW_SNAPSHOT_AT'
)
$originalEnvironment = @{}
foreach ($name in $environmentNames) {
  $originalEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
  [Environment]::SetEnvironmentVariable($name, $null, 'Process')
}
try {
  $env:APP_ENV = 'development'
  $env:ADMIN_PREVIEW_ONLY = 'true'
  $env:DATABASE_URL = $settings.database_url
  $env:BIND_HOST = '127.0.0.1'
  $env:DB_MAX_CONNS = '4'
  $env:DB_MIN_CONNS = '0'
  $env:CORS_ALLOWED_ORIGINS = 'http://localhost:3003,http://127.0.0.1:3003'
  $env:ADMIN_PREVIEW_SNAPSHOT_AT = $settings.snapshot_at
  foreach ($engine in @(@{ name = 'core-backend'; port = 8081 }, @{ name = 'global-backend'; port = 8082 })) {
    $executable = Join-Path $previewRoot ($engine.name + '.exe')
    if (-not (Test-Path -LiteralPath $executable)) { throw 'Build the reviewed local Go backends before starting preview.' }
    $listener = Get-NetTCPConnection -LocalPort $engine.port -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($listener) {
      $running = Get-CimInstance Win32_Process -Filter "ProcessId=$($listener.OwningProcess)"
      if ($running.ExecutablePath -ne $executable) { throw 'A preview port is occupied by another application; it was not stopped.' }
      continue
    }
    $env:PORT = [string]$engine.port
    Start-Process -FilePath $executable -WindowStyle Hidden -WorkingDirectory $previewRoot -RedirectStandardOutput (Join-Path $previewRoot ($engine.name + '.log')) -RedirectStandardError (Join-Path $previewRoot ($engine.name + '.error.log')) | Out-Null
  }
  if (-not (Get-NetTCPConnection -LocalPort 3003 -State Listen -ErrorAction SilentlyContinue)) {
    Start-Process -FilePath 'C:/Windows/System32/cmd.exe' -ArgumentList '/c npm run dev --workspace @hackways/admin-web' -WorkingDirectory $repositoryRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $previewRoot 'admin-web.log') -RedirectStandardError (Join-Path $previewRoot 'admin-web.error.log') | Out-Null
  }
} finally {
  foreach ($name in $environmentNames) { [Environment]::SetEnvironmentVariable($name, $originalEnvironment[$name], 'Process') }
}
Write-Output 'Local admin preview is running or starting at http://localhost:3003/. Production is not contacted by this launcher.'
