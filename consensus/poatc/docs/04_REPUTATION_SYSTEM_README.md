# Reputation System

## Tổng quan

Reputation System là một hệ thống tính điểm danh tiếng on-chain được tích hợp vào POA consensus engine, cung cấp đánh giá minh bạch và công bằng về hiệu suất của các validators dựa trên hoạt động thực tế của họ trên blockchain.

## Tính năng chính

### 🎯 **Scoring Components (Thành phần tính điểm)**

1. **Block Mining Score (40%)**
   - Điểm thưởng cho việc ký block thành công
   - Theo dõi số lượng block đã ký
   - Đánh giá hiệu suất mining

2. **Uptime Score (30%)**
   - Điểm thưởng cho thời gian hoạt động
   - Theo dõi thời gian online của validator
   - Đánh giá độ tin cậy

3. **Consistency Score (20%)**
   - Điểm thưởng cho tính nhất quán
   - Đánh giá khoảng thời gian giữa các block
   - Phân tích độ ổn định

4. **Penalty Score (10%)**
   - Điểm phạt cho các vi phạm
   - Theo dõi số lần vi phạm
   - Áp dụng penalty khi vượt ngưỡng

### 📊 **Scoring Algorithm (Thuật toán tính điểm)**

```go
Total Score = (BlockMiningWeight × BlockMiningScore) +
              (UptimeWeight × UptimeScore) +
              (ConsistencyWeight × ConsistencyScore) -
              (PenaltyWeight × PenaltyScore)
```

### ⚙️ **Configuration (Cấu hình)**

```go
type ReputationConfig struct {
    EnableReputationSystem bool    // Bật/tắt hệ thống reputation
    InitialReputation      float64 // Điểm khởi tạo (1.0)
    MaxReputation          float64 // Điểm tối đa (10.0)
    MinReputation          float64 // Điểm tối thiểu (0.1)
    
    // Trọng số tính điểm
    BlockMiningWeight      float64 // 40%
    UptimeWeight          float64 // 30%
    ConsistencyWeight     float64 // 20%
    PenaltyWeight         float64 // 10%
    
    // Tham số thưởng/phạt
    BlockMiningReward     float64 // 0.1 điểm/block
    UptimeReward          float64 // 0.05 điểm/giờ
    ConsistencyReward     float64 // 0.08 điểm
    PenaltyAmount         float64 // 0.5 điểm phạt
    
    // Thời gian
    EvaluationWindow      time.Duration // 24 giờ
    UpdateInterval        time.Duration // 1 giờ
    DecayFactor           float64       // 0.99 (1% decay)
    
    // Ngưỡng
    HighReputationThreshold float64 // 7.0
    LowReputationThreshold  float64 // 3.0
    PenaltyThreshold        int     // 3 vi phạm
}
```

## API Endpoints

### 1. Lấy thống kê reputation system
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getReputationStats","params":[],"id":1}' \
  http://localhost:8547
```

### 2. Lấy điểm reputation của validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getReputationScore","params":["0x1234..."],"id":1}' \
  http://localhost:8547
```

### 3. Lấy top validators theo reputation
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getTopValidators","params":[5],"id":1}' \
  http://localhost:8547
```

### 4. Lấy lịch sử reputation events
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getReputationEvents","params":[10],"id":1}' \
  http://localhost:8547
```

### 5. Ghi nhận vi phạm
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_recordViolation","params":["0x1234...", 123, "late_block", "Block was late"],"id":1}' \
  http://localhost:8547
```

### 6. Cập nhật reputation scores
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_updateReputation","params":[],"id":1}' \
  http://localhost:8547
```

### 7. Đánh dấu validator offline
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_markValidatorOffline","params":["0x1234..."],"id":1}' \
  http://localhost:8547
```

### 8. Cập nhật uptime validator
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_updateValidatorUptime","params":["0x1234..."],"id":1}' \
  http://localhost:8547
```

## Cách sử dụng

### 1. Khởi động nodes
```powershell
.\start_nodes.ps1
```

### 2. Test reputation system
```powershell
.\test_reputation_system.ps1
```

### 3. Test toàn bộ hệ thống
```powershell
.\quick_test.ps1
```

## Tích hợp với các hệ thống khác

### 🔗 **Validator Selection System**
- Reputation score được sử dụng trong validator selection
- Validators có reputation cao có khả năng được chọn cao hơn
- Tự động cập nhật reputation vào validator selection manager

### 🔗 **Anomaly Detection**
- Phát hiện vi phạm và ghi nhận vào reputation system
- Tự động áp dụng penalty cho validators vi phạm
- Theo dõi patterns bất thường

### 🔗 **Whitelist/Blacklist**
- Reputation thấp có thể dẫn đến blacklist
- Reputation cao có thể được whitelist ưu tiên
- Tích hợp với validation rules

## Lợi ích

### 1. **Minh bạch (Transparency)**
- Tất cả điểm số được lưu trữ on-chain
- Lịch sử events được ghi lại đầy đủ
- Có thể audit và verify

### 2. **Công bằng (Fairness)**
- Đánh giá dựa trên hiệu suất thực tế
- Không có bias hay favoritism
- Thuật toán công khai và minh bạch

### 3. **Động lực (Incentive)**
- Khuyến khích validators hoạt động tốt
- Penalty cho hành vi xấu
- Reward cho performance tốt

### 4. **Tự động (Automation)**
- Tự động tính điểm và cập nhật
- Tự động áp dụng penalty
- Tự động decay theo thời gian

## Ví dụ sử dụng

### Theo dõi performance validator
```bash
# Lấy điểm reputation hiện tại
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getReputationScore","params":["0x1111..."],"id":1}' \
  http://localhost:8547

# Kết quả:
{
  "result": {
    "address": "0x1111...",
    "current_score": 7.5,
    "block_mining_score": 3.2,
    "uptime_score": 2.1,
    "consistency_score": 1.8,
    "penalty_score": 0.0,
    "total_blocks_mined": 32,
    "uptime_hours": 42.0,
    "violation_count": 0,
    "is_active": true
  }
}
```

### Ghi nhận vi phạm
```bash
# Ghi nhận validator ký block muộn
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_recordViolation","params":["0x1111...", 123, "late_block", "Block was 5 seconds late"],"id":1}' \
  http://localhost:8547
```

### Lấy top validators
```bash
# Lấy 3 validators có reputation cao nhất
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getTopValidators","params":[3],"id":1}' \
  http://localhost:8547
```

## Monitoring và Analytics

### 📈 **Metrics được theo dõi:**
- Reputation score trends
- Block mining performance
- Uptime statistics
- Violation patterns
- Consistency metrics

### 📊 **Reports có thể tạo:**
- Validator performance reports
- Network health reports
- Anomaly detection reports
- Reputation distribution analysis

## Troubleshooting

### Lỗi "reputation system not initialized"
- Đảm bảo nodes đã khởi động hoàn toàn
- Kiểm tra genesis.json có đúng signers

### Reputation score không cập nhật
- Kiểm tra có blocks mới được tạo không
- Chạy `clique_updateReputation` để force update

### Validator không có reputation score
- Kiểm tra validator có được thêm vào hệ thống không
- Đảm bảo validator đã ký ít nhất 1 block

## Kết luận

Reputation System cung cấp một cách minh bạch và công bằng để đánh giá hiệu suất của validators trong POA consensus engine. Hệ thống tự động theo dõi, tính điểm và cập nhật reputation dựa trên hoạt động thực tế, tạo ra một môi trường cạnh tranh lành mạnh và khuyến khích validators hoạt động tốt nhất.
