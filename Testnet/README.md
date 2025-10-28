# 🚀 POATC Testnet - Quick Start Guide

## 📋 Tổng Quan

POATC (Proof-of-Authority with Tracing and Consensus) là blockchain Layer 1 với 6 tính năng nâng cao:
1. 🛡️ Anomaly Detection System
2. ⭐ Reputation System  
3. 🎯 Smart Validator Selection
4. 🔍 Transaction Tracing
5. ⏰ Dynamic Time Management
6. 🛡️ Access Control (Whitelist/Blacklist)

## 🎯 Khởi Động Nhanh

### Yêu Cầu
- Go 1.19+
- Python 3.x
- 2 terminal windows

### Các Lệnh Trong `command.txt`

Tất cả lệnh khởi động đã được ghi sẵn trong file `command.txt`. Copy và chạy từng phần:

#### **Terminal 1: Node 1 (Validator chính)**
```bash
cd testnet
../hdchain.exe --datadir node1 --http --http.port 8545 --http.api "admin,eth,net,web3,personal,miner,poatc" --http.corsdomain "*" --port 30303 --networkid 1369 --mine --miner.etherbase 0x3003d6498603fAD5F232452B21c8B6EB798d20f1 --unlock 0x3003d6498603fAD5F232452B21c8B6EB798d20f1 --password password.txt --allow-insecure-unlock --nodiscover console
```

#### **Terminal 2: Node 2 (Validator phụ)**
```bash
cd testnet
../hdchain.exe --datadir node2 --http --http.port 8549 --http.api "admin,eth,net,web3,personal,miner,poatc" --http.corsdomain "*" --port 30304 --networkid 1369 --mine --miner.etherbase 0xE22bb120826219E8ec00d3af3d16EFE7cADe7B08 --unlock 0xE22bb120826219E8ec00d3af3d16EFE7cADe7B08 --password password.txt --allow-insecure-unlock console
```

#### **Kết Nối 2 Nodes**
Trong console Node 1:
```javascript
// Lấy enode của Node 2
admin.nodeInfo.enode

// Trong console Node 2, connect về Node 1:
admin.addPeer("enode://...@127.0.0.1:30303")

// Kiểm tra kết nối
admin.peers
```

## 🌐 Explorer Dashboard

### Khởi động Explorer
```bash
cd testnet/explorer
python serve.py 8080
```

Truy cập: http://localhost:8080

### Tính Năng Explorer
- ✅ Real-time blockchain data từ RPC
- ✅ Xem blocks và transactions
- ✅ MetaMask wallet connection
- ✅ Faucet để lấy test tokens
- ✅ Monitor POATC features
- ✅ Smart contract interaction
- ✅ No backend/database required

## 🔗 Smart Contracts

### Deploy Contract

```bash
cd testnet/contracts
npm install

# Deploy
node deploy.js

# Test contract
node test_contract.js
```

### Contract đã deploy
- **DataTraceability**: `0x586b3b0c8f79a72c2AE7a25eeD1B56e2b0a2671B`
- Có thể interact qua Explorer UI

## 🧪 Testing

### Test Gửi Giao Dịch
Trong console Node 1:
```javascript
eth.sendTransaction({
  from: eth.coinbase,
  to: "0xE22bb120826219E8ec00d3af3d16EFE7cADe7B08",
  value: web3.toWei(1, "ether")
})
```

### Test POATC APIs
```javascript
// Anomaly stats
poatc.getAnomalyStats()

// Reputation  
poatc.getReputationStats()
poatc.getReputation("0x3003d6498603fAD5F232452B21c8B6EB798d20f1")

// Validator selection
poatc.getValidatorSelectionStats()

// Tracing
poatc.getTracingStats()

// Time dynamic
poatc.getTimeDynamicStats()

// Whitelist/Blacklist
poatc.getWhitelistBlacklistStats()
```

### TPS Testing (Manual)

Tạo nhiều transactions để test throughput:

```javascript
// Trong console Node 1
for (var i = 0; i < 50; i++) {
  eth.sendTransaction({
    from: eth.coinbase,
    to: "0xE22bb120826219E8ec00d3af3d16EFE7cADe7B08", 
    value: web3.toWei(0.01, "ether")
  })
}
```

Monitor TPS trong Explorer dashboard.

## 📊 Blockscout (Tùy chọn)

Nếu muốn dùng Blockscout như blockchain explorer chuyên nghiệp:

### Yêu Cầu
- Docker Desktop
- Docker Compose 2.x+

### Khởi động
```bash
cd testnet/blockscout
docker-compose up -d
```

Truy cập: http://localhost:4000

**Lưu ý**: Blockscout nặng và phức tạp. Explorer built-in đã đủ cho demo.

## 🔧 Troubleshooting

### Nodes không kết nối
```bash
# Kiểm tra firewall
# Kiểm tra port đã dùng chưa
netstat -ano | findstr "30303"
netstat -ano | findstr "30304"

# Restart nodes với nodiscover nếu test local
--nodiscover
```

### Explorer không load data
1. Kiểm tra nodes có chạy không
2. Test RPC:
```bash
curl -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://127.0.0.1:8545
```

### CORS errors
Đảm bảo nodes chạy với:
```
--http.corsdomain "*"
```

## 📁 Cấu Trúc Thư Mục

```
testnet/
├── node1/                 # Data node 1
├── node2/                 # Data node 2
├── explorer/              # Web explorer (RPC only)
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── serve.py
├── contracts/             # Smart contracts
│   ├── contracts/
│   ├── deploy.js
│   └── test_contract.js
├── genesis.json           # Genesis config
├── whitelist_blacklist.json
├── command.txt            # Tất cả lệnh cần thiết
└── README.md             # File này
```

## 🎯 Demo Flow

1. **Start nodes** (2 terminals)
2. **Connect nodes** (admin.addPeer)
3. **Start explorer** (python serve.py 8080)
4. **Show features**:
   - Overview tab: Network status, blocks
   - POATC Features: 6 advanced systems
   - Click blocks/txs: Etherscan-like details
5. **Test transactions**: Send via console or Explorer
6. **Show POATC APIs**: Test trong console
7. **Deploy contract**: node deploy.js
8. **Interact with contract**: Via Explorer UI

## 📚 Tài Liệu Thêm

- **POATC Layer 1 Docs**: Xem `explorer/docs.html` (mở trong browser)
- **Technical Docs**: `TECHNICAL_DOCUMENTATION.md` (root folder)
- **Algorithm Analysis**: `PHAN_TICH_DU_AN_THUAT_TOAN.md`

## 🔑 Account Thông Tin

### Node 1
- Address: `0x3003d6498603fAD5F232452B21c8B6EB798d20f1`
- Password: trong `password.txt`

### Node 2
- Address: `0xE22bb120826219E8ec00d3af3d16EFE7cADe7B08`
- Password: trong `password.txt`

## 🎉 Quick Commands

```bash
# Xem balance
eth.getBalance(eth.coinbase)

# Unlock account (nếu cần)
personal.unlockAccount(eth.coinbase, "123", 0)

# Mining status
eth.mining

# Current block
eth.blockNumber

# Peer count
admin.peers.length
```

---

**Happy Testing! 🚀**

Mọi lệnh chi tiết đều có trong `command.txt`. Explorer không cần backend/database, chỉ cần RPC connection!

