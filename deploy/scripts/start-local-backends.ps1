param([switch]$CoreOnly,[switch]$GlobalOnly)
$ErrorActionPreference = 'Stop'
$workspace = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$envFile = Join-Path $workspace '.env'
if (-not (Test-Path -LiteralPath $envFile)) { throw 'Configure the ignored workspace .env from .env.example first.' }
Get-Content -LiteralPath $envFile | ForEach-Object {
 if ($_ -match '^([A-Z_][A-Z_0-9]*)=(.*)$') { $environmentName=$Matches[1];$environmentValue=($Matches[2] -replace '\s+#.*$','').Trim().Trim('"').Trim("'");[Environment]::SetEnvironmentVariable($environmentName,$environmentValue,'Process') }
}
$logDirectory=Join-Path $workspace 'logs'
New-Item -ItemType Directory -Force -Path $logDirectory | Out-Null
$services=@()
if (-not $GlobalOnly) {$services+=@{Name='core-backend';Port=8081;Memory='100MiB'}}
if (-not $CoreOnly) {$services+=@{Name='global-backend';Port=8082;Memory='90MiB'}}
foreach($service in $services){
 if(Get-NetTCPConnection -LocalPort $service.Port -State Listen -ErrorAction SilentlyContinue){Write-Output "$($service.Name) port is already in use; existing process left unchanged.";continue}
 $env:PORT=$service.Port.ToString();$env:GOMEMLIMIT=$service.Memory;$env:GOGC='100'
 $process=Start-Process -FilePath 'go' -ArgumentList @('run','./cmd/server') -WorkingDirectory (Join-Path $workspace "services/$($service.Name)") -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $logDirectory "$($service.Name).stdout.log") -RedirectStandardError (Join-Path $logDirectory "$($service.Name).stderr.log")
 Write-Output "$($service.Name) started with launcher PID $($process.Id); inspect logs and /health for readiness."
}
