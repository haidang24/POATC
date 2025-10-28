# 🎯 Blockscout cho POATC Blockchain

> **Blockscout Explorer** đã được cấu hình và tối ưu để hoạt động với **POATC Blockchain**

## 📚 Tài Liệu

| File | Mô tả |
|------|-------|
| 📖 **[POATC_BLOCKSCOUT_GUIDE.md](POATC_BLOCKSCOUT_GUIDE.md)** | Hướng dẫn chi tiết đầy đủ |
| ⚡ **[QUICK_START.md](QUICK_START.md)** | Hướng dẫn chạy nhanh (3 bước) |
| 🔧 **[CONFIG_CHANGES.md](CONFIG_CHANGES.md)** | Tóm tắt các thay đổi cấu hình |
| 🚀 **[start_blockscout.ps1](start_blockscout.ps1)** | Script tự động khởi động |
| 🎨 **[FRONTEND_CUSTOMIZATION_GUIDE.md](FRONTEND_CUSTOMIZATION_GUIDE.md)** | Hướng dẫn tùy chỉnh giao diện |

## 🎨 Tùy Chỉnh Giao Diện

| Script | Mô tả |
|--------|-------|
| 🎨 **[customize_all.ps1](customize_all.ps1)** | Script tổng hợp (khuyến nghị) |
| 🚫 **[disable_ads.ps1](disable_ads.ps1)** | Loại bỏ quảng cáo |
| 🖼️ **[replace_logo.ps1](replace_logo.ps1)** | Thay logo POATC |
| 🎨 **[customize_theme.ps1](customize_theme.ps1)** | Tùy chỉnh màu sắc theme |

## 🚀 Cách Sử Dụng

### Phương pháp 1: Script tự động (Khuyến nghị)
```powershell
.\start_blockscout.ps1
```

### Phương pháp 2: Lệnh thủ công
```bash
cd docker-compose
docker-compose up -d
```

## 🎨 Tùy Chỉnh Giao Diện Nhanh

### Tùy chỉnh hoàn chỉnh (Khuyến nghị)
```powershell
.\customize_all.ps1
```

### Tùy chỉnh với màu sắc tùy chỉnh
```powershell
.\customize_all.ps1 -PrimaryColor "#ff6b35" -SecondaryColor "#2563eb" -BackgroundColor "#ffffff"
```

### Tùy chỉnh với logo tùy chỉnh
```powershell
.\customize_all.ps1 -LogoPath "C:\path\to\your\logo.svg"
```

## 🌐 Truy Cập

- **🔍 Explorer**: http://localhost:80
- **🔧 API**: http://localhost:4000/api/v2/
- **🗄️ Database**: localhost:7432

## ✅ Đã Hoàn Thành

- ✅ Cấu hình cho POATC blockchain (Chain ID: 1337)
- ✅ Tối ưu cho private network
- ✅ Giải quyết lỗi "historical state unavailable"
- ✅ Tắt các fetcher không cần thiết
- ✅ Expose backend API port
- ✅ Script tự động khởi động
- ✅ Tài liệu đầy đủ

## 🎯 Tính Năng

- 📊 **Block Explorer**: Xem blocks, transactions, addresses
- 🔍 **Search**: Tìm kiếm theo hash, address, block number
- 📈 **Statistics**: Thống kê network
- 🎨 **Custom Theme**: Giao diện tùy chỉnh cho POATC
- 🔧 **API**: RESTful API đầy đủ

## ⚠️ Yêu Cầu Hệ Thống

- **Docker Desktop** đang chạy
- **POATC Node** trên port 8545
- **RAM**: Tối thiểu 8GB (Khuyến nghị 16GB)
- **Storage**: Tối thiểu 20GB free space

## 🔧 Quản Lý

```bash
# Xem trạng thái
docker-compose ps

# Xem logs
docker-compose logs -f

# Dừng
docker-compose down

# Reset hoàn toàn
docker-compose down -v
```

## 🆘 Hỗ Trợ

Nếu gặp vấn đề:
1. Kiểm tra POATC node: `curl http://localhost:8545`
2. Kiểm tra Docker: `docker --version`
3. Xem logs: `docker-compose logs`
4. Đọc tài liệu: `POATC_BLOCKSCOUT_GUIDE.md`

---

**🎉 Chúc mừng! Blockscout đã sẵn sàng để khám phá POATC Blockchain!**
