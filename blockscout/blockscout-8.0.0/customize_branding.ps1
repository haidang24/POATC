# Script Tuy Chinh Branding va Text trong Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$ProjectName = "POATC",
    [string]$ProjectDescription = "Blockchain Explorer",
    [string]$Tagline = "Secure, Fast, Decentralized",
    [string]$FooterText = "Powered by POATC Blockchain",
    [string]$CompanyName = "POATC Team"
)

Write-Host "Tuy Chinh Branding cho $ProjectName..." -ForegroundColor Magenta

$frontendPath = "apps\block_scout_web\assets"
$templatesPath = "apps\block_scout_web\lib\block_scout_web\templates"
$staticPath = "$frontendPath\static"

# Kiem tra thu muc
if (-not (Test-Path $frontendPath)) {
    Write-Host "Khong tim thay thu muc frontend" -ForegroundColor Red
    exit 1
}

Write-Host "Dang o thu muc: $frontendPath" -ForegroundColor Cyan

# Tao favicon moi
Write-Host "Tao favicon moi..." -ForegroundColor Yellow
$faviconContent = @"
<svg width="32" height="32" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="faviconGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#6366f1;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#8b5cf6;stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="32" height="32" fill="#0f172a" rx="6"/>
  <circle cx="16" cy="16" r="10" fill="url(#faviconGradient)"/>
  <text x="16" y="20" font-family="Arial, sans-serif" font-size="14" font-weight="bold" 
        text-anchor="middle" fill="white">P</text>
</svg>
"@

Set-Content -Path (Join-Path $staticPath "images\favicon.svg") -Value $faviconContent -Encoding UTF8

# Tao apple-touch-icon
Write-Host "Tao apple-touch-icon..." -ForegroundColor Yellow
$appleIconContent = @"
<svg width="180" height="180" viewBox="0 0 180 180" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="appleGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#6366f1;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#8b5cf6;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#06b6d4;stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="180" height="180" fill="#0f172a" rx="20"/>
  <circle cx="90" cy="90" r="50" fill="url(#appleGradient)"/>
  <text x="90" y="100" font-family="Arial, sans-serif" font-size="36" font-weight="bold" 
        text-anchor="middle" fill="white">P</text>
  <text x="90" y="130" font-family="Arial, sans-serif" font-size="12" 
        text-anchor="middle" fill="#64748b">$ProjectName</text>
</svg>
"@

Set-Content -Path (Join-Path $staticPath "images\apple-touch-icon.svg") -Value $appleIconContent -Encoding UTF8

# Tao custom CSS cho branding
Write-Host "Tao custom branding CSS..." -ForegroundColor Yellow
$brandingCssPath = "$frontendPath\css\custom-branding.scss"

$brandingCss = @"
// Custom Branding cho $ProjectName
// Generated on: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

// ===== PROJECT BRANDING =====
:root {
  --project-name: '$ProjectName';
  --project-description: '$ProjectDescription';
  --project-tagline: '$Tagline';
  --footer-text: '$FooterText';
  --company-name: '$CompanyName';
}

// ===== HEADER BRANDING =====
.header {
  .navbar-brand {
    .project-name {
      font-size: 1.5rem;
      font-weight: 800;
      background: linear-gradient(135deg, #6366f1, #8b5cf6, #06b6d4);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      margin-left: 10px;
    }
  }
}

// ===== HOMEPAGE BRANDING =====
.hero-section {
  text-align: center;
  padding: 60px 20px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.1));
  border-radius: 16px;
  margin: 40px 0;
  
  .hero-title {
    font-size: 3rem;
    font-weight: 900;
    background: linear-gradient(135deg, #6366f1, #8b5cf6, #06b6d4);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin-bottom: 16px;
  }
  
  .hero-subtitle {
    font-size: 1.25rem;
    color: #64748b;
    font-weight: 500;
    margin-bottom: 8px;
  }
  
  .hero-tagline {
    font-size: 1rem;
    color: #94a3b8;
    font-weight: 400;
  }
}

// ===== FOOTER BRANDING =====
.footer {
  background: linear-gradient(135deg, #0f172a, #1e293b);
  color: #cbd5e1;
  padding: 40px 20px;
  text-align: center;
  border-top: 3px solid #6366f1;
  
  .footer-content {
    max-width: 1200px;
    margin: 0 auto;
    
    .footer-brand {
      margin-bottom: 20px;
      
      .footer-logo {
        font-size: 1.5rem;
        font-weight: 800;
        background: linear-gradient(135deg, #6366f1, #8b5cf6, #06b6d4);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        margin-bottom: 8px;
      }
      
      .footer-description {
        color: #94a3b8;
        font-size: 0.9rem;
      }
    }
    
    .footer-text {
      color: #64748b;
      font-size: 0.85rem;
      margin-top: 20px;
      padding-top: 20px;
      border-top: 1px solid #334155;
    }
  }
}

// ===== PAGE TITLES =====
.page-title {
  font-size: 2.5rem;
  font-weight: 800;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 16px;
  
  &::before {
    content: var(--project-name) ' ';
    font-size: 0.6em;
    opacity: 0.7;
  }
}

// ===== BREADCRUMBS =====
.breadcrumb {
  background: rgba(99, 102, 241, 0.05);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 24px;
  
  .breadcrumb-item {
    color: #6366f1;
    font-weight: 500;
    
    &.active {
      color: #64748b;
    }
    
    a {
      color: #6366f1;
      text-decoration: none;
      
      &:hover {
        color: #8b5cf6;
        text-decoration: underline;
      }
    }
  }
}

// ===== CUSTOM BADGES =====
.badge-project {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  font-weight: 600;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  box-shadow: 0 2px 4px rgba(99, 102, 241, 0.3);
}

// ===== LOADING STATES =====
.loading-spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 3px solid rgba(99, 102, 241, 0.3);
  border-radius: 50%;
  border-top-color: #6366f1;
  animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// ===== NOTIFICATION STYLES =====
.notification {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(139, 92, 246, 0.1));
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 12px;
  padding: 16px;
  margin: 16px 0;
  
  .notification-icon {
    color: #6366f1;
    margin-right: 12px;
  }
  
  .notification-content {
    color: #0f172a;
    font-weight: 500;
  }
}

// ===== RESPONSIVE BRANDING =====
@media (max-width: 768px) {
  .hero-section {
    padding: 40px 15px;
    
    .hero-title {
      font-size: 2rem;
    }
    
    .hero-subtitle {
      font-size: 1.1rem;
    }
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .footer {
    padding: 30px 15px;
  }
}

// ===== DARK MODE BRANDING =====
@media (prefers-color-scheme: dark) {
  .hero-section {
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.2));
  }
  
  .notification {
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.2));
    border-color: rgba(99, 102, 241, 0.3);
    
    .notification-content {
      color: #f8fafc;
    }
  }
}
"@

Set-Content -Path $brandingCssPath -Value $brandingCss -Encoding UTF8
Write-Host "Tao custom branding CSS thanh cong" -ForegroundColor Green

# Cap nhat app.scss de import custom branding
Write-Host "Cap nhat app.scss..." -ForegroundColor Yellow
$appScssPath = "$frontendPath\css\app.scss"
if (Test-Path $appScssPath) {
    $appContent = Get-Content $appScssPath -Raw
    if ($appContent -notmatch "custom-branding") {
        $newContent = $appContent + "`n@import 'custom-branding';"
        Set-Content -Path $appScssPath -Value $newContent -Encoding UTF8
        Write-Host "Them custom branding vao app.scss" -ForegroundColor Green
    } else {
        Write-Host "Custom branding da co trong app.scss" -ForegroundColor Blue
    }
}

# Tao meta tags cho SEO
Write-Host "Tao meta tags SEO..." -ForegroundColor Yellow
$metaTags = @"
<!-- $ProjectName Meta Tags -->
<meta name="application-name" content="$ProjectName">
<meta name="description" content="$ProjectDescription - $Tagline">
<meta name="keywords" content="$ProjectName, blockchain, explorer, crypto, $CompanyName">
<meta name="author" content="$CompanyName">
<meta name="robots" content="index, follow">

<!-- Open Graph / Facebook -->
<meta property="og:type" content="website">
<meta property="og:url" content="https://explorer.poatc.com/">
<meta property="og:title" content="$ProjectName - $ProjectDescription">
<meta property="og:description" content="$ProjectDescription - $Tagline">
<meta property="og:image" content="/images/logo.svg">

<!-- Twitter -->
<meta property="twitter:card" content="summary_large_image">
<meta property="twitter:url" content="https://explorer.poatc.com/">
<meta property="twitter:title" content="$ProjectName - $ProjectDescription">
<meta property="twitter:description" content="$ProjectDescription - $Tagline">
<meta property="twitter:image" content="/images/logo.svg">

<!-- Theme Color -->
<meta name="theme-color" content="#6366f1">
<meta name="msapplication-TileColor" content="#6366f1">
"@

$metaPath = "$staticPath\meta-tags.html"
Set-Content -Path $metaPath -Value $metaTags -Encoding UTF8
Write-Host "Tao meta tags SEO thanh cong" -ForegroundColor Green

Write-Host ""
Write-Host "HOAN THANH TUY CHINH BRANDING!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Ten du an: $ProjectName" -ForegroundColor Cyan
Write-Host "Mo ta: $ProjectDescription" -ForegroundColor Cyan
Write-Host "Tagline: $Tagline" -ForegroundColor Cyan
Write-Host "Footer: $FooterText" -ForegroundColor Cyan
Write-Host "Cong ty: $CompanyName" -ForegroundColor Cyan
Write-Host ""
Write-Host "Da tao:" -ForegroundColor Yellow
Write-Host "  • Favicon moi" -ForegroundColor White
Write-Host "  • Apple touch icon" -ForegroundColor White
Write-Host "  • Custom branding CSS" -ForegroundColor White
Write-Host "  • Meta tags SEO" -ForegroundColor White
Write-Host "  • Hero section styling" -ForegroundColor White
Write-Host "  • Footer branding" -ForegroundColor White
Write-Host ""
Write-Host "Cac buoc tiep theo:" -ForegroundColor Yellow
Write-Host "1. cd docker-compose" -ForegroundColor White
Write-Host "2. docker-compose restart frontend" -ForegroundColor White
Write-Host "3. Kiem tra: http://localhost:80" -ForegroundColor White
