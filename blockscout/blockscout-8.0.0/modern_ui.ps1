# Script Tuy Chinh Giao Dien Hien Dai cho Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$ProjectName = "POATC",
    [string]$ProjectDescription = "Blockchain Explorer",
    [string]$PrimaryColor = "#6366f1",
    [string]$SecondaryColor = "#8b5cf6",
    [string]$AccentColor = "#06b6d4",
    [string]$BackgroundColor = "#f8fafc",
    [string]$DarkMode = "true"
)

Write-Host "Tuy Chinh Giao Dien Hien Dai cho $ProjectName..." -ForegroundColor Magenta

$cssPath = "apps\block_scout_web\assets\css"
$themePath = "$cssPath\theme"
$imagesPath = "apps\block_scout_web\assets\static\images"

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
    $backupPath = "$themeFile.backup.$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    Copy-Item $themeFile $backupPath
    Write-Host "Backup theme hien tai: $backupPath" -ForegroundColor Green
}

# Tao logo hien dai
Write-Host "Tao logo hien dai cho $ProjectName..." -ForegroundColor Yellow
$modernLogo = @"
<svg width="220" height="60" viewBox="0 0 220 60" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="modernGradient" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:$PrimaryColor;stop-opacity:1" />
      <stop offset="50%" style="stop-color:$SecondaryColor;stop-opacity:1" />
      <stop offset="100%" style="stop-color:$AccentColor;stop-opacity:1" />
    </linearGradient>
    <filter id="glow">
      <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
      <feMerge> 
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
  
  <!-- Background với glass effect -->
  <rect width="220" height="60" fill="rgba(255,255,255,0.1)" rx="12" stroke="url(#modernGradient)" stroke-width="2"/>
  
  <!-- Icon hiện đại -->
  <circle cx="30" cy="30" r="12" fill="url(#modernGradient)" filter="url(#glow)"/>
  <path d="M22 30 L26 34 L38 22" stroke="white" stroke-width="3" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
  
  <!-- Project Name -->
  <text x="55" y="35" font-family="'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif" 
        font-size="24" font-weight="700" fill="url(#modernGradient)" filter="url(#glow)">$ProjectName</text>
  
  <!-- Subtitle -->
  <text x="55" y="48" font-family="'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif" 
        font-size="10" font-weight="500" fill="#64748b" opacity="0.8">$ProjectDescription</text>
</svg>
"@

Set-Content -Path (Join-Path $imagesPath "logo.svg") -Value $modernLogo -Encoding UTF8

# Tao logo light mode
$modernLogoLight = @"
<svg width="220" height="60" viewBox="0 0 220 60" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="modernLightGradient" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:$PrimaryColor;stop-opacity:1" />
      <stop offset="50%" style="stop-color:$SecondaryColor;stop-opacity:1" />
      <stop offset="100%" style="stop-color:$AccentColor;stop-opacity:1" />
    </linearGradient>
    <filter id="glowLight">
      <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
      <feMerge> 
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
  
  <!-- Background light -->
  <rect width="220" height="60" fill="rgba(248,250,252,0.9)" rx="12" stroke="url(#modernLightGradient)" stroke-width="2"/>
  
  <!-- Icon hiện đại -->
  <circle cx="30" cy="30" r="12" fill="url(#modernLightGradient)" filter="url(#glowLight)"/>
  <path d="M22 30 L26 34 L38 22" stroke="white" stroke-width="3" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
  
  <!-- Project Name -->
  <text x="55" y="35" font-family="'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif" 
        font-size="24" font-weight="700" fill="url(#modernLightGradient)">$ProjectName</text>
  
  <!-- Subtitle -->
  <text x="55" y="48" font-family="'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif" 
        font-size="10" font-weight="500" fill="#64748b" opacity="0.8">$ProjectDescription</text>
</svg>
"@

Set-Content -Path (Join-Path $imagesPath "logo-light.svg") -Value $modernLogoLight -Encoding UTF8

# Tao theme hien dai
Write-Host "Tao theme hien dai..." -ForegroundColor Yellow

$modernTheme = @"
// Modern UI Theme cho $ProjectName
// Generated on: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

// ===== MODERN COLOR PALETTE =====
`$primary-color: $PrimaryColor;
`$secondary-color: $SecondaryColor;
`$accent-color: $AccentColor;
`$background-color: $BackgroundColor;
`$surface-color: #ffffff;
`$surface-elevated: #ffffff;
`$text-primary: #0f172a;
`$text-secondary: #64748b;
`$text-muted: #94a3b8;
`$border-color: #e2e8f0;
`$border-light: #f1f5f9;
`$shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
`$shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
`$shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
`$shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

// ===== SUCCESS/ERROR COLORS =====
`$success-color: #10b981;
`$warning-color: #f59e0b;
`$error-color: #ef4444;
`$info-color: `$accent-color;

// ===== MODERN LAYOUT =====
`$container-max-width: 1280px;
`$border-radius-sm: 6px;
`$border-radius-md: 8px;
`$border-radius-lg: 12px;
`$border-radius-xl: 16px;
`$spacing-xs: 4px;
`$spacing-sm: 8px;
`$spacing-md: 16px;
`$spacing-lg: 24px;
`$spacing-xl: 32px;
`$spacing-2xl: 48px;

// ===== MODERN TYPOGRAPHY =====
`$font-family-primary: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
`$font-family-mono: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
`$font-weight-light: 300;
`$font-weight-normal: 400;
`$font-weight-medium: 500;
`$font-weight-semibold: 600;
`$font-weight-bold: 700;
`$font-weight-extrabold: 800;

// ===== MODERN HEADER =====
.header {
  background: linear-gradient(135deg, `$primary-color 0%, `$secondary-color 50%, `$accent-color 100%);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255,255,255,0.1);
  box-shadow: `$shadow-lg;
  position: sticky;
  top: 0;
  z-index: 1000;
  
  .navbar {
    padding: `$spacing-md 0;
    
    .navbar-brand {
      .logo {
        max-height: 45px;
        filter: brightness(0) invert(1);
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        
        &:hover {
          transform: scale(1.05) rotate(2deg);
          filter: brightness(0) invert(1) drop-shadow(0 0 10px rgba(255,255,255,0.3));
        }
      }
    }
    
    .navbar-nav {
      .nav-link {
        color: rgba(255,255,255,0.9) !important;
        font-weight: `$font-weight-medium;
        padding: `$spacing-sm `$spacing-md;
        border-radius: `$border-radius-md;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        position: relative;
        overflow: hidden;
        
        &::before {
          content: '';
          position: absolute;
          top: 0;
          left: -100%;
          width: 100%;
          height: 100%;
          background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
          transition: left 0.5s;
        }
        
        &:hover {
          color: white !important;
          background: rgba(255,255,255,0.1);
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.15);
          
          &::before {
            left: 100%;
          }
        }
        
        &.active {
          background: rgba(255,255,255,0.2);
          color: white !important;
          box-shadow: inset 0 2px 4px rgba(0,0,0,0.1);
        }
      }
    }
  }
}

// ===== MODERN CARDS =====
.card {
  background: `$surface-color;
  border: 1px solid `$border-light;
  border-radius: `$border-radius-lg;
  box-shadow: `$shadow-sm;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: linear-gradient(90deg, `$primary-color, `$secondary-color, `$accent-color);
    opacity: 0;
    transition: opacity 0.3s ease;
  }
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: `$shadow-xl;
    border-color: `$primary-color;
    
    &::before {
      opacity: 1;
    }
  }
  
  .card-header {
    background: linear-gradient(135deg, `$primary-color, `$secondary-color);
    color: white;
    border: none;
    padding: `$spacing-lg;
    position: relative;
    overflow: hidden;
    
    &::before {
      content: '';
      position: absolute;
      top: -50%;
      left: -50%;
      width: 200%;
      height: 200%;
      background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
      animation: shimmer 3s infinite;
    }
    
    h1, h2, h3, h4, h5, h6 {
      color: white;
      margin: 0;
      font-weight: `$font-weight-semibold;
      position: relative;
      z-index: 1;
    }
  }
  
  .card-body {
    padding: `$spacing-lg;
  }
}

@keyframes shimmer {
  0% { transform: translateX(-100%) translateY(-100%) rotate(30deg); }
  100% { transform: translateX(100%) translateY(100%) rotate(30deg); }
}

// ===== MODERN BUTTONS =====
.btn {
  font-family: `$font-family-primary;
  font-weight: `$font-weight-medium;
  border-radius: `$border-radius-md;
  padding: `$spacing-sm `$spacing-lg;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  border: none;
  cursor: pointer;
  
  &::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    background: rgba(255,255,255,0.2);
    border-radius: 50%;
    transform: translate(-50%, -50%);
    transition: width 0.3s, height 0.3s;
  }
  
  &:hover::before {
    width: 300px;
    height: 300px;
  }
  
  span {
    position: relative;
    z-index: 1;
  }
}

.btn-primary {
  background: linear-gradient(135deg, `$primary-color, `$secondary-color);
  color: white;
  box-shadow: `$shadow-md;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: `$shadow-lg;
    background: linear-gradient(135deg, lighten(`$primary-color, 5%), lighten(`$secondary-color, 5%));
  }
  
  &:active {
    transform: translateY(0);
    box-shadow: `$shadow-sm;
  }
}

.btn-secondary {
  background: `$surface-elevated;
  color: `$text-primary;
  border: 1px solid `$border-color;
  box-shadow: `$shadow-sm;
  
  &:hover {
    background: `$border-light;
    transform: translateY(-1px);
    box-shadow: `$shadow-md;
  }
}

.btn-outline {
  background: transparent;
  color: `$primary-color;
  border: 2px solid `$primary-color;
  
  &:hover {
    background: `$primary-color;
    color: white;
    transform: translateY(-2px);
    box-shadow: `$shadow-lg;
  }
}

// ===== MODERN TABLES =====
.table {
  background: `$surface-color;
  border-radius: `$border-radius-lg;
  overflow: hidden;
  box-shadow: `$shadow-sm;
  
  thead {
    background: linear-gradient(135deg, `$primary-color, `$secondary-color);
    
    th {
      color: white;
      font-weight: `$font-weight-semibold;
      border: none;
      padding: `$spacing-lg;
      text-transform: uppercase;
      font-size: 0.875rem;
      letter-spacing: 0.05em;
      position: relative;
      
      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        height: 2px;
        background: rgba(255,255,255,0.3);
      }
    }
  }
  
  tbody {
    tr {
      transition: all 0.2s ease;
      border-bottom: 1px solid `$border-light;
      
      &:hover {
        background: linear-gradient(135deg, rgba(`$primary-color, 0.05), rgba(`$accent-color, 0.05));
        transform: scale(1.01);
      }
      
      &:last-child {
        border-bottom: none;
      }
      
      td {
        padding: `$spacing-lg;
        border: none;
        color: `$text-primary;
        font-weight: `$font-weight-normal;
      }
    }
  }
}

// ===== MODERN FORMS =====
.form-control {
  border: 2px solid `$border-color;
  border-radius: `$border-radius-md;
  padding: `$spacing-md;
  font-family: `$font-family-primary;
  font-weight: `$font-weight-normal;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: `$surface-color;
  
  &:focus {
    border-color: `$primary-color;
    box-shadow: 0 0 0 3px rgba(`$primary-color, 0.1);
    outline: none;
    transform: translateY(-1px);
  }
  
  &::placeholder {
    color: `$text-muted;
    font-weight: `$font-weight-normal;
  }
}

.form-label {
  font-weight: `$font-weight-medium;
  color: `$text-primary;
  margin-bottom: `$spacing-sm;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

// ===== MODERN BADGES =====
.badge {
  border-radius: `$border-radius-xl;
  font-weight: `$font-weight-medium;
  padding: `$spacing-xs `$spacing-md;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  box-shadow: `$shadow-sm;
}

.badge-success {
  background: linear-gradient(135deg, `$success-color, lighten(`$success-color, 10%));
  color: white;
}

.badge-warning {
  background: linear-gradient(135deg, `$warning-color, lighten(`$warning-color, 10%));
  color: white;
}

.badge-error {
  background: linear-gradient(135deg, `$error-color, lighten(`$error-color, 10%));
  color: white;
}

.badge-info {
  background: linear-gradient(135deg, `$info-color, lighten(`$info-color, 10%));
  color: white;
}

// ===== MODERN ALERTS =====
.alert {
  border-radius: `$border-radius-lg;
  border: none;
  font-weight: `$font-weight-medium;
  padding: `$spacing-lg;
  box-shadow: `$shadow-sm;
  position: relative;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 4px;
    height: 100%;
  }
}

.alert-success {
  background: linear-gradient(135deg, rgba(`$success-color, 0.1), rgba(`$success-color, 0.05));
  color: darken(`$success-color, 20%);
  border-left: 4px solid `$success-color;
}

.alert-warning {
  background: linear-gradient(135deg, rgba(`$warning-color, 0.1), rgba(`$warning-color, 0.05));
  color: darken(`$warning-color, 20%);
  border-left: 4px solid `$warning-color;
}

.alert-danger {
  background: linear-gradient(135deg, rgba(`$error-color, 0.1), rgba(`$error-color, 0.05));
  color: darken(`$error-color, 20%);
  border-left: 4px solid `$error-color;
}

.alert-info {
  background: linear-gradient(135deg, rgba(`$info-color, 0.1), rgba(`$info-color, 0.05));
  color: darken(`$info-color, 20%);
  border-left: 4px solid `$info-color;
}

// ===== MODERN NAVIGATION =====
.nav-tabs {
  border-bottom: 2px solid `$border-light;
  
  .nav-link {
    border: none;
    color: `$text-secondary;
    font-weight: `$font-weight-medium;
    padding: `$spacing-md `$spacing-lg;
    border-radius: `$border-radius-md `$border-radius-md 0 0;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    
    &:hover {
      color: `$primary-color;
      background: rgba(`$primary-color, 0.05);
      transform: translateY(-2px);
    }
    
    &.active {
      color: `$primary-color;
      background: `$surface-color;
      border-bottom: 3px solid `$primary-color;
      box-shadow: `$shadow-sm;
    }
  }
}

// ===== MODERN PAGINATION =====
.pagination {
  .page-link {
    color: `$primary-color;
    border: 1px solid `$border-color;
    margin: 0 `$spacing-xs;
    border-radius: `$border-radius-md;
    padding: `$spacing-sm `$spacing-md;
    font-weight: `$font-weight-medium;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    
    &:hover {
      background: `$primary-color;
      color: white;
      transform: translateY(-2px);
      box-shadow: `$shadow-md;
      border-color: `$primary-color;
    }
  }
  
  .page-item.active .page-link {
    background: linear-gradient(135deg, `$primary-color, `$secondary-color);
    border-color: `$primary-color;
    box-shadow: `$shadow-md;
  }
}

// ===== MODERN SCROLLBAR =====
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: `$border-light;
  border-radius: `$border-radius-sm;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, `$primary-color, `$secondary-color);
  border-radius: `$border-radius-sm;
  transition: background 0.3s ease;
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, lighten(`$primary-color, 10%), lighten(`$secondary-color, 10%));
}

// ===== MODERN ANIMATIONS =====
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(30px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

.fade-in-up {
  animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-in-right {
  animation: slideInRight 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.pulse {
  animation: pulse 2s infinite;
}

// ===== MODERN UTILITIES =====
.text-gradient {
  background: linear-gradient(135deg, `$primary-color, `$secondary-color, `$accent-color);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: `$font-weight-bold;
}

.bg-gradient {
  background: linear-gradient(135deg, `$primary-color, `$secondary-color, `$accent-color);
}

.glass-effect {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

// ===== RESPONSIVE DESIGN =====
@media (max-width: 768px) {
  .container {
    padding: `$spacing-md;
  }
  
  .card {
    margin-bottom: `$spacing-lg;
    border-radius: `$border-radius-md;
  }
  
  .header .navbar {
    padding: `$spacing-sm 0;
    
    .navbar-brand .logo {
      max-height: 35px;
    }
  }
  
  .btn {
    padding: `$spacing-md `$spacing-lg;
    font-size: 1rem;
  }
}

// ===== DARK MODE =====
@media (prefers-color-scheme: dark) {
  `$background-color: #0f172a;
  `$surface-color: #1e293b;
  `$surface-elevated: #334155;
  `$text-primary: #f8fafc;
  `$text-secondary: #cbd5e1;
  `$text-muted: #64748b;
  `$border-color: #334155;
  `$border-light: #475569;
}

// ===== HIDE ADS =====
.ad-container,
.js-ad-dependant-mb-2,
.js-ad-dependant-mb-3,
.js-ad-dependant-mb-5-reverse,
.ad-banner,
.text-ad,
[class*="ad-"],
[id*="ad-"],
[class*="banner-"],
[class*="coinzilla"],
[class*="adsense"],
#coinzilla-widget,
.google-ads,
.adsbygoogle {
  display: none !important;
}

// ===== CUSTOM PROJECT BRANDING =====
.project-branding {
  .project-name {
    font-family: `$font-family-primary;
    font-weight: `$font-weight-extrabold;
    font-size: 2rem;
    background: linear-gradient(135deg, `$primary-color, `$secondary-color, `$accent-color);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    text-align: center;
    margin-bottom: `$spacing-sm;
  }
  
  .project-description {
    color: `$text-secondary;
    font-weight: `$font-weight-medium;
    text-align: center;
    font-size: 1.1rem;
    margin-bottom: `$spacing-xl;
  }
}
"@

Set-Content -Path $themeFile -Value $modernTheme -Encoding UTF8
Write-Host "Tao theme hien dai thanh cong" -ForegroundColor Green

# Disable ads
Write-Host "Disable quang cao..." -ForegroundColor Yellow
& ".\disable_ads.ps1"

Write-Host ""
Write-Host "HOAN THANH TUY CHINH GIAO DIEN HIEN DAI!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Ten du an: $ProjectName" -ForegroundColor Cyan
Write-Host "Mo ta: $ProjectDescription" -ForegroundColor Cyan
Write-Host "Mau chinh: $PrimaryColor" -ForegroundColor Cyan
Write-Host "Mau phu: $SecondaryColor" -ForegroundColor Cyan
Write-Host "Mau accent: $AccentColor" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tinh nang hien dai:" -ForegroundColor Yellow
Write-Host "  • Glass effect va backdrop blur" -ForegroundColor White
Write-Host "  • Gradient backgrounds" -ForegroundColor White
Write-Host "  • Smooth animations va transitions" -ForegroundColor White
Write-Host "  • Modern cards va buttons" -ForegroundColor White
Write-Host "  • Custom scrollbar" -ForegroundColor White
Write-Host "  • Responsive design" -ForegroundColor White
Write-Host "  • Dark mode support" -ForegroundColor White
Write-Host "  • Hidden ads" -ForegroundColor White
Write-Host ""
Write-Host "Cac buoc tiep theo:" -ForegroundColor Yellow
Write-Host "1. cd docker-compose" -ForegroundColor White
Write-Host "2. docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White
