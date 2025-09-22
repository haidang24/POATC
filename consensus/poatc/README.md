# Enhanced Clique Consensus Engine

## Tổng Quan

Đây là phiên bản nâng cao của Clique Proof-of-Authority consensus engine với nhiều tính năng mở rộng để tăng cường bảo mật, minh bạch và hiệu suất.

## Cấu Trúc Thư Mục

### Core Files (Các file cốt lõi)
- `clique.go` - Engine chính của Clique consensus
- `snapshot.go` - Quản lý snapshot và trạng thái validator
- `api.go` - RPC API endpoints cho tất cả các tính năng

### Enhanced Systems (Các hệ thống mở rộng)
- `01_anomaly_detection.go` - Hệ thống phát hiện bất thường
- `02_whitelist_blacklist.go` - Quản lý danh sách trắng/đen
- `03_validator_selection.go` - Hệ thống chọn validator 2-tier
- `04_reputation_system.go` - Hệ thống tính điểm reputation
- `05_tracing_system.go` - Hệ thống tracing với Merkle Tree

### Test Files (Các file test)
- `clique_test.go` - Test cho engine chính
- `snapshot_test.go` - Test cho snapshot management
- `random_poa_test.go` - Test cho random POA algorithm
- `01_anomaly_detection_test.go` - Test cho anomaly detection
- `02_whitelist_blacklist_test.go` - Test cho whitelist/blacklist
- `03_validator_selection_test.go` - Test cho validator selection
- `04_reputation_system_test.go` - Test cho reputation system

### Documentation (Tài liệu)
- `RANDOM_POA_README.md` - Tài liệu về Random POA algorithm
- `FAIRNESS_MECHANISMS_README.md` - Tài liệu về cơ chế công bằng
- `01_ANOMALY_DETECTION_README.md` - Tài liệu về anomaly detection
- `02_WHITELIST_BLACKLIST_README.md` - Tài liệu về whitelist/blacklist
- `03_VALIDATOR_SELECTION_README.md` - Tài liệu về validator selection
- `04_REPUTATION_SYSTEM_README.md` - Tài liệu về reputation system
- `05_TRACING_SYSTEM_README.md` - Tài liệu về tracing system

## Tính Năng Chính

### 1. **Random POA Algorithm**
- Thay thế round-robin bằng random selection
- Deterministic randomness sử dụng block hash và number
- Công bằng hơn trong việc chọn validator

### 2. **Anomaly Detection System**
- Phát hiện rapid signing, suspicious patterns
- Phát hiện timestamp drift và missing signers
- Tích hợp với reputation system

### 3. **Whitelist/Blacklist Management**
- Quản lý danh sách validator được phép/cấm
- Persistence với JSON file
- Tích hợp với reputation system

### 4. **2-Tier Validator Selection**
- Tier 1: Chọn tập validator nhỏ
- Tier 2: Random selection từ tập nhỏ
- Nhiều phương pháp selection (Random, Stake-based, Reputation-based, Hybrid)

### 5. **On-Chain Reputation System**
- Multi-factor scoring (Block Mining, Uptime, Consistency, Penalty)
- Decay factor và violation tracking
- Fairness mechanisms (score capping, partial reset, new validator boost)

### 6. **Tracing System với Merkle Tree**
- Trace tất cả hành vi validator
- Merkle Tree để tạo immutable audit trail
- Event verification và Merkle proof
- Export/import trace data

## Cách Sử Dụng

### Build Executable
```bash
go build -o hdchain.exe ./cmd/geth
```

### Chạy Node
```bash
./hdchain.exe --datadir ./Testnet/node1 --port 30303 --rpc --rpcport 8549 --rpcaddr 0.0.0.0 --rpcapi "eth,net,web3,personal,clique" --mine --miner.etherbase 0x6519B747fC2c4DD4393843855Bef77f28875B07C --unlock 0x6519B747fC2c4DD4393843855Bef77f28875B07C --password ./Testnet/node1/password.txt
```

### Test Scripts
```powershell
# Test toàn bộ hệ thống
.\Testnet\test_complete_system.ps1

# Test từng component
.\Testnet\test_anomaly_detection.ps1
.\Testnet\test_whitelist_blacklist.ps1
.\Testnet\test_validator_selection.ps1
.\Testnet\test_reputation_system.ps1
.\Testnet\test_tracing_system.ps1
```

## API Endpoints

### Core Clique APIs
- `clique_getSigners` - Lấy danh sách signers
- `clique_propose` - Đề xuất thêm/xóa signer
- `clique_discard` - Hủy đề xuất

### Enhanced System APIs
- **Anomaly Detection**: `clique_getAnomalyStats`, `clique_detectAnomalies`
- **Whitelist/Blacklist**: `clique_getWhitelist`, `clique_addToWhitelist`, `clique_addToBlacklist`
- **Validator Selection**: `clique_getValidatorSelectionStats`, `clique_getSmallValidatorSet`
- **Reputation System**: `clique_getReputationStats`, `clique_getReputationScore`
- **Tracing System**: `clique_getTracingStats`, `clique_getTraceEvents`, `clique_getMerkleRoot`

## Cấu Hình

Tất cả các hệ thống đều có cấu hình mặc định và có thể được tùy chỉnh thông qua API hoặc code.

## Lợi Ích

1. **Bảo Mật**: Anomaly detection và whitelist/blacklist
2. **Công Bằng**: Random selection và fairness mechanisms
3. **Minh Bạch**: Tracing system với Merkle Tree
4. **Hiệu Suất**: 2-tier validator selection
5. **Tin Cậy**: Reputation system và violation tracking
6. **Kiểm Soát**: Comprehensive API và monitoring

## Tài Liệu Chi Tiết

Xem các file README riêng biệt cho từng hệ thống để biết thêm chi tiết về cách sử dụng và cấu hình.
