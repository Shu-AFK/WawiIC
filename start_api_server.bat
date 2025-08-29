@echo off
setlocal

rem Use provided path if given, else default
if "%~1"=="" (
    set "exe=C:\Programm Files(x86)\JTL-Software\JTL.Wawi.Rest.Exe"
) else (
    set "exe=%~1"
)

rem Check existence
if not exist "%exe%" (
    echo Error: Executable not found at "%exe%"
    exit /b 1
)

rem Run with required arguments
"%exe%" -w "Standard" -l 127.0.0.1 --dev
