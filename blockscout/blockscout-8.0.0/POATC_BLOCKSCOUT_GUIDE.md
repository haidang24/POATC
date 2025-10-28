# 🚀 Hướng Dẫn Cài Đặt và Chạy Blockscout cho POATC Blockchain

## 📋 Tổng Quan

Hướng dẫn này mô tả cách cài đặt và chạy Blockscout Explorer cho blockchain POATC, bao gồm các thay đổi cấu hình cần thiết để tương thích với node POATC.

## 🛠️ Các Thay Đổi Đã Thực Hiện

### 1. **File: `docker-compose/docker-compose.yml`**

#### Thay đổi chính:
```yaml
backend:
  environment:
    ETHEREUM_JSONRPC_HTTP_URL: http://host.docker.internal:8545/
    ETHEREUM_JSONRPC_TRACE_URL: http://host.docker.internal:8545/
    ETHEREUM_JSONRPC_WS_URL: ws://host.docker.internal:8545/
    CHAIN_ID: '1337'                    # ← Thay đổi từ default
    COIN_NAME: 'POATC'                  # ← Thêm mới
    COIN: 'POATC'                       # ← Thêm mới
    DISABLE_MARKET: true                # ← Thêm mới (tắt market features)
    ETHEREUM_JSONRPC_DISABLE_ARCHIVE_BALANCES: true  # ← Thêm mới
    INDEXER_DISABLE_INTERNAL_TRANSACTIONS_FETCHER: true  # ← Thêm mới
  ports:
    - "4000:4000"                       # ← Thêm mới (expose backend port)
```

#### Giải thích các thay đổi:
- **CHAIN_ID**: Đặt thành 1337 (ID của POATC network)
- **COIN_NAME & COIN**: Đặt tên đồng coin là "POATC"
- **DISABLE_MARKET**: Tắt các tính năng market (không cần thiết cho private network)
- **ETHEREUM_JSONRPC_DISABLE_ARCHIVE_BALANCES**: Tắt archive balances (tránh lỗi historical state)
- **INDEXER_DISABLE_INTERNAL_TRANSACTIONS_FETCHER**: Tắt internal transactions fetcher (tránh lỗi "required historical state unavailable")
- **Ports**: Expose port 4000 để có thể truy cập backend API trực tiếp

## 🚀 Cách Chạy Blockscout

### Bước 1: Đảm bảo POATC Node đang chạy
```bash
# Kiểm tra POATC node đang chạy trên port 8545
curl http://localhost:8545
```

### Bước 2: Đảm bảo Docker Desktop đang chạy
- Mở Docker Desktop
- Đợi cho đến khi Docker daemon sẵn sàng

### Bước 3: Di chuyển đến thư mục docker-compose
```bash
cd blockscout/blockscout-8.0.0/docker-compose
```

### Bước 4: Khởi động Blockscout
```bash
# Khởi động tất cả services
docker-compose up -d

# Hoặc với build (nếu cần)
docker-compose up --build -d
```

### Bước 5: Kiểm tra trạng thái
```bash
# Xem trạng thái containers
docker-compose ps

# Xem logs backend
docker-compose logs backend --tail=20

# Xem logs frontend
docker-compose logs frontend --tail=20
```

## 🌐 Truy Cập Blockscout

### URLs:
- **Blockscout Explorer**: http://localhost:80
- **Backend API**: http://localhost:4000/api/v2/
- **Database**: localhost:7432 (blockscout), localhost:7433 (stats)

### Test API:
```bash
# Test backend API
curl "http://localhost:4000/api/v2/blocks"

# Test frontend
curl http://localhost:80
```

## 🔧 Quản Lý Services

### Dừng Blockscout:
```bash
docker-compose down
```

### Dừng và xóa volumes (reset database):
```bash
docker-compose down -v
```

### Restart một service cụ thể:
```bash
# Restart backend
docker-compose restart backend

# Restart frontend
docker-compose restart frontend
```

### Xem logs real-time:
```bash
# Tất cả services
docker-compose logs -f

# Chỉ backend
docker-compose logs -f backend

# Chỉ frontend
docker-compose logs -f frontend
```

## 📊 Services và Ports

| Service | Port | Mô tả |
|---------|------|-------|
| **proxy** | 80 | Nginx proxy (frontend access) |
| **backend** | 4000 | Blockscout API server |
| **frontend** | 3000 (internal) | Next.js frontend |
| **db** | 7432 | PostgreSQL main database |
| **stats-db** | 7433 | PostgreSQL stats database |
| **redis-db** | 6379 (internal) | Redis cache |
| **stats** | - | Statistics service |
| **visualizer** | - | Visualization service |
| **sig-provider** | - | Signature provider |
| **nft_media_handler** | - | NFT media handler |
| **user-ops-indexer** | - | User operations indexer |

## ⚠️ Lưu Ý Quan Trọng

### 1. **POATC Node Requirements**
- POATC node phải đang chạy trên port 8545
- Node phải có RPC endpoints enabled
- Node phải có WebSocket enabled (cho real-time updates)

### 2. **Memory Requirements**
- Tối thiểu 8GB RAM
- Khuyến nghị 16GB RAM cho production

### 3. **Storage Requirements**
- Tối thiểu 20GB free space
- Database sẽ phát triển theo thời gian

### 4. **Network Requirements**
- Port 80: Frontend access
- Port 4000: Backend API access
- Port 7432: Database access (nếu cần)
- Port 8545: POATC node connection

## 🐛 Troubleshooting

### Backend không khởi động:
```bash
# Kiểm tra logs
docker-compose logs backend

# Kiểm tra POATC node
curl http://localhost:8545

# Restart backend
docker-compose restart backend
```

### Frontend không load:
```bash
# Kiểm tra proxy
docker-compose logs proxy

# Kiểm tra frontend
docker-compose logs frontend

# Restart frontend
docker-compose restart frontend
```

### Database errors:
```bash
# Reset database
docker-compose down -v
docker-compose up -d
```

### Port conflicts:
```bash
# Kiểm tra port đang sử dụng
netstat -ano | findstr ":80"
netstat -ano | findstr ":4000"

# Thay đổi port trong docker-compose.yml nếu cần
```

## 📈 Performance Tuning

### Để tối ưu performance:
1. **Tăng memory limits** trong docker-compose.yml
2. **Tăng database pool size** nếu cần
3. **Enable caching** cho Redis
4. **Monitor logs** để phát hiện bottlenecks

### Monitoring:
```bash
# Xem resource usage
docker stats

# Xem database connections
docker-compose exec db psql -U blockscout -d blockscout -c "SELECT * FROM pg_stat_activity;"
```

## 🔄 Updates và Maintenance

### Update Blockscout:
```bash
# Pull latest images
docker-compose pull

# Restart với images mới
docker-compose up -d
```

### Backup Database:
```bash
# Backup main database
docker-compose exec db pg_dump -U blockscout blockscout > blockscout_backup.sql

# Backup stats database
docker-compose exec stats-db pg_dump -U blockscout blockscout > stats_backup.sql
```

### Restore Database:
```bash
# Restore main database
docker-compose exec -T db psql -U blockscout blockscout < blockscout_backup.sql
```

## 📞 Support

Nếu gặp vấn đề:
1. Kiểm tra logs: `docker-compose logs`
2. Kiểm tra POATC node connection
3. Kiểm tra port conflicts
4. Restart services: `docker-compose restart`

## 📝 Changelog

### Version 1.0 (Current)
- ✅ Cấu hình cơ bản cho POATC blockchain
- ✅ Disable internal transactions fetcher
- ✅ Disable market features
- ✅ Expose backend port 4000
- ✅ Tối ưu cho private network

---

**Tác giả**: AI Assistant  
**Ngày tạo**: 2025-01-14  
**Phiên bản**: 1.0  
**Tương thích**: Blockscout 8.0.0, POATC Blockchain
