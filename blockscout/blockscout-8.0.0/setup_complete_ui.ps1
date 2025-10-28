# Script Setup Giao Dien Hoan Chinh cho Blockscout
# Tac gia: AI Assistant
# Ngay: 2025-01-14

param(
    [string]$ProjectName = "POATC",
    [string]$ProjectDescription = "Blockchain Explorer",
    [string]$Tagline = "Secure, Fast, Decentralized",
    [string]$PrimaryColor = "#6366f1",
    [string]$SecondaryColor = "#8b5cf6",
    [string]$AccentColor = "#06b6d4"
)

Write-Host "🚀 SETUP GIAO DIEN HOAN CHINH CHO $ProjectName" -ForegroundColor Magenta
Write-Host "==========================================" -ForegroundColor Magenta

# 1. Setup giao diện hiện đại
Write-Host "`n📱 Bước 1: Setup giao diện hiện đại..." -ForegroundColor Yellow
& ".\modern_ui.ps1" -ProjectName $ProjectName -ProjectDescription $ProjectDescription -PrimaryColor $PrimaryColor -SecondaryColor $SecondaryColor -AccentColor $AccentColor

# 2. Setup branding
Write-Host "`n🎨 Bước 2: Setup branding..." -ForegroundColor Yellow
& ".\customize_branding.ps1" -ProjectName $ProjectName -ProjectDescription $ProjectDescription -Tagline $Tagline

# 3. Restart frontend
Write-Host "`n🔄 Bước 3: Restart frontend..." -ForegroundColor Yellow
Set-Location "docker-compose"
docker-compose restart frontend

# 4. Đợi và test
Write-Host "`n⏳ Đợi frontend khởi động..." -ForegroundColor Yellow
Start-Sleep 15

Write-Host "`n✅ Test kết quả..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:80" -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Frontend hoạt động bình thường!" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Frontend chưa sẵn sàng, vui lòng đợi thêm..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 HOÀN THÀNH SETUP GIAO DIỆN!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host "📱 Giao diện hiện đại với:" -ForegroundColor Cyan
Write-Host "  • Glass effects & backdrop blur" -ForegroundColor White
Write-Host "  • Gradient backgrounds đẹp mắt" -ForegroundColor White
Write-Host "  • Smooth animations & transitions" -ForegroundColor White
Write-Host "  • Modern cards & buttons" -ForegroundColor White
Write-Host "  • Custom scrollbar" -ForegroundColor White
Write-Host "  • Responsive design" -ForegroundColor White
Write-Host "  • Dark mode support" -ForegroundColor White
Write-Host ""
Write-Host "🎨 Branding đã được tùy chỉnh:" -ForegroundColor Cyan
Write-Host "  • Logo $ProjectName hiện đại" -ForegroundColor White
Write-Host "  • Favicon & Apple touch icon" -ForegroundColor White
Write-Host "  • Meta tags SEO" -ForegroundColor White
Write-Host "  • Hero section styling" -ForegroundColor White
Write-Host "  • Footer branding" -ForegroundColor White
Write-Host "  • Không còn quảng cáo" -ForegroundColor White
Write-Host ""
Write-Host "🌐 Truy cập Blockscout:" -ForegroundColor Yellow
Write-Host "  🔍 Explorer: http://localhost:80" -ForegroundColor White
Write-Host "  🔧 API: http://localhost:4000/api/v2/" -ForegroundColor White
Write-Host ""
Write-Host "💡 Tips:" -ForegroundColor Yellow
Write-Host "  • Refresh trang để thấy thay đổi" -ForegroundColor White
Write-Host "  • Sử dụng Ctrl+F5 để hard refresh" -ForegroundColor White
Write-Host "  • Kiểm tra trên mobile để test responsive" -ForegroundColor White
