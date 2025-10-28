# 🚀 POATC Dashboard - Professional Blockchain Explorer

## ✨ Tính năng chính

### 📊 **Overview Tab**
- **Network Status**: Block hiện tại, coinbase, peers, block time
- **Authorized Signers**: Danh sách validators với reputation lookup
- **Recent Blocks**: 10 blocks gần nhất với chi tiết đầy đủ
- **Block Production Chart**: Biểu đồ thời gian tạo block real-time

### 🔒 **POATC Features Tab**
- **Anomaly Detection**: Phát hiện bất thường trong mạng
- **Validator Selection**: Thống kê lựa chọn validator
- **Reputation System**: Hệ thống danh tiếng validator
- **Transaction Tracing**: Theo dõi giao dịch chi tiết
- **Time Dynamics**: Điều chỉnh thời gian động
- **Access Control**: Whitelist/Blacklist management

### 🌐 **Network Tab**
- **Network Topology**: Visualization kết nối nodes
- **Connection Info**: Thông tin chi tiết kết nối

### 💸 **Transactions Tab**
- **Send Transaction**: Gửi giao dịch với gas settings
- **Recent Transactions**: Lịch sử giao dịch gần đây

## 🎯 **Tính năng như Etherscan**

### 📦 **Block Explorer**
- Click vào bất kỳ block nào để xem chi tiết:
  - Block hash, parent hash, timestamp
  - Miner information, difficulty
  - Gas used/limit, utilization percentage
  - Block size, transaction count
  - Danh sách transactions trong block
  - Extra data và metadata

### 💰 **Transaction Details**
- Click vào transaction để xem:
  - Transaction hash, status (success/failed)
  - From/To addresses
  - Value transferred, gas used/price
  - Block number, transaction index
  - Input data (cho smart contracts)
  - Receipt details

### 👥 **Validator Information**
- Click vào signer để xem reputation
- Real-time validator statistics
- Performance metrics

## 🚀 **Cách sử dụng**

### **Khởi động Explorer**
```bash
# Trong thư mục testnet/explorer/
python serve.py 8080

# Hoặc chạy batch file
start_dashboard.bat
```

Sau đó mở: http://localhost:8080

**Lưu ý**: Explorer chỉ cần RPC connection đến nodes, không cần backend hay database. Tất cả dữ liệu được load trực tiếp từ blockchain qua RPC.

## ⚙️ **Cấu hình**

### **RPC Endpoints**
- **Node1**: http://127.0.0.1:8545
- **Node2**: http://127.0.0.1:8549

### **Chuyển đổi Node**
- Click nút "Node1" hoặc "Node2" để chuyển endpoint
- Hoặc nhập RPC URL tùy chỉnh và click "Connect"

### **Auto-refresh**
- Toggle switch để bật/tắt tự động refresh (3 giây)
- Refresh manual bằng các nút 🔄

## 🎨 **Giao diện**

### **Professional Design**
- Dark theme với gradient accents
- Animated logos và status indicators
- Hover effects và transitions
- Responsive design cho mobile

### **Real-time Updates**
- Block production monitoring
- TPS (Transactions Per Second) calculation
- Network latency measurement
- Connection status indicators

### **Interactive Elements**
- Clickable blocks và transactions
- Modal popups với chi tiết đầy đủ
- Toast notifications cho actions
- Network topology visualization

## 🔧 **Troubleshooting**

### **Dashboard hiển thị OFFLINE**
1. Kiểm tra nodes có chạy không:
   ```bash
   # Test Node1
   curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' http://127.0.0.1:8545
   
   # Test Node2  
   curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' http://127.0.0.1:8549
   ```

2. Restart nodes với CORS settings:
   ```bash
   # Trong testnet/command.txt có lệnh đầy đủ
   ```

### **CORS Errors**
- Đảm bảo nodes chạy với `--http.corsdomain "*"`
- Sử dụng HTTP server thay vì mở file trực tiếp

### **Buttons không hoạt động**
- Kiểm tra JavaScript console cho errors
- Đảm bảo RPC endpoint đúng
- Verify network connectivity

## 📈 **Demo Features**

### **Hackathon Ready**
- Professional UI/UX
- Real-time data visualization
- Interactive block explorer
- Complete transaction tracking
- POATC algorithm monitoring
- Network health dashboard

### **Performance Metrics**
- Block time tracking
- TPS calculation
- Gas utilization monitoring
- Validator performance stats
- Anomaly detection alerts

## 🎯 **Sử dụng cho Demo**

1. **Start nodes**: Chạy cả Node1 và Node2
2. **Start dashboard**: `python serve.py 8080`
3. **Open browser**: http://localhost:8080
4. **Demo flow**:
   - Overview: Hiển thị network status
   - Click blocks: Xem chi tiết như Etherscan
   - POATC tab: Show advanced features
   - Send transaction: Demo giao dịch
   - Network tab: Visualize topology

Dashboard này cung cấp trải nghiệm tương tự Etherscan với các tính năng POATC độc đáo, hoàn hảo cho hackathon demo! 🏆
