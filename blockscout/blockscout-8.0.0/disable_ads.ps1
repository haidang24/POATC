# Script Tu Dong Disable Quang Cao trong Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

Write-Host "Disable Quang Cao trong Blockscout..." -ForegroundColor Red

$assetsPath = "apps\block_scout_web\assets"
$jsLibPath = "$assetsPath\js\lib"

# Kiem tra thu muc
if (-not (Test-Path $jsLibPath)) {
    Write-Host "Khong tim thay thu muc $jsLibPath" -ForegroundColor Red
    exit 1
}

Write-Host "Dang o thu muc: $jsLibPath" -ForegroundColor Cyan

# Backup files goc
Write-Host "Backup files goc..." -ForegroundColor Yellow
$filesToBackup = @("ad.js", "banner.js", "text_ad.js")
foreach ($file in $filesToBackup) {
    $filePath = Join-Path $jsLibPath $file
    if (Test-Path $filePath) {
        $backupPath = "$filePath.backup"
        if (-not (Test-Path $backupPath)) {
            Copy-Item $filePath $backupPath
            Write-Host "Backup $file" -ForegroundColor Green
        } else {
            Write-Host "$file da co backup" -ForegroundColor Blue
        }
    }
}

# Disable ads trong ad.js
Write-Host "Disable ads trong ad.js..." -ForegroundColor Yellow
$adJsPath = Join-Path $jsLibPath "ad.js"
if (Test-Path $adJsPath) {
    $adJsContent = @"
import `$ from 'jquery'
import customAds from './custom_ad.json'

function countImpressions (impressionUrl) {
  // Disabled - khong dem impressions
}

function showAd () {
  // Luon tra ve false de disable ads
  return false;
}

function fetchTextAdData () {
  // Disabled - khong fetch text ads
}

export { showAd, fetchTextAdData, countImpressions }
"@
    Set-Content -Path $adJsPath -Value $adJsContent -Encoding UTF8
    Write-Host "Disabled ads trong ad.js" -ForegroundColor Green
}

# Disable banner ads trong banner.js
Write-Host "Disable banner ads trong banner.js..." -ForegroundColor Yellow
$bannerJsPath = Join-Path $jsLibPath "banner.js"
if (Test-Path $bannerJsPath) {
    $bannerJsContent = @"
import `$ from 'jquery'
import { showAd } from './ad.js'

// Disabled - khong hien thi banner ads
`$('.ad-container').hide()

// Comment out tat ca ads code
/*
if (showAd()) {
  window.coinzilla_display = window.coinzilla_display || []
  var c_display_preferences = {}
  c_display_preferences.zone = '26660bf627543e46851'
  c_display_preferences.width = '728'
  c_display_preferences.height = '90'
  window.coinzilla_display.push(c_display_preferences)
  `$('.ad-container').show()
} else {
  `$('.ad-container').hide()
}
*/
"@
    Set-Content -Path $bannerJsPath -Value $bannerJsContent -Encoding UTF8
    Write-Host "Disabled banner ads trong banner.js" -ForegroundColor Green
}

# Disable text ads trong text_ad.js
Write-Host "Disable text ads trong text_ad.js..." -ForegroundColor Yellow
$textAdJsPath = Join-Path $jsLibPath "text_ad.js"
if (Test-Path $textAdJsPath) {
    $textAdJsContent = @"
import `$ from 'jquery'
import { showAd, fetchTextAdData } from './ad.js'

`$(function () {
  // Disabled - khong hien thi text ads
  // Comment out ads code
  /*
  if (showAd()) {
    fetchTextAdData()
  }
  */
})
"@
    Set-Content -Path $textAdJsPath -Value $textAdJsContent -Encoding UTF8
    Write-Host "Disabled text ads trong text_ad.js" -ForegroundColor Green
}

# Them CSS de an ad containers
Write-Host "Them CSS de an ad containers..." -ForegroundColor Yellow
$themePath = "$assetsPath\css\theme\poatc-theme.scss"
if (Test-Path $themePath) {
    $hideAdsCss = @"

// Hide all ads containers
.ad-container,
.js-ad-dependant-mb-2,
.js-ad-dependant-mb-3,
.js-ad-dependant-mb-5-reverse,
.ad-banner,
.text-ad {
  display: none !important;
}

// Hide elements with ad-related classes
[class*="ad-"],
[id*="ad-"],
[class*="banner-"],
[class*="coinzilla"],
[class*="adsense"] {
  display: none !important;
}

// Hide specific ad containers
#coinzilla-widget,
.google-ads,
.adsbygoogle {
  display: none !important;
}
"@
    
    $currentContent = Get-Content $themePath -Raw
    if ($currentContent -notmatch "Hide all ads containers") {
        Add-Content -Path $themePath -Value $hideAdsCss -Encoding UTF8
        Write-Host "Them CSS an ads vao theme" -ForegroundColor Green
    } else {
        Write-Host "CSS an ads da co san" -ForegroundColor Blue
    }
}

Write-Host ""
Write-Host "HOAN THANH DISABLE QUANG CAO!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Disabled ads trong ad.js" -ForegroundColor Cyan
Write-Host "Disabled banner ads trong banner.js" -ForegroundColor Cyan
Write-Host "Disabled text ads trong text_ad.js" -ForegroundColor Cyan
Write-Host "Them CSS an ad containers" -ForegroundColor Cyan
Write-Host ""
Write-Host "Cac buoc tiep theo:" -ForegroundColor Yellow
Write-Host "1. Rebuild frontend: npm run build" -ForegroundColor White
Write-Host "2. Restart Docker: docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White
Write-Host ""
Write-Host "Backup files da duoc tao voi extension .backup" -ForegroundColor Yellow