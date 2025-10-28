# Script Tuy Chinh Theme Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$PrimaryColor = "#ff6b35",
    [string]$SecondaryColor = "#2563eb",
    [string]$BackgroundColor = "#ffffff",
    [string]$TextColor = "#2d3748",
    [string]$BorderColor = "#e2e8f0",
    [string]$ThemeName = "POATC"
)

Write-Host "Tuy Chinh Theme Blockscout..." -ForegroundColor Magenta

$cssPath = "apps\block_scout_web\assets\css"
$themePath = "$cssPath\theme"

# Kiem tra thu muc
if (-not (Test-Path $themePath)) {
    Write-Host "Khong tim thay thu muc $themePath" -ForegroundColor Red
    exit 1
}

Write-Host "Dang o thu muc: $themePath" -ForegroundColor Cyan

# Backup theme hien tai
Write-Host "Backup theme hien tai..." -ForegroundColor Yellow
$themeFile = Join-Path $themePath "poatc-theme.scss"
if (Test-Path $themeFile) {
    $backupPath = "$themeFile.backup"
    if (-not (Test-Path $backupPath)) {
        Copy-Item $themeFile $backupPath
        Write-Host "Backup theme hien tai" -ForegroundColor Green
    }
}

# Tao theme moi
Write-Host "Tao theme moi voi mau sac tuy chinh..." -ForegroundColor Yellow

$customTheme = @"
// $ThemeName Custom Theme
// Generated on: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

// ===== COLOR VARIABLES =====
`$primary-color: $PrimaryColor;
`$secondary-color: $SecondaryColor;
`$background-color: $BackgroundColor;
`$card-background: lighten(`$background-color, 2%);
`$text-color: $TextColor;
`$text-muted: darken(`$text-color, 20%);
`$border-color: $BorderColor;
`$border-light: lighten(`$border-color, 10%);

// ===== DERIVED COLORS =====
`$success-color: #10b981;
`$warning-color: #f59e0b;
`$error-color: #ef4444;
`$info-color: `$secondary-color;

// ===== LAYOUT VARIABLES =====
`$container-padding: 20px;
`$card-padding: 16px;
`$border-radius: 8px;
`$button-border-radius: 6px;
`$font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;

// ===== HEADER STYLING =====
.header {
  background: linear-gradient(135deg, `$primary-color, `$secondary-color);
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  border-bottom: 3px solid `$primary-color;
  
  .navbar-brand {
    .logo {
      max-height: 40px;
      filter: brightness(0) invert(1);
      transition: all 0.3s ease;
      
      &:hover {
        transform: scale(1.05);
      }
    }
  }
  
  .navbar-nav {
    .nav-link {
      color: rgba(255,255,255,0.9) !important;
      font-weight: 500;
      transition: all 0.3s ease;
      
      &:hover {
        color: white !important;
        transform: translateY(-1px);
      }
    }
  }
}

// ===== CARD STYLING =====
.card {
  background: `$card-background;
  border: 1px solid `$border-color;
  border-radius: `$border-radius;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  transition: all 0.3s ease;
  
  &:hover {
    box-shadow: 0 4px 16px rgba(0,0,0,0.1);
    transform: translateY(-2px);
    border-color: `$primary-color;
  }
  
  .card-header {
    background: linear-gradient(135deg, `$primary-color, `$secondary-color);
    color: white;
    border-radius: `$border-radius `$border-radius 0 0;
    font-weight: 600;
    
    h1, h2, h3, h4, h5, h6 {
      color: white;
      margin: 0;
    }
  }
  
  .card-body {
    padding: `$card-padding;
  }
}

// ===== BUTTON STYLING =====
.btn-primary {
  background: linear-gradient(135deg, `$primary-color, darken(`$primary-color, 10%));
  border: none;
  border-radius: `$button-border-radius;
  font-weight: 600;
  padding: 10px 20px;
  transition: all 0.3s ease;
  
  &:hover {
    background: linear-gradient(135deg, darken(`$primary-color, 5%), darken(`$primary-color, 15%));
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  }
  
  &:active {
    transform: translateY(0);
  }
}

.btn-secondary {
  background: `$secondary-color;
  border: none;
  border-radius: `$button-border-radius;
  font-weight: 500;
  
  &:hover {
    background: darken(`$secondary-color, 10%);
    transform: translateY(-1px);
  }
}

// ===== TABLE STYLING =====
.table {
  background: `$card-background;
  border-radius: `$border-radius;
  overflow: hidden;
  
  thead {
    background: linear-gradient(135deg, `$primary-color, `$secondary-color);
    
    th {
      color: white;
      font-weight: 600;
      border: none;
      padding: 12px;
    }
  }
  
  tbody {
    tr {
      transition: all 0.2s ease;
      
      &:hover {
        background-color: lighten(`$primary-color, 45%);
        transform: scale(1.01);
      }
      
      td {
        border-color: `$border-light;
        padding: 12px;
      }
    }
  }
}

// ===== BADGE STYLING =====
.badge {
  border-radius: 20px;
  font-weight: 500;
  padding: 6px 12px;
}

.badge-success {
  background: `$success-color;
}

.badge-warning {
  background: `$warning-color;
}

.badge-error {
  background: `$error-color;
}

.badge-info {
  background: `$info-color;
}

// ===== FORM STYLING =====
.form-control {
  border: 2px solid `$border-color;
  border-radius: `$border-radius;
  padding: 10px 12px;
  transition: all 0.3s ease;
  
  &:focus {
    border-color: `$primary-color;
    box-shadow: 0 0 0 3px rgba(255, 107, 53, 0.1);
  }
}

// ===== ALERT STYLING =====
.alert {
  border-radius: `$border-radius;
  border: none;
  font-weight: 500;
}

.alert-success {
  background: lighten(`$success-color, 40%);
  color: darken(`$success-color, 30%);
}

.alert-warning {
  background: lighten(`$warning-color, 40%);
  color: darken(`$warning-color, 30%);
}

.alert-danger {
  background: lighten(`$error-color, 40%);
  color: darken(`$error-color, 30%);
}

.alert-info {
  background: lighten(`$info-color, 40%);
  color: darken(`$info-color, 30%);
}

// ===== NAVIGATION STYLING =====
.nav-tabs {
  border-bottom: 2px solid `$border-color;
  
  .nav-link {
    border: none;
    color: `$text-muted;
    font-weight: 500;
    transition: all 0.3s ease;
    
    &:hover {
      color: `$primary-color;
      border-bottom: 2px solid `$primary-color;
    }
    
    &.active {
      color: `$primary-color;
      background: none;
      border-bottom: 2px solid `$primary-color;
    }
  }
}

// ===== PAGINATION STYLING =====
.pagination {
  .page-link {
    color: `$primary-color;
    border: 1px solid `$border-color;
    margin: 0 2px;
    border-radius: `$button-border-radius;
    
    &:hover {
      background: `$primary-color;
      color: white;
      transform: translateY(-1px);
    }
  }
  
  .page-item.active .page-link {
    background: `$primary-color;
    border-color: `$primary-color;
  }
}

// ===== UTILITY CLASSES =====
.text-primary {
  color: `$primary-color !important;
}

.bg-primary {
  background-color: `$primary-color !important;
}

.border-primary {
  border-color: `$primary-color !important;
}

// ===== DARK MODE SUPPORT =====
@media (prefers-color-scheme: dark) {
  `$background-color: #1a1a1a;
  `$text-color: #ffffff;
  `$border-color: #333333;
  `$card-background: #2d2d2d;
}

// ===== RESPONSIVE DESIGN =====
@media (max-width: 768px) {
  .container {
    padding: 10px;
  }
  
  .card {
    margin-bottom: 15px;
  }
  
  .table-responsive {
    border-radius: `$border-radius;
  }
}

// ===== ANIMATIONS =====
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-in {
  animation: fadeIn 0.5s ease-out;
}

// ===== CUSTOM SCROLLBAR =====
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: `$border-light;
}

::-webkit-scrollbar-thumb {
  background: `$primary-color;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: darken(`$primary-color, 10%);
}

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

Set-Content -Path $themeFile -Value $customTheme -Encoding UTF8
Write-Host "Tao theme tuy chinh thanh cong" -ForegroundColor Green

# Hien thi thong tin theme
Write-Host ""
Write-Host "THEME MOI DA DUOC TAO!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Ten theme: $ThemeName" -ForegroundColor Cyan
Write-Host "Mau chinh: $PrimaryColor" -ForegroundColor Cyan
Write-Host "Mau phu: $SecondaryColor" -ForegroundColor Cyan
Write-Host "Mau nen: $BackgroundColor" -ForegroundColor Cyan
Write-Host "Mau text: $TextColor" -ForegroundColor Cyan
Write-Host "Mau border: $BorderColor" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tinh nang da bao gom:" -ForegroundColor Yellow
Write-Host "  • Gradient backgrounds" -ForegroundColor White
Write-Host "  • Hover animations" -ForegroundColor White
Write-Host "  • Custom buttons & cards" -ForegroundColor White
Write-Host "  • Responsive design" -ForegroundColor White
Write-Host "  • Dark mode support" -ForegroundColor White
Write-Host "  • Custom scrollbar" -ForegroundColor White
Write-Host "  • Hidden ads" -ForegroundColor White
Write-Host ""
Write-Host "Cac buoc tiep theo:" -ForegroundColor Yellow
Write-Host "1. Rebuild frontend: npm run build" -ForegroundColor White
Write-Host "2. Restart Docker: docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White
Write-Host ""
Write-Host "Backup theme cu: poatc-theme.scss.backup" -ForegroundColor Yellow