# POA Anomaly Detection System

## Tổng quan

Hệ thống phát hiện bất thường (Anomaly Detection) được tích hợp vào thuật toán Proof of Authority (POA) để giám sát và phát hiện các hành vi bất thường trong quá trình ký block.

## Tính năng chính

### 1. Phát hiện Rapid Signing
- **Mục đích**: Phát hiện signer ký quá nhiều blocks trong thời gian ngắn
- **Cấu hình**: `MaxBlocksPerSigner` - số blocks tối đa một signer có thể ký
- **Severity**: Medium/High/Critical tùy theo mức độ vượt quá

### 2. Phát hiện Suspicious Patterns
- **Mục đích**: Phát hiện pattern ký liên tiếp bất thường
- **Cấu hình**: `SuspiciousThreshold` - số blocks liên tiếp tối đa
- **Severity**: Low/Medium/High tùy theo số blocks liên tiếp

### 3. Phát hiện Frequency Anomalies
- **Mục đích**: Phát hiện signer xuất hiện quá thường xuyên hoặc quá ít
- **Cấu hình**: 
  - `MaxSignerFrequency` - tần suất tối đa (0.0-1.0)
  - `MinSignerFrequency` - tần suất tối thiểu (0.0-1.0)
- **Severity**: Medium cho tần suất cao, Low cho tần suất thấp

### 4. Phát hiện Timestamp Drift
- **Mục đích**: Phát hiện timestamp bất thường giữa các blocks
- **Cấu hình**: `MaxTimestampDrift` - độ lệch timestamp tối đa (giây)
- **Severity**: Medium

### 5. Phát hiện Missing Signers
- **Mục đích**: Phát hiện signer không ký blocks trong thời gian dài
- **Cấu hình**: Kiểm tra 10 blocks gần nhất
- **Severity**: Low

## Cấu hình mặc định

```go
AnomalyDetectionConfig{
    AnalysisWindow:        1 * time.Hour,
    BlockTimeWindow:       15 * time.Second,
    MaxBlocksPerSigner:    10,
    MaxSignerFrequency:    0.6, // 60%
    MinSignerFrequency:    0.1, // 10%
    MaxTimestampDrift:     30,  // 30 seconds
    PatternWindowSize:     20,
    SuspiciousThreshold:   5,
}
```

## Cách sử dụng

### 1. Tự động phát hiện
Hệ thống tự động phát hiện anomalies khi verify blocks:
```go
// Trong verifySeal()
c.anomalyDetector.AddBlock(header, signer)
anomalies := c.anomalyDetector.DetectAnomalies()
c.anomalyDetector.LogAnomalies(anomalies)
```

### 2. API truy cập
```javascript
// Lấy thống kê anomalies
clique.getAnomalyStats()

// Phát hiện anomalies thủ công
clique.detectAnomalies()

// Lấy cấu hình
clique.getAnomalyConfig()
```

### 3. Logging
Anomalies được log với các mức độ khác nhau:
- **Critical**: log.Error()
- **High**: log.Warn()
- **Medium**: log.Info()
- **Low**: log.Debug()

## Cấu trúc dữ liệu

### AnomalyResult
```go
type AnomalyResult struct {
    Type        AnomalyType           `json:"type"`
    Severity    string                `json:"severity"`
    Message     string                `json:"message"`
    Signer      common.Address        `json:"signer,omitempty"`
    BlockNumber uint64                `json:"block_number"`
    Timestamp   time.Time             `json:"timestamp"`
    Details     map[string]interface{} `json:"details,omitempty"`
}
```

### AnomalyType
```go
const (
    AnomalyNone AnomalyType = iota
    AnomalyRapidSigning
    AnomalySuspiciousPattern
    AnomalyHighFrequency
    AnomalyMissingSigner
    AnomalyTimestampDrift
)
```

## Test Results

### TestAnomalyDetectorBasic
- ✅ Không phát hiện anomalies trong pattern bình thường

### TestAnomalyDetectorRapidSigning
- ✅ Phát hiện signer ký quá nhiều blocks (5/3)

### TestAnomalyDetectorSuspiciousPattern
- ✅ Phát hiện pattern ký liên tiếp (3 blocks)

### TestAnomalyDetectorTimestampDrift
- ✅ Phát hiện timestamp drift (60 giây)

### TestAnomalyDetectorStats
- ✅ Thống kê hoạt động đúng

## Tích hợp với Random POA

Hệ thống anomaly detection hoạt động hoàn hảo với thuật toán Random POA mới:
- Phát hiện các pattern bất thường trong việc chọn signer ngẫu nhiên
- Giám sát tần suất xuất hiện của các signers
- Đảm bảo tính công bằng trong việc phân phối blocks

## Lợi ích

1. **Bảo mật**: Phát hiện sớm các hành vi bất thường
2. **Giám sát**: Theo dõi hoạt động của các signers
3. **Cảnh báo**: Thông báo kịp thời về các vấn đề tiềm ẩn
4. **Phân tích**: Cung cấp thống kê chi tiết về hoạt động mạng
5. **Tự động**: Hoạt động tự động không cần can thiệp thủ công

## Files liên quan

1. `anomaly_detection.go` - Core implementation
2. `anomaly_detection_test.go` - Test cases
3. `clique.go` - Integration với POA engine
4. `api.go` - RPC API endpoints
5. `ANOMALY_DETECTION_README.md` - Tài liệu này

## Build và Test

```bash
cd consensus/clique
go test -v -run TestAnomalyDetector
```

Tất cả tests đều pass, xác nhận hệ thống hoạt động đúng.
