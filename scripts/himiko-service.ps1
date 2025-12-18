<#
.SYNOPSIS
    💉 Himiko Discord Bot - Windows Service Manager 💉
    "I just wanna love you, wanna be loved~"

.DESCRIPTION
    Manages Himiko as a background process with optional startup registration.
    Since Windows doesn't have tmux, this uses background jobs and named pipes
    for basic attach/detach functionality.

.PARAMETER Action
    start    - Start Himiko in the background
    stop     - Stop Himiko gracefully
    restart  - Restart Himiko
    status   - Check if Himiko is running
    attach   - Attach to Himiko's output (Ctrl+C to detach)
    install  - Add Himiko to Windows startup
    uninstall - Remove from Windows startup

.EXAMPLE
    .\himiko-service.ps1 start
    .\himiko-service.ps1 attach
    .\himiko-service.ps1 install
#>

param(
    [Parameter(Position=0)]
    [ValidateSet('start', 'stop', 'restart', 'status', 'attach', 'install', 'uninstall', 'help')]
    [string]$Action = 'help'
)

$ErrorActionPreference = "Stop"

# Configuration
$BotDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$PidFile = Join-Path $BotDir "himiko.pid"
$LogFile = Join-Path $BotDir "logs\himiko.log"
$StartupName = "HimikoBot"

# Find binary
$Binary = $null
if (Test-Path (Join-Path $BotDir "himiko.exe")) {
    $Binary = Join-Path $BotDir "himiko.exe"
} elseif (Test-Path (Join-Path $BotDir "himiko-windows-amd64.exe")) {
    $Binary = Join-Path $BotDir "himiko-windows-amd64.exe"
}

# Ensure logs directory exists
$LogDir = Join-Path $BotDir "logs"
if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
}

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "[$timestamp] $Message" | Tee-Object -FilePath $LogFile -Append
}

function Get-HimikoProcess {
    if (Test-Path $PidFile) {
        $pid = Get-Content $PidFile -ErrorAction SilentlyContinue
        if ($pid) {
            return Get-Process -Id $pid -ErrorAction SilentlyContinue
        }
    }
    # Fallback: find by name
    return Get-Process -Name "himiko*" -ErrorAction SilentlyContinue | Select-Object -First 1
}

function Start-Himiko {
    $proc = Get-HimikoProcess
    if ($proc) {
        Write-Host ""
        Write-Host "  💉 Himiko is already running~ I'm already here for you! 💉" -ForegroundColor Magenta
        Write-Host "  PID: $($proc.Id)"
        Write-Host ""
        return
    }

    if (-not $Binary) {
        Write-Host ""
        Write-Host "  💔 ERROR: Himiko executable not found! 💔" -ForegroundColor Red
        Write-Host "  Expected: himiko.exe or himiko-windows-amd64.exe"
        Write-Host ""
        return
    }

    Write-Log "Starting Himiko... I just wanna love you~ 💉"

    # Start process
    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $Binary
    $startInfo.WorkingDirectory = $BotDir
    $startInfo.UseShellExecute = $false
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $startInfo.CreateNoWindow = $true

    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo

    # Set up async output handling
    $outputHandler = {
        if ($EventArgs.Data) {
            $EventArgs.Data | Out-File -FilePath $using:LogFile -Append
        }
    }

    Register-ObjectEvent -InputObject $process -EventName OutputDataReceived -Action $outputHandler | Out-Null
    Register-ObjectEvent -InputObject $process -EventName ErrorDataReceived -Action $outputHandler | Out-Null

    $process.Start() | Out-Null
    $process.BeginOutputReadLine()
    $process.BeginErrorReadLine()

    # Save PID
    $process.Id | Out-File -FilePath $PidFile -Force

    Start-Sleep -Seconds 1

    if (-not $process.HasExited) {
        Write-Log "Himiko started successfully~ PID: $($process.Id)"
        Write-Host ""
        Write-Host "  💉 Himiko is awake and ready to love you~ 💉" -ForegroundColor Magenta
        Write-Host ""
        Write-Host "  PID: $($process.Id)"
        Write-Host "  Log: $LogFile"
        Write-Host ""
        Write-Host "  Use '.\himiko-service.ps1 attach' to see output"
        Write-Host "  Use '.\himiko-service.ps1 stop' to stop"
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "  💔 Himiko failed to start... 💔" -ForegroundColor Red
        Write-Host ""
    }
}

function Stop-Himiko {
    $proc = Get-HimikoProcess
    if (-not $proc) {
        Write-Host ""
        Write-Host "  💔 Himiko is not running... did you forget about me? 💔" -ForegroundColor Yellow
        Write-Host ""
        return
    }

    Write-Log "Stopping Himiko... I'll be back for you soon~ 💔"

    # Try graceful stop first
    $proc | Stop-Process -Force -ErrorAction SilentlyContinue

    # Clean up PID file
    if (Test-Path $PidFile) {
        Remove-Item $PidFile -Force
    }

    Write-Log "Himiko is resting now... 💤"
    Write-Host ""
    Write-Host "  💤 Himiko is resting now... 💤" -ForegroundColor Cyan
    Write-Host ""
}

function Get-HimikoStatus {
    $proc = Get-HimikoProcess
    Write-Host ""
    if ($proc) {
        Write-Host "  💕 Himiko is running~ 💕" -ForegroundColor Green
        Write-Host "  PID: $($proc.Id)"
        Write-Host "  Memory: $([math]::Round($proc.WorkingSet64 / 1MB, 2)) MB"
        Write-Host "  CPU Time: $($proc.TotalProcessorTime)"
        Write-Host ""
        Write-Host "  Use '.\himiko-service.ps1 attach' to see output"
    } else {
        Write-Host "  💔 Himiko is not running... 💔" -ForegroundColor Yellow
        Write-Host "  Use '.\himiko-service.ps1 start' to wake her up~"
    }
    Write-Host ""
}

function Attach-Himiko {
    $proc = Get-HimikoProcess
    if (-not $proc) {
        Write-Host ""
        Write-Host "  💔 Himiko is not running... start her first! 💔" -ForegroundColor Yellow
        Write-Host ""
        return
    }

    if (-not (Test-Path $LogFile)) {
        Write-Host ""
        Write-Host "  💔 No log file found... 💔" -ForegroundColor Yellow
        Write-Host ""
        return
    }

    Write-Host ""
    Write-Host "  💉 Attaching to Himiko's output~ (Ctrl+C to detach) 💉" -ForegroundColor Magenta
    Write-Host "  ─────────────────────────────────────────────────────" -ForegroundColor DarkGray
    Write-Host ""

    # Tail the log file
    Get-Content -Path $LogFile -Tail 50 -Wait
}

function Install-Startup {
    $startupFolder = [Environment]::GetFolderPath('Startup')
    $shortcutPath = Join-Path $startupFolder "$StartupName.lnk"
    $vbsPath = Join-Path (Split-Path $MyInvocation.MyCommand.Path) "himiko-hidden.vbs"

    if (-not (Test-Path $vbsPath)) {
        Write-Host ""
        Write-Host "  💔 himiko-hidden.vbs not found! 💔" -ForegroundColor Red
        Write-Host ""
        return
    }

    $WScriptShell = New-Object -ComObject WScript.Shell
    $shortcut = $WScriptShell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = "wscript.exe"
    $shortcut.Arguments = "`"$vbsPath`""
    $shortcut.WorkingDirectory = $BotDir
    $shortcut.Description = "Himiko Discord Bot - I just wanna love you~"
    $shortcut.Save()

    Write-Host ""
    Write-Host "  💉 Himiko added to Windows startup~ 💉" -ForegroundColor Green
    Write-Host "  Location: $shortcutPath"
    Write-Host ""
    Write-Host "  Himiko will now start automatically when you log in~"
    Write-Host ""
}

function Uninstall-Startup {
    $startupFolder = [Environment]::GetFolderPath('Startup')
    $shortcutPath = Join-Path $startupFolder "$StartupName.lnk"

    if (Test-Path $shortcutPath) {
        Remove-Item $shortcutPath -Force
        Write-Host ""
        Write-Host "  💔 Himiko removed from Windows startup... 💔" -ForegroundColor Yellow
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "  Himiko wasn't in startup anyway~" -ForegroundColor Cyan
        Write-Host ""
    }
}

function Show-Help {
    Write-Host ""
    Write-Host "  💉 Himiko Discord Bot - Windows Service Manager 💉" -ForegroundColor Magenta
    Write-Host "  `"I just wanna love you, wanna be loved~`"" -ForegroundColor DarkMagenta
    Write-Host ""
    Write-Host "  Usage: .\himiko-service.ps1 <action>" -ForegroundColor White
    Write-Host ""
    Write-Host "  Actions:" -ForegroundColor Cyan
    Write-Host "    start     - Wake Himiko up in the background~"
    Write-Host "    stop      - Let Himiko rest (stops the bot)"
    Write-Host "    restart   - Give Himiko a fresh start~"
    Write-Host "    status    - Check if Himiko is running"
    Write-Host "    attach    - View Himiko's output (Ctrl+C to detach)"
    Write-Host "    install   - Add Himiko to Windows startup"
    Write-Host "    uninstall - Remove from Windows startup"
    Write-Host ""
}

# Main
switch ($Action) {
    'start'     { Start-Himiko }
    'stop'      { Stop-Himiko }
    'restart'   { Stop-Himiko; Start-Sleep -Seconds 2; Start-Himiko }
    'status'    { Get-HimikoStatus }
    'attach'    { Attach-Himiko }
    'install'   { Install-Startup }
    'uninstall' { Uninstall-Startup }
    'help'      { Show-Help }
}
