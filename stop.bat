@echo off
chcp 65001 >nul 2>&1
setlocal EnableExtensions EnableDelayedExpansion

set "SILENT=0"
if /i "%~1"=="silent" set "SILENT=1"

set "KILLED=0"

call :killPort 30302
call :killPort 30301

if "%SILENT%"=="0" (
  if "!KILLED!"=="1" (
    echo [stop] ports 30302 / 30301 released
  ) else (
    echo [stop] ports 30302 / 30301 not in use
  )
  pause
)
exit /b 0

:killPort
set "TARGET_PORT=%~1"
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":%TARGET_PORT% " ^| findstr LISTENING') do (
  if not "%%p"=="0" (
    taskkill /F /PID %%p >nul 2>&1
    if not errorlevel 1 (
      set "KILLED=1"
      if "%SILENT%"=="0" echo [stop] killed PID %%p port %TARGET_PORT%
    )
  )
)
exit /b 0
