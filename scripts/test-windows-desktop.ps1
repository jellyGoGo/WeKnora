param(
    [Parameter(Mandatory = $true)]
    [string]$InstallerPath,
    [int]$TimeoutSeconds = 90
)

$ErrorActionPreference = 'Stop'
$smokeRoot = Join-Path ([IO.Path]::GetTempPath()) ('weknora-desktop-smoke-' + [guid]::NewGuid().ToString('N'))
$installDir = Join-Path $smokeRoot 'WeKnora Lite'
$appData = Join-Path $smokeRoot 'AppData'
$prefsDir = Join-Path $appData 'WeKnora Lite'
$appProcess = $null
New-Item -ItemType Directory -Path $prefsDir -Force | Out-Null

try {
    $installer = Start-Process -FilePath (Resolve-Path -LiteralPath $InstallerPath).Path -ArgumentList "/S /D=$installDir" -WindowStyle Hidden -PassThru
    if (-not $installer.WaitForExit(60000)) { throw 'Desktop installer did not finish within 60 seconds.' }
    if ($installer.ExitCode -ne 0) { throw "Desktop installer failed: $($installer.ExitCode)" }

    foreach ($path in @(
        'WeKnora Lite.exe', '.env', 'config/config.yaml', 'web/index.html',
        'dict/jieba.dict.utf8', 'dict/hmm_model.utf8', 'dict/user.dict.utf8',
        'dict/idf.utf8', 'dict/stop_words.utf8', 'migrations/sqlite'
    )) {
        if (-not (Test-Path -LiteralPath (Join-Path $installDir $path))) {
            throw "Installed desktop resource is missing: $path"
        }
    }

    $portProbe = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, 0)
    $portProbe.Start()
    $port = $portProbe.LocalEndpoint.Port
    $portProbe.Stop()
    @{ http_port = $port } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $prefsDir 'desktop-prefs.json') -Encoding utf8NoBOM

    # Isolate startup from the runner/user's data, credentials and dictionary cache.
    $env:APPDATA = $appData
    $env:LOCALAPPDATA = Join-Path $smokeRoot 'LocalAppData'
    $env:XDG_CONFIG_HOME = $appData
    foreach ($name in @('DB_DRIVER', 'DB_PATH', 'JIEBA_DICT_DIR', 'WEKNORA_WEB_DIR')) {
        Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
    }
    $env:LOG_PATH = Join-Path $smokeRoot 'startup.log'
    $appProcess = Start-Process -FilePath (Join-Path $installDir 'WeKnora Lite.exe') -WorkingDirectory $smokeRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot 'stdout.log') -RedirectStandardError (Join-Path $smokeRoot 'stderr.log') -PassThru
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    $ready = $false
    while ([DateTime]::UtcNow -lt $deadline) {
        $appProcess.Refresh()
        if ($appProcess.HasExited) { throw "Installed desktop exited during startup: $($appProcess.ExitCode)" }
        try {
            $health = Invoke-WebRequest -Uri "http://127.0.0.1:$port/health" -TimeoutSec 2
            if ($health.StatusCode -eq 200) { $ready = $true; break }
        } catch { }
        Start-Sleep -Milliseconds 500
    }
    if (-not $ready) { throw 'Installed desktop did not become healthy.' }

    $page = Invoke-WebRequest -Uri "http://127.0.0.1:$port/" -TimeoutSec 10
    if ($page.StatusCode -ne 200 -or $page.Content -notmatch '(?i)<!doctype html') {
        throw 'Installed desktop did not serve the frontend HTML.'
    }
    $asset = [regex]::Match($page.Content, '<script[^>]+src="([^"]+)"')
    if (-not $asset.Success) { throw 'Desktop HTML has no frontend script.' }
    $assetUrl = [Uri]::new([Uri]"http://127.0.0.1:$port/", $asset.Groups[1].Value)
    $script = Invoke-WebRequest -Uri $assetUrl -TimeoutSec 10
    if ($script.StatusCode -ne 200 -or $script.RawContentLength -eq 0) { throw 'Desktop frontend script is unavailable.' }
    if (-not (Test-Path -LiteralPath (Join-Path $prefsDir 'data/weknora.db'))) { throw 'Desktop database was not created in the user data directory.' }
    $appProcess.Refresh()
    if ($appProcess.HasExited) { throw 'Desktop exited after becoming healthy.' }
    Write-Host 'PASS: installed desktop stays running, serves its frontend and stores its database outside the install directory.'
} catch {
    foreach ($name in @('stderr.log', 'stdout.log', 'startup.log')) {
        $path = Join-Path $smokeRoot $name
        if (Test-Path -LiteralPath $path) { Get-Content -LiteralPath $path -Tail 40 | Write-Host }
    }
    throw
} finally {
    if ($appProcess -and -not $appProcess.HasExited) { Stop-Process -Id $appProcess.Id -Force }
    Write-Host "Desktop smoke-test files: $smokeRoot"
}
