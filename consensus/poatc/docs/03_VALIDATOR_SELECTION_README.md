# Validator Selection System

## Tổng quan

Validator Selection System là một cơ chế xác thực 2 tầng được tích hợp vào POA consensus engine, cung cấp tính bảo mật và hiệu quả cao hơn so với việc chọn validator ngẫu nhiên trực tiếp từ tất cả validators.

## Cơ chế hoạt động

### Tầng 1: Chọn Small Validator Set
- Từ tất cả validators có sẵn, hệ thống chọn ra một tập validator nhỏ (mặc định 3 validators)
- Việc chọn có thể dựa trên:
  - **Random**: Chọn ngẫu nhiên
  - **Stake**: Dựa trên stake của validator
  - **Reputation**: Dựa trên danh tiếng của validator
  - **Hybrid**: Kết hợp stake, reputation và random

### Tầng 2: Random Selection từ Small Set
- Từ tập validator nhỏ đã chọn, hệ thống random chọn 1 validator để ký block
- Sử dụng block hash và block number làm seed để đảm bảo tính deterministic

## Cấu hình

```go
type ValidatorSelectionConfig struct {
    EnableValidatorSelection bool          // Bật/tắt validator selection
    SmallValidatorSetSize    int           // Kích thước tập validator nhỏ
    SelectionWindow          time.Duration // Thời gian giữ nguyên tập validator
    SelectionMethod          string        // Phương pháp chọn: "random", "stake", "reputation", "hybrid"
    StakeWeight              float64       // Trọng số stake (0.0-1.0)
    ReputationWeight         float64       // Trọng số reputation (0.0-1.0)
    RandomWeight             float64       // Trọng số random (0.0-1.0)
}
```

### Cấu hình mặc định
- `EnableValidatorSelection`: true
- `SmallValidatorSetSize`: 3
- `SelectionWindow`: 1 giờ
- `SelectionMethod`: "hybrid"
- `StakeWeight`: 0.4 (40%)
- `ReputationWeight`: 0.3 (30%)
- `RandomWeight`: 0.3 (30%)

## API Endpoints

### 1. Lấy thống kê validator selection
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getValidatorSelectionStats","params":[],"id":1}' \
  http://localhost:8547
```

### 2. Lấy tập validator nhỏ hiện tại
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getSmallValidatorSet","params":[],"id":1}' \
  http://localhost:8547
```

### 3. Lấy thông tin validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getValidatorInfo","params":["0x1234..."],"id":1}' \
  http://localhost:8547
```

### 4. Thêm validator mới
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_addValidator","params":["0x1234...", "1000000", 1.5],"id":1}' \
  http://localhost:8547
```

### 5. Cập nhật stake của validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_updateValidatorStake","params":["0x1234...", "5000000"],"id":1}' \
  http://localhost:8547
```

### 6. Cập nhật reputation của validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_updateValidatorReputation","params":["0x1234...", 2.0],"id":1}' \
  http://localhost:8547
```

### 7. Lấy lịch sử chọn validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getSelectionHistory","params":[],"id":1}' \
  http://localhost:8547
```

### 8. Ép buộc chọn validator mới
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_forceValidatorSelection","params":[123, "0xabcd..."],"id":1}' \
  http://localhost:8547
```

## Cách sử dụng

### 1. Khởi động nodes
```powershell
.\start_nodes.ps1
```

### 2. Test validator selection
```powershell
.\test_validator_selection.ps1
```

### 3. Test toàn bộ hệ thống
```powershell
.\quick_test.ps1
```

## Lợi ích

### 1. Bảo mật cao hơn
- Giảm khả năng tấn công bằng cách giới hạn số lượng validator có thể ký block
- Tăng tính ngẫu nhiên trong việc chọn validator

### 2. Hiệu quả cao hơn
- Giảm thời gian xử lý bằng cách chỉ xem xét một tập validator nhỏ
- Tối ưu hóa việc quản lý validator

### 3. Linh hoạt
- Hỗ trợ nhiều phương pháp chọn validator
- Có thể điều chỉnh trọng số theo nhu cầu
- Có thể bật/tắt tính năng

### 4. Minh bạch
- Lưu trữ lịch sử chọn validator
- Cung cấp thống kê chi tiết
- API đầy đủ để quản lý

## Ví dụ sử dụng

### Thêm validator với stake cao
```bash
# Thêm validator với stake 10M wei và reputation 2.0
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_addValidator","params":["0x1111...", "10000000", 2.0],"id":1}' \
  http://localhost:8547
```

### Kiểm tra tập validator nhỏ
```bash
# Lấy tập validator nhỏ hiện tại
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getSmallValidatorSet","params":[],"id":1}' \
  http://localhost:8547
```

### Theo dõi lịch sử chọn
```bash
# Lấy lịch sử 10 lần chọn gần nhất
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getSelectionHistory","params":[],"id":1}' \
  http://localhost:8547
```

## Tích hợp với các tính năng khác

Validator Selection System được tích hợp hoàn toàn với:
- **Random POA Algorithm**: Sử dụng cơ chế random cải tiến
- **Anomaly Detection**: Phát hiện bất thường trong việc chọn validator
- **Whitelist/Blacklist**: Quản lý validator được phép/cấm

## Troubleshooting

### Lỗi "validator selection manager not initialized"
- Đảm bảo nodes đã khởi động hoàn toàn
- Kiểm tra genesis.json có đúng signers

### Lỗi "no active validators available"
- Kiểm tra có validators nào được thêm vào hệ thống
- Đảm bảo validators có trạng thái active

### Validator không được chọn
- Kiểm tra validator có trong whitelist không
- Kiểm tra validator có bị blacklist không
- Kiểm tra stake và reputation của validator

## Kết luận

Validator Selection System cung cấp một cơ chế xác thực 2 tầng mạnh mẽ và linh hoạt, giúp tăng cường bảo mật và hiệu quả của POA consensus engine. Hệ thống dễ sử dụng, có API đầy đủ và tích hợp tốt với các tính năng khác.
