# Script Thay Logo Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$LogoPath = "",
    [string]$LogoLightPath = "",
    [string]$FaviconPath = ""
)

Write-Host "Thay Logo Blockscout..." -ForegroundColor Cyan

$imagesPath = "apps\block_scout_web\assets\static\images"

# Kiem tra thu muc
if (-not (Test-Path $imagesPath)) {
    Write-Host "Khong tim thay thu muc $imagesPath" -ForegroundColor Red
    exit 1
}

Write-Host "Dang o thu muc: $imagesPath" -ForegroundColor Cyan

# Backup files goc
Write-Host "Backup logo cu..." -ForegroundColor Yellow
$logoFiles = @("logo.svg", "logo-light.svg", "favicon.ico", "apple-touch-icon.png")
foreach ($file in $logoFiles) {
    $filePath = Join-Path $imagesPath $file
    if (Test-Path $filePath) {
        $backupPath = "$filePath.backup"
        if (-not (Test-Path $backupPath)) {
            Copy-Item $filePath $backupPath
            Write-Host "Backup $file" -ForegroundColor Green
        }
    }
}

# Ham thay the logo
function Replace-Logo {
    param(
        [string]$SourcePath,
        [string]$TargetFileName,
        [string]$Description
    )
    
    if ($SourcePath -and (Test-Path $SourcePath)) {
        $targetPath = Join-Path $imagesPath $TargetFileName
        Copy-Item $SourcePath $targetPath -Force
        Write-Host "Thay the $Description" -ForegroundColor Green
    } else {
        Write-Host "Khong tim thay file cho $Description" -ForegroundColor Yellow
    }
}

# Thay the logo chinh
if ($LogoPath) {
    Replace-Logo -SourcePath $LogoPath -TargetFileName "logo.svg" -Description "logo chinh"
} else {
    Write-Host "Tao logo mac dinh cho POATC..." -ForegroundColor Yellow
    
    # Tao logo SVG mac dinh cho POATC
    $defaultLogo = @"
<svg width="200" height="50" viewBox="0 0 200 50" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="poatcGradient" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#ff6b35;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#2563eb;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background -->
  <rect width="200" height="50" fill="#1a1a1a" rx="8"/>
  
  <!-- POATC Text -->
  <text x="100" y="32" font-family="Arial, sans-serif" font-size="20" font-weight="bold" 
        text-anchor="middle" fill="url(#poatcGradient)">POATC</text>
  
  <!-- Subtitle -->
  <text x="100" y="42" font-family="Arial, sans-serif" font-size="8" 
        text-anchor="middle" fill="#888">Blockchain Explorer</text>
</svg>
"@
    Set-Content -Path (Join-Path $imagesPath "logo.svg") -Value $defaultLogo -Encoding UTF8
    Write-Host "Tao logo POATC mac dinh" -ForegroundColor Green
}

# Thay the logo light (cho dark mode)
if ($LogoLightPath) {
    Replace-Logo -SourcePath $LogoLightPath -TargetFileName "logo-light.svg" -Description "logo light mode"
} else {
    # Tao logo light mac dinh
    $defaultLogoLight = @"
<svg width="200" height="50" viewBox="0 0 200 50" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="poatcLightGradient" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#ff6b35;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#2563eb;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background -->
  <rect width="200" height="50" fill="#ffffff" rx="8"/>
  
  <!-- POATC Text -->
  <text x="100" y="32" font-family="Arial, sans-serif" font-size="20" font-weight="bold" 
        text-anchor="middle" fill="url(#poatcLightGradient)">POATC</text>
  
  <!-- Subtitle -->
  <text x="100" y="42" font-family="Arial, sans-serif" font-size="8" 
        text-anchor="middle" fill="#666">Blockchain Explorer</text>
</svg>
"@
    Set-Content -Path (Join-Path $imagesPath "logo-light.svg") -Value $defaultLogoLight -Encoding UTF8
    Write-Host "Tao logo light POATC mac dinh" -ForegroundColor Green
}

# Thay the favicon
if ($FaviconPath) {
    Replace-Logo -SourcePath $FaviconPath -TargetFileName "favicon.ico" -Description "favicon"
} else {
    Write-Host "Tao favicon mac dinh..." -ForegroundColor Yellow
    
    # Tao favicon SVG don gian
    $defaultFavicon = @"
<svg width="32" height="32" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
  <rect width="32" height="32" fill="#1a1a1a" rx="6"/>
  <text x="16" y="22" font-family="Arial, sans-serif" font-size="14" font-weight="bold" 
        text-anchor="middle" fill="#ff6b35">P</text>
</svg>
"@
    Set-Content -Path (Join-Path $imagesPath "favicon.svg") -Value $defaultFavicon -Encoding UTF8
    Write-Host "Tao favicon SVG mac dinh" -ForegroundColor Green
}

# Tao apple-touch-icon tu logo chinh
Write-Host "Tao apple-touch-icon..." -ForegroundColor Yellow
$appleTouchIcon = @"
<svg width="180" height="180" viewBox="0 0 180 180" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="poatcAppleGradient" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#ff6b35;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#2563eb;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background -->
  <rect width="180" height="180" fill="#1a1a1a" rx="20"/>
  
  <!-- POATC Text -->
  <text x="90" y="110" font-family="Arial, sans-serif" font-size="36" font-weight="bold" 
        text-anchor="middle" fill="url(#poatcAppleGradient)">POATC</text>
  
  <!-- Subtitle -->
  <text x="90" y="135" font-family="Arial, sans-serif" font-size="12" 
        text-anchor="middle" fill="#888">Explorer</text>
</svg>
"@
Set-Content -Path (Join-Path $imagesPath "apple-touch-icon.svg") -Value $appleTouchIcon -Encoding UTF8
Write-Host "Tao apple-touch-icon" -ForegroundColor Green

Write-Host ""
Write-Host "HOAN THANH THAY LOGO!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Logo chinh: logo.svg" -ForegroundColor Cyan
Write-Host "Logo light mode: logo-light.svg" -ForegroundColor Cyan
Write-Host "Favicon: favicon.svg" -ForegroundColor Cyan
Write-Host "Apple touch icon: apple-touch-icon.svg" -ForegroundColor Cyan
Write-Host ""
Write-Host "Cac buoc tiep theo:" -ForegroundColor Yellow
Write-Host "1. Rebuild frontend: npm run build" -ForegroundColor White
Write-Host "2. Restart Docker: docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White
Write-Host ""
Write-Host "Backup files da duoc tao voi extension .backup" -ForegroundColor Yellow