# ⚡ Quick Start - Blockscout cho POATC

## 🚀 Chạy Nhanh (3 bước)

### 1. Đảm bảo POATC Node đang chạy
```bash
curl http://localhost:8545
```

### 2. Khởi động Blockscout
```bash
cd blockscout/blockscout-8.0.0/docker-compose
docker-compose up -d
```

### 3. Truy cập
- **Explorer**: http://localhost:80
- **API**: http://localhost:4000/api/v2/

## 🔧 Quản lý nhanh

```bash
# Xem trạng thái
docker-compose ps

# Xem logs
docker-compose logs -f

# Dừng
docker-compose down

# Reset hoàn toàn
docker-compose down -v
docker-compose up -d
```

## ⚠️ Lưu ý
- Cần Docker Desktop đang chạy
- Cần POATC node trên port 8545
- Cần ít nhất 8GB RAM

---
📖 **Chi tiết**: Xem `POATC_BLOCKSCOUT_GUIDE.md`
