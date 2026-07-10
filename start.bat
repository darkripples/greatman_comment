@echo off
chcp 65001 >nul 2>&1
setlocal EnableExtensions

if not defined GOCACHE set "GOCACHE=G:\cache\go\build"
if not defined GOMODCACHE set "GOMODCACHE=G:\cache\go\mod"
if not defined GOPATH set "GOPATH=G:\cache\go\work"
if not defined GOPROXY set "GOPROXY=https://goproxy.cn,direct"
if not defined GOSUMDB set "GOSUMDB=sum.golang.google.cn"

cd /d "%~dp0"

echo ========================================
echo   renwen local dev
echo ========================================
echo.

where go >nul 2>&1
if errorlevel 1 (
  echo [ERROR] Go not found in PATH
  pause
  exit /b 1
)

where npm >nul 2>&1
if errorlevel 1 (
  echo [ERROR] npm not found in PATH
  pause
  exit /b 1
)

if not exist "web\.env.local" (
  if exist "web\.env.local.example" (
    echo [INFO] creating web\.env.local
    copy /y "web\.env.local.example" "web\.env.local" >nul
  )
)

if not exist "web\node_modules" (
  echo [INFO] npm install...
  pushd web
  call npm install
  if errorlevel 1 (
    echo [ERROR] npm install failed
    popd
    pause
    exit /b 1
  )
  popd
)

echo [INFO] free ports 30302 / 30301
call "%~dp0stop.bat" silent

set "SERVER_PORT=30302"
set "WEB_PORT=30301"
set "CORS_ORIGINS=http://localhost:30301,http://127.0.0.1:30301"
if not defined RENWEN_DATA_DIR set "RENWEN_DATA_DIR=%~dp0server\data"

echo.
echo [CONFIG]
echo   SERVER_PORT=%SERVER_PORT%
echo   WEB_PORT=%WEB_PORT%
echo   RENWEN_DATA_DIR=%RENWEN_DATA_DIR%
echo   LLM/Mock/cache: use Settings page in app
if defined DEEPSEEK_API_KEY (echo   DEEPSEEK_API_KEY=set) else (echo   DEEPSEEK_API_KEY=missing)
if defined ZHIHU_API_KEY (echo   ZHIHU_API_KEY=set) else (echo   ZHIHU_API_KEY=missing)
echo.

echo [1/2] start Go server...
start "renwen-server" cmd /k "cd /d %~dp0server && title renwen-server && set SERVER_PORT=30302 && set CORS_ORIGINS=http://localhost:30301,http://127.0.0.1:30301 && set RENWEN_DATA_DIR=%RENWEN_DATA_DIR% && echo Backend http://127.0.0.1:30302 && go run ./cmd/server"

echo [WAIT] 2s...
timeout /t 2 /nobreak >nul

echo [2/2] start Next frontend...
start "renwen-web" cmd /k "cd /d %~dp0web && title renwen-web && set PORT=30301 && echo Frontend http://localhost:30301 && npm run dev"

echo.
echo ========================================
echo   renwen-server  http://127.0.0.1:30302
echo   renwen-web     http://localhost:30301
echo   run stop.bat to free ports
echo ========================================
echo.
pause
