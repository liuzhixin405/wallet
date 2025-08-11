@echo off
setlocal

set "ROOT=%~dp0"

rem Backend
start "Wallet Backend" cmd /k "cd /d %ROOT%wallet-backend && set WALLET_JWT_SECRET=your-secret-key-here-change-in-production && go run cmd/main.go"

rem Frontend
start "Wallet Frontend" cmd /k "cd /d %ROOT%wallet-frontend && npm start"

endlocal 