# 🚀 Blockscout Startup Script cho POATC Blockchain
# Tác giả: AI Assistant
# Ngày: 2025-01-14

Write-Host "🚀 Khởi động Blockscout cho POATC Blockchain..." -ForegroundColor Green

# Kiểm tra Docker Desktop
Write-Host "📋 Kiểm tra Docker Desktop..." -ForegroundColor Yellow
try {
    docker --version | Out-Null
    Write-Host "✅ Docker đã sẵn sàng" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker Desktop chưa chạy! Vui lòng khởi động Docker Desktop trước." -ForegroundColor Red
    Read-Host "Nhấn Enter để thoát"
    exit 1
}

# Kiểm tra POATC Node
Write-Host "📋 Kiểm tra POATC Node..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8545" -Method POST -Body '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' -ContentType "application/json" -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ POATC Node đang chạy trên port 8545" -ForegroundColor Green
    } else {
        Write-Host "⚠️ POATC Node phản hồi nhưng có thể có vấn đề" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Không thể kết nối đến POATC Node tại localhost:8545" -ForegroundColor Red
    Write-Host "💡 Vui lòng đảm bảo POATC Node đang chạy trước khi tiếp tục" -ForegroundColor Yellow
    $continue = Read-Host "Bạn có muốn tiếp tục? (y/N)"
    if ($continue -ne "y" -and $continue -ne "Y") {
        exit 1
    }
}

# Di chuyển đến thư mục docker-compose
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
$dockerComposePath = Join-Path $scriptPath "docker-compose"
Set-Location $dockerComposePath

Write-Host "📁 Đang ở thư mục: $dockerComposePath" -ForegroundColor Cyan

# Kiểm tra file docker-compose.yml
if (-not (Test-Path "docker-compose.yml")) {
    Write-Host "❌ Không tìm thấy docker-compose.yml" -ForegroundColor Red
    exit 1
}

# Dừng containers cũ (nếu có)
Write-Host "🛑 Dừng containers cũ..." -ForegroundColor Yellow
try {
    docker-compose down 2>$null
    Write-Host "✅ Đã dừng containers cũ" -ForegroundColor Green
} catch {
    Write-Host "ℹ️ Không có containers cũ để dừng" -ForegroundColor Blue
}

# Khởi động Blockscout
Write-Host "🚀 Khởi động Blockscout..." -ForegroundColor Yellow
try {
    docker-compose up -d
    Write-Host "✅ Blockscout đang khởi động..." -ForegroundColor Green
} catch {
    Write-Host "❌ Lỗi khi khởi động Blockscout" -ForegroundColor Red
    exit 1
}

# Đợi services khởi động
Write-Host "⏳ Đợi services khởi động (30 giây)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# Kiểm tra trạng thái
Write-Host "📊 Kiểm tra trạng thái containers..." -ForegroundColor Yellow
docker-compose ps

# Kiểm tra frontend
Write-Host "🌐 Kiểm tra Frontend..." -ForegroundColor Yellow
try {
    $frontendResponse = Invoke-WebRequest -Uri "http://localhost:80" -TimeoutSec 10
    if ($frontendResponse.StatusCode -eq 200) {
        Write-Host "✅ Frontend đã sẵn sàng tại http://localhost:80" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Frontend chưa sẵn sàng, có thể cần thêm thời gian" -ForegroundColor Yellow
}

# Kiểm tra backend API
Write-Host "🔧 Kiểm tra Backend API..." -ForegroundColor Yellow
try {
    $apiResponse = Invoke-WebRequest -Uri "http://localhost:4000/api/v2/blocks" -TimeoutSec 10
    if ($apiResponse.StatusCode -eq 200) {
        Write-Host "✅ Backend API đã sẵn sàng tại http://localhost:4000/api/v2/" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Backend API chưa sẵn sàng, có thể cần thêm thời gian" -ForegroundColor Yellow
}

# Hiển thị thông tin truy cập
Write-Host ""
Write-Host "🎉 BLOCKSCOUT ĐÃ KHỞI ĐỘNG!" -ForegroundColor Green
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
Write-Host "🌐 Blockscout Explorer: http://localhost:80" -ForegroundColor Cyan
Write-Host "🔧 Backend API: http://localhost:4000/api/v2/" -ForegroundColor Cyan
Write-Host "🗄️ Database: localhost:7432 (blockscout), localhost:7433 (stats)" -ForegroundColor Cyan
Write-Host ""
Write-Host "📋 Các lệnh hữu ích:" -ForegroundColor Yellow
Write-Host "   Xem logs: docker-compose logs -f" -ForegroundColor White
Write-Host "   Dừng: docker-compose down" -ForegroundColor White
Write-Host "   Reset: docker-compose down -v" -ForegroundColor White
Write-Host ""
Write-Host "📖 Tài liệu: POATC_BLOCKSCOUT_GUIDE.md" -ForegroundColor Yellow
Write-Host "⚡ Quick Start: QUICK_START.md" -ForegroundColor Yellow
Write-Host "🔧 Config Changes: CONFIG_CHANGES.md" -ForegroundColor Yellow

# Hỏi người dùng có muốn xem logs không
Write-Host ""
$showLogs = Read-Host "Bạn có muốn xem logs real-time? (y/N)"
if ($showLogs -eq "y" -or $showLogs -eq "Y") {
    Write-Host "📋 Hiển thị logs (Ctrl+C để dừng)..." -ForegroundColor Yellow
    docker-compose logs -f
}
