# Hệ Thống Tracing với Merkle Tree

## Tổng Quan

Hệ thống Tracing với Merkle Tree là một lớp truy vết hành vi validator toàn diện, tạo ra dấu vết bất biến về tất cả các sự kiện quan trọng trong quá trình đồng thuận. Mỗi sự kiện được hash và lưu trữ trong Merkle Tree, với Merkle root được chèn vào block như một trường đặc biệt.

## Kiến Trúc Hệ Thống

### 1. **TraceEvent Structure**
```go
type TraceEvent struct {
    ID          string                 `json:"id"`
    Type        TraceEventType         `json:"type"`
    Timestamp   time.Time              `json:"timestamp"`
    BlockNumber uint64                 `json:"block_number"`
    Round       uint64                 `json:"round"`
    Address     common.Address         `json:"address"`
    Message     string                 `json:"message"`
    Data        map[string]interface{} `json:"data"`
    Level       TraceLevel             `json:"level"`
    Duration    time.Duration          `json:"duration,omitempty"`
    Hash        common.Hash            `json:"hash"`        // Hash của event
    MerklePath  []common.Hash          `json:"merkle_path"` // Path đến Merkle root
}
```

### 2. **Merkle Tree Structure**
```go
type MerkleTree struct {
    Root   *MerkleNode `json:"root"`
    Leaves []common.Hash `json:"leaves"`
    Events []TraceEvent `json:"events"`
}
```

### 3. **Trace Levels**
- **TraceLevelOff (0)**: Tắt tracing
- **TraceLevelBasic (1)**: Chỉ trace các sự kiện cơ bản
- **TraceLevelDetailed (2)**: Trace chi tiết các sự kiện
- **TraceLevelVerbose (3)**: Trace tất cả sự kiện với thông tin đầy đủ

## Các Loại Sự Kiện Được Trace

### 1. **Random POA Events**
- **Type**: `random_poa`
- **Mô tả**: Trace việc chọn validator ngẫu nhiên
- **Data**: seed, selected_signer, signers_count, is_selected

### 2. **Leader Selection Events**
- **Type**: `leader_selection`
- **Mô tả**: Trace việc chọn leader validator
- **Data**: selected_leader, selection_method, validator_count

### 3. **Block Signing Events**
- **Type**: `block_signing`
- **Mô tả**: Trace việc ký block thành công/thất bại
- **Data**: success, duration_ms, error

### 4. **Block Validation Events**
- **Type**: `block_validation`
- **Mô tả**: Trace việc xác thực block
- **Data**: signer, difficulty, timestamp, validation_result

### 5. **Timeout Events**
- **Type**: `timeout`
- **Mô tả**: Trace các sự kiện timeout
- **Data**: timeout_type, expected_time, actual_time, delay

### 6. **Accusation Events**
- **Type**: `accusation`
- **Mô tả**: Trace các cáo buộc gian lận
- **Data**: accuser, accused, accusation_type, evidence

### 7. **AI Gate Evaluation Events**
- **Type**: `ai_gate_evaluation`
- **Mô tả**: Trace đánh giá AI gate
- **Data**: evaluation_result, confidence, ai_model

### 8. **Reputation Events**
- **Type**: `reputation`
- **Mô tả**: Trace cập nhật reputation
- **Data**: event_type, old_score, new_score, score_change

### 9. **Anomaly Detection Events**
- **Type**: `anomaly_detection`
- **Mô tả**: Trace phát hiện anomaly
- **Data**: anomaly_type, severity, message, timestamp

### 10. **Whitelist/Blacklist Events**
- **Type**: `whitelist_blacklist`
- **Mô tả**: Trace quản lý access control
- **Data**: action, address, reason

### 11. **Validator Selection Events**
- **Type**: `validator_selection`
- **Mô tả**: Trace chọn validator
- **Data**: method, selected_count, selected_validators

## Merkle Tree Implementation

### 1. **Event Hashing**
```go
func (ts *TracingSystem) calculateEventHash(event TraceEvent) common.Hash {
    eventData := map[string]interface{}{
        "id":           event.ID,
        "type":         event.Type,
        "timestamp":    event.Timestamp.UnixNano(),
        "block_number": event.BlockNumber,
        "round":        event.Round,
        "address":      event.Address.Hex(),
        "message":      event.Message,
        "data":         event.Data,
        "duration":     event.Duration.Nanoseconds(),
    }
    
    jsonData, _ := json.Marshal(eventData)
    hash := sha256.Sum256(jsonData)
    return common.BytesToHash(hash[:])
}
```

### 2. **Merkle Tree Building**
```go
func (ts *TracingSystem) buildMerkleTree(leaves []common.Hash) *MerkleNode {
    if len(leaves) == 0 {
        return nil
    }
    
    if len(leaves) == 1 {
        return &MerkleNode{
            Hash: leaves[0],
            Data: leaves[0].Bytes(),
        }
    }
    
    // Build next level
    var nextLevel []common.Hash
    for i := 0; i < len(leaves); i += 2 {
        left := leaves[i]
        right := leaves[i+1]
        
        combined := append(left.Bytes(), right.Bytes()...)
        hash := sha256.Sum256(combined)
        nextLevel = append(nextLevel, common.BytesToHash(hash[:]))
    }
    
    return ts.buildMerkleTree(nextLevel)
}
```

### 3. **Merkle Proof Generation**
```go
func (ts *TracingSystem) generateMerkleProof(index int, leaves []common.Hash) []common.Hash {
    var proof []common.Hash
    
    // Sort leaves for consistent ordering
    sortedLeaves := make([]common.Hash, len(leaves))
    copy(sortedLeaves, leaves)
    sort.Slice(sortedLeaves, func(i, j int) bool {
        return sortedLeaves[i].Hex() < sortedLeaves[j].Hex()
    })
    
    // Generate proof by traversing the tree
    // ... (implementation details)
    
    return proof
}
```

## API Endpoints

### 1. **GetTracingStats**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getTracingStats","params":[],"id":1}' \
  http://localhost:8549
```

**Response:**
```json
{
  "result": {
    "config": {
      "enable_tracing": true,
      "trace_level": 2,
      "max_trace_events": 10000,
      "enable_merkle_tree": true,
      "merkle_root_in_block": true
    },
    "current_events": 150,
    "total_events": 150,
    "system_uptime": "2h30m15s",
    "current_round": 1045,
    "merkle_root": "0x1234...",
    "merkle_tree_events": 150
  }
}
```

### 2. **GetTraceEvents**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getTraceEvents","params":["block_validation", 2, 10],"id":1}' \
  http://localhost:8549
```

**Parameters:**
- `eventType`: Loại sự kiện ("" = tất cả)
- `level`: Mức trace (0-3)
- `limit`: Số lượng events tối đa

### 3. **GetMerkleRoot**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getMerkleRoot","params":[],"id":1}' \
  http://localhost:8549
```

### 4. **VerifyEventInMerkleTree**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_verifyEventInMerkleTree","params":[eventObject],"id":1}' \
  http://localhost:8549
```

### 5. **GetMerkleProof**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getMerkleProof","params":[eventObject],"id":1}' \
  http://localhost:8549
```

### 6. **ExportTraceEvents**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_exportTraceEvents","params":[],"id":1}' \
  http://localhost:8549
```

### 7. **SetTraceLevel**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_setTraceLevel","params":[3],"id":1}' \
  http://localhost:8549
```

### 8. **EnableTracing**
```bash
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_enableTracing","params":[true],"id":1}' \
  http://localhost:8549
```

## Cấu Hình

### Default Configuration
```go
type TracingConfig struct {
    EnableTracing     bool        // true
    TraceLevel        TraceLevel  // TraceLevelDetailed (2)
    MaxTraceEvents    int         // 10000
    TraceRetention    time.Duration // 24 hours
    EnableMerkleTree  bool        // true
    EnablePersistence bool        // true
    EnableMetrics     bool        // true
    MerkleRootInBlock bool        // true
}
```

## Lợi Ích Của Hệ Thống

### 1. **Tính Minh Bạch**
- Tất cả hành vi validator được ghi lại
- Có thể audit và verify mọi sự kiện
- Merkle root cung cấp bằng chứng bất biến

### 2. **Tính Bảo Mật**
- Events không thể bị thay đổi sau khi được hash
- Merkle proof cho phép verify từng event cụ thể
- Tamper-resistant audit trail

### 3. **Tính Hiệu Quả**
- O(log n) proof size cho verification
- Chỉ cần Merkle root để verify toàn bộ tree
- Efficient storage và retrieval

### 4. **Tính Linh Hoạt**
- Có thể filter events theo type và level
- Export/import toàn bộ trace data
- Real-time metrics và monitoring

## Use Cases

### 1. **Audit và Compliance**
- Kiểm tra hành vi validator theo thời gian
- Verify các sự kiện cụ thể
- Tạo audit report cho regulators

### 2. **Debug và Troubleshooting**
- Trace các vấn đề trong consensus
- Phân tích performance của validators
- Identify patterns và anomalies

### 3. **Community Monitoring**
- Cộng đồng có thể verify events
- Tự động phát hiện gian lận
- Tạo trust và transparency

### 4. **Research và Analysis**
- Phân tích behavior patterns
- Nghiên cứu consensus mechanisms
- Optimize validator performance

## Test Script

Sử dụng script `test_tracing_system.ps1` để test toàn bộ hệ thống:

```powershell
.\Testnet\test_tracing_system.ps1
```

**Script test:**
1. ✅ Tracing statistics
2. ✅ Trace events retrieval
3. ✅ Merkle root functionality
4. ✅ Trace metrics
5. ✅ Trace level control
6. ✅ Event filtering
7. ✅ Merkle proof functionality
8. ✅ Event verification
9. ✅ Export functionality
10. ✅ Clear functionality

## Kết Luận

Hệ thống Tracing với Merkle Tree cung cấp:

- **Immutable Audit Trail**: Dấu vết bất biến về tất cả hành vi validator
- **Tamper-Proof Evidence**: Merkle root cung cấp bằng chứng không thể thay đổi
- **Community Verifiability**: Bất kỳ ai cũng có thể verify events
- **Comprehensive Monitoring**: Theo dõi toàn diện tất cả aspects của consensus
- **Transparent Governance**: Minh bạch trong quản lý và vận hành

Hệ thống này tạo ra một nền tảng vững chắc cho việc xây dựng trust, transparency và accountability trong blockchain networks.
