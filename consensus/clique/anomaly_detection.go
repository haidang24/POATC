// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package clique

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// AnomalyType represents different types of anomalies that can be detected
type AnomalyType int

const (
	AnomalyNone              AnomalyType = iota
	AnomalyRapidSigning                  // Signer signs too many blocks in a short time
	AnomalySuspiciousPattern             // Suspicious signing pattern detected
	AnomalyHighFrequency                 // Signer appears too frequently
	AnomalyMissingSigner                 // Expected signer is missing for too long
	AnomalyTimestampDrift                // Block timestamps show unusual patterns
)

// AnomalyDetectionConfig contains configuration for anomaly detection
type AnomalyDetectionConfig struct {
	// Time windows for analysis
	AnalysisWindow  time.Duration // Time window to analyze (e.g., 1 hour)
	BlockTimeWindow time.Duration // Expected block time window

	// Thresholds for anomaly detection
	MaxBlocksPerSigner int     // Maximum blocks a signer can sign in analysis window
	MaxSignerFrequency float64 // Maximum frequency a signer can appear (0.0-1.0)
	MinSignerFrequency float64 // Minimum frequency a signer should appear (0.0-1.0)
	MaxTimestampDrift  int64   // Maximum timestamp drift in seconds

	// Pattern detection
	PatternWindowSize   int // Number of blocks to analyze for patterns
	SuspiciousThreshold int // Number of consecutive blocks by same signer to trigger alert
}

// DefaultAnomalyDetectionConfig returns a default configuration for anomaly detection
func DefaultAnomalyDetectionConfig() *AnomalyDetectionConfig {
	return &AnomalyDetectionConfig{
		AnalysisWindow:      1 * time.Hour,
		BlockTimeWindow:     15 * time.Second,
		MaxBlocksPerSigner:  10,
		MaxSignerFrequency:  0.6, // 60% max frequency
		MinSignerFrequency:  0.1, // 10% min frequency
		MaxTimestampDrift:   30,  // 30 seconds
		PatternWindowSize:   20,
		SuspiciousThreshold: 5,
	}
}

// AnomalyResult contains the result of anomaly detection
type AnomalyResult struct {
	Type        AnomalyType            `json:"type"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Message     string                 `json:"message"`
	Signer      common.Address         `json:"signer,omitempty"`
	BlockNumber uint64                 `json:"block_number"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// AnomalyDetector handles anomaly detection for POA consensus
type AnomalyDetector struct {
	config       *AnomalyDetectionConfig
	signers      map[common.Address]struct{}
	blockHistory []BlockRecord
}

// BlockRecord represents a block for anomaly analysis
type BlockRecord struct {
	Number     uint64
	Hash       common.Hash
	Signer     common.Address
	Timestamp  time.Time
	Difficulty *big.Int
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config *AnomalyDetectionConfig, signers []common.Address) *AnomalyDetector {
	if config == nil {
		config = DefaultAnomalyDetectionConfig()
	}

	signerMap := make(map[common.Address]struct{})
	for _, signer := range signers {
		signerMap[signer] = struct{}{}
	}

	return &AnomalyDetector{
		config:       config,
		signers:      signerMap,
		blockHistory: make([]BlockRecord, 0),
	}
}

// AddBlock adds a new block to the analysis history
func (ad *AnomalyDetector) AddBlock(header *types.Header, signer common.Address) {
	record := BlockRecord{
		Number:     header.Number.Uint64(),
		Hash:       header.Hash(),
		Signer:     signer,
		Timestamp:  time.Unix(int64(header.Time), 0),
		Difficulty: header.Difficulty,
	}

	ad.blockHistory = append(ad.blockHistory, record)

	// Keep only recent history to avoid memory issues
	if len(ad.blockHistory) > ad.config.PatternWindowSize*2 {
		ad.blockHistory = ad.blockHistory[len(ad.blockHistory)-ad.config.PatternWindowSize:]
	}
}

// DetectAnomalies analyzes the block history and detects anomalies
func (ad *AnomalyDetector) DetectAnomalies() []AnomalyResult {
	var anomalies []AnomalyResult

	if len(ad.blockHistory) < 3 {
		return anomalies // Need at least 3 blocks for meaningful analysis
	}

	// Check for rapid signing
	anomalies = append(anomalies, ad.detectRapidSigning()...)

	// Check for suspicious patterns
	anomalies = append(anomalies, ad.detectSuspiciousPatterns()...)

	// Check for frequency anomalies
	anomalies = append(anomalies, ad.detectFrequencyAnomalies()...)

	// Check for timestamp drift
	anomalies = append(anomalies, ad.detectTimestampDrift()...)

	// Check for missing signers
	anomalies = append(anomalies, ad.detectMissingSigners()...)

	return anomalies
}

// detectRapidSigning detects if a signer is signing too many blocks in a short time
func (ad *AnomalyDetector) detectRapidSigning() []AnomalyResult {
	var anomalies []AnomalyResult
	signerCounts := make(map[common.Address]int)

	// Count blocks per signer in recent history
	for _, record := range ad.blockHistory {
		signerCounts[record.Signer]++
	}

	// Check if any signer exceeds the threshold
	for signer, count := range signerCounts {
		if count > ad.config.MaxBlocksPerSigner {
			severity := "medium"
			if count > ad.config.MaxBlocksPerSigner*2 {
				severity = "high"
			}
			if count > ad.config.MaxBlocksPerSigner*3 {
				severity = "critical"
			}

			anomalies = append(anomalies, AnomalyResult{
				Type:     AnomalyRapidSigning,
				Severity: severity,
				Message: fmt.Sprintf("Signer %s has signed %d blocks in recent history (max: %d)",
					signer.Hex(), count, ad.config.MaxBlocksPerSigner),
				Signer:      signer,
				BlockNumber: ad.blockHistory[len(ad.blockHistory)-1].Number,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"blocks_signed": count,
					"max_allowed":   ad.config.MaxBlocksPerSigner,
					"window_size":   len(ad.blockHistory),
				},
			})
		}
	}

	return anomalies
}

// detectSuspiciousPatterns detects suspicious signing patterns
func (ad *AnomalyDetector) detectSuspiciousPatterns() []AnomalyResult {
	var anomalies []AnomalyResult

	if len(ad.blockHistory) < ad.config.SuspiciousThreshold {
		return anomalies
	}

	// Check for consecutive blocks by the same signer
	consecutiveCount := 1
	currentSigner := ad.blockHistory[len(ad.blockHistory)-1].Signer

	for i := len(ad.blockHistory) - 2; i >= 0; i-- {
		if ad.blockHistory[i].Signer == currentSigner {
			consecutiveCount++
		} else {
			break
		}
	}

	// Debug logging
	log.Debug("Suspicious pattern detection",
		"consecutive_count", consecutiveCount,
		"threshold", ad.config.SuspiciousThreshold,
		"current_signer", currentSigner.Hex())

	if consecutiveCount >= ad.config.SuspiciousThreshold {
		severity := "low"
		if consecutiveCount >= ad.config.SuspiciousThreshold*2 {
			severity = "medium"
		}
		if consecutiveCount >= ad.config.SuspiciousThreshold*3 {
			severity = "high"
		}

		anomalies = append(anomalies, AnomalyResult{
			Type:     AnomalySuspiciousPattern,
			Severity: severity,
			Message: fmt.Sprintf("Signer %s has signed %d consecutive blocks",
				currentSigner.Hex(), consecutiveCount),
			Signer:      currentSigner,
			BlockNumber: ad.blockHistory[len(ad.blockHistory)-1].Number,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"consecutive_blocks": consecutiveCount,
				"threshold":          ad.config.SuspiciousThreshold,
			},
		})
	}

	return anomalies
}

// detectFrequencyAnomalies detects if signers appear too frequently or too rarely
func (ad *AnomalyDetector) detectFrequencyAnomalies() []AnomalyResult {
	var anomalies []AnomalyResult
	signerCounts := make(map[common.Address]int)
	totalBlocks := len(ad.blockHistory)

	// Count blocks per signer
	for _, record := range ad.blockHistory {
		signerCounts[record.Signer]++
	}

	// Check frequency for each signer
	for signer := range ad.signers {
		count := signerCounts[signer]
		frequency := float64(count) / float64(totalBlocks)

		if frequency > ad.config.MaxSignerFrequency {
			anomalies = append(anomalies, AnomalyResult{
				Type:     AnomalyHighFrequency,
				Severity: "medium",
				Message: fmt.Sprintf("Signer %s appears too frequently: %.2f%% (max: %.2f%%)",
					signer.Hex(), frequency*100, ad.config.MaxSignerFrequency*100),
				Signer:      signer,
				BlockNumber: ad.blockHistory[len(ad.blockHistory)-1].Number,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"frequency":     frequency,
					"max_frequency": ad.config.MaxSignerFrequency,
					"blocks_signed": count,
					"total_blocks":  totalBlocks,
				},
			})
		}

		if frequency < ad.config.MinSignerFrequency && totalBlocks > 10 {
			anomalies = append(anomalies, AnomalyResult{
				Type:     AnomalyMissingSigner,
				Severity: "low",
				Message: fmt.Sprintf("Signer %s appears too rarely: %.2f%% (min: %.2f%%)",
					signer.Hex(), frequency*100, ad.config.MinSignerFrequency*100),
				Signer:      signer,
				BlockNumber: ad.blockHistory[len(ad.blockHistory)-1].Number,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"frequency":     frequency,
					"min_frequency": ad.config.MinSignerFrequency,
					"blocks_signed": count,
					"total_blocks":  totalBlocks,
				},
			})
		}
	}

	return anomalies
}

// detectTimestampDrift detects unusual timestamp patterns
func (ad *AnomalyDetector) detectTimestampDrift() []AnomalyResult {
	var anomalies []AnomalyResult

	if len(ad.blockHistory) < 2 {
		return anomalies
	}

	// Check timestamp differences between consecutive blocks
	for i := 1; i < len(ad.blockHistory); i++ {
		prevTime := ad.blockHistory[i-1].Timestamp
		currTime := ad.blockHistory[i].Timestamp
		diff := currTime.Sub(prevTime)

		// Check if timestamp difference is too large or too small
		if diff.Seconds() > float64(ad.config.MaxTimestampDrift) {
			anomalies = append(anomalies, AnomalyResult{
				Type:     AnomalyTimestampDrift,
				Severity: "medium",
				Message: fmt.Sprintf("Large timestamp drift detected: %.2f seconds between blocks %d and %d",
					diff.Seconds(), ad.blockHistory[i-1].Number, ad.blockHistory[i].Number),
				BlockNumber: ad.blockHistory[i].Number,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"time_drift_seconds": diff.Seconds(),
					"max_drift_seconds":  ad.config.MaxTimestampDrift,
					"prev_block":         ad.blockHistory[i-1].Number,
					"curr_block":         ad.blockHistory[i].Number,
				},
			})
		}
	}

	return anomalies
}

// detectMissingSigners detects if expected signers are missing for too long
func (ad *AnomalyDetector) detectMissingSigners() []AnomalyResult {
	var anomalies []AnomalyResult

	if len(ad.blockHistory) < 10 {
		return anomalies // Need enough history
	}

	// Check last 10 blocks for missing signers
	recentSigners := make(map[common.Address]bool)
	for i := len(ad.blockHistory) - 10; i < len(ad.blockHistory); i++ {
		recentSigners[ad.blockHistory[i].Signer] = true
	}

	// Check if any authorized signer is missing
	for signer := range ad.signers {
		if !recentSigners[signer] {
			anomalies = append(anomalies, AnomalyResult{
				Type:     AnomalyMissingSigner,
				Severity: "low",
				Message: fmt.Sprintf("Signer %s has not signed any blocks in recent history",
					signer.Hex()),
				Signer:      signer,
				BlockNumber: ad.blockHistory[len(ad.blockHistory)-1].Number,
				Timestamp:   time.Now(),
				Details: map[string]interface{}{
					"recent_blocks_checked": 10,
					"total_signers":         len(ad.signers),
				},
			})
		}
	}

	return anomalies
}

// LogAnomalies logs detected anomalies with appropriate log levels
func (ad *AnomalyDetector) LogAnomalies(anomalies []AnomalyResult) {
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "critical":
			log.Error("POA Anomaly Detected",
				"type", anomaly.Type,
				"severity", anomaly.Severity,
				"message", anomaly.Message,
				"signer", anomaly.Signer.Hex(),
				"block", anomaly.BlockNumber,
				"details", anomaly.Details)
		case "high":
			log.Warn("POA Anomaly Detected",
				"type", anomaly.Type,
				"severity", anomaly.Severity,
				"message", anomaly.Message,
				"signer", anomaly.Signer.Hex(),
				"block", anomaly.BlockNumber,
				"details", anomaly.Details)
		case "medium":
			log.Info("POA Anomaly Detected",
				"type", anomaly.Type,
				"severity", anomaly.Severity,
				"message", anomaly.Message,
				"signer", anomaly.Signer.Hex(),
				"block", anomaly.BlockNumber,
				"details", anomaly.Details)
		case "low":
			log.Debug("POA Anomaly Detected",
				"type", anomaly.Type,
				"severity", anomaly.Severity,
				"message", anomaly.Message,
				"signer", anomaly.Signer.Hex(),
				"block", anomaly.BlockNumber,
				"details", anomaly.Details)
		}
	}
}

// GetAnomalyStats returns statistics about detected anomalies
func (ad *AnomalyDetector) GetAnomalyStats() map[string]interface{} {
	anomalies := ad.DetectAnomalies()

	stats := map[string]interface{}{
		"total_anomalies":    len(anomalies),
		"by_type":            make(map[AnomalyType]int),
		"by_severity":        make(map[string]int),
		"block_history_size": len(ad.blockHistory),
	}

	for _, anomaly := range anomalies {
		// Count by type
		if count, ok := stats["by_type"].(map[AnomalyType]int); ok {
			count[anomaly.Type]++
		}

		// Count by severity
		if count, ok := stats["by_severity"].(map[string]int); ok {
			count[anomaly.Severity]++
		}
	}

	return stats
}
