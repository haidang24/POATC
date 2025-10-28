# Script Tuy Chinh Hoan Chinh Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$LogoPath = "",
    [string]$PrimaryColor = "#ff6b35",
    [string]$SecondaryColor = "#2563eb",
    [string]$BackgroundColor = "#ffffff",
    [string]$ThemeName = "POATC"
)

Write-Host "Tuy Chinh Hoan Chinh Blockscout..." -ForegroundColor Magenta

# 1. Disable ads
Write-Host "Buoc 1: Disable quang cao..." -ForegroundColor Yellow
& ".\disable_ads.ps1"

# 2. Thay logo
Write-Host "Buoc 2: Thay logo..." -ForegroundColor Yellow
if ($LogoPath) {
    & ".\replace_logo.ps1" -LogoPath $LogoPath
} else {
    & ".\replace_logo.ps1"
}

# 3. Tuy chinh theme
Write-Host "Buoc 3: Tuy chinh theme..." -ForegroundColor Yellow
& ".\customize_theme.ps1" -PrimaryColor $PrimaryColor -SecondaryColor $SecondaryColor -BackgroundColor $BackgroundColor -ThemeName $ThemeName

Write-Host ""
Write-Host "HOAN THANH TAT CA!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Disabled quang cao" -ForegroundColor Cyan
Write-Host "Thay logo POATC" -ForegroundColor Cyan
Write-Host "Tuy chinh theme $ThemeName" -ForegroundColor Cyan
Write-Host ""
Write-Host "Cac buoc cuoi cung:" -ForegroundColor Yellow
Write-Host "1. cd docker-compose" -ForegroundColor White
Write-Host "2. docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White