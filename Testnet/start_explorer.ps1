$ErrorActionPreference = "Stop"
Write-Host "Building POATCscan explorer server..." -ForegroundColor Cyan
Push-Location "$PSScriptRoot\explorer"
go build -o explorer.exe .\serve.go | Out-Null
Write-Host "Starting explorer at http://localhost:8080" -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit -Command & { cd '$PWD'; .\explorer.exe }" | Out-Null
Pop-Location

