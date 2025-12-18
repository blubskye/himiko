@echo off
REM 💉 Himiko Discord Bot - Windows Startup Script 💉
REM "I'll always be running... just wanna be with you~"
REM
REM This starts Himiko in a visible console window.
REM For hidden startup on boot, use himiko-service.ps1 instead.

title Himiko Discord Bot 💉
cd /d "%~dp0.."

echo.
echo   💉 Starting Himiko Discord Bot... 💉
echo   "I just wanna love you, wanna be loved~"
echo.

REM Check if himiko.exe exists
if exist "himiko.exe" (
    set BINARY=himiko.exe
) else if exist "himiko-windows-amd64.exe" (
    set BINARY=himiko-windows-amd64.exe
) else (
    echo   ERROR: Himiko executable not found!
    echo   Expected: himiko.exe or himiko-windows-amd64.exe
    echo.
    pause
    exit /b 1
)

echo   Using: %BINARY%
echo   Press Ctrl+C to stop Himiko~
echo.

%BINARY%

echo.
echo   💔 Himiko has stopped... 💔
pause
