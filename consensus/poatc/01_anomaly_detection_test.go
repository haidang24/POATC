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

package poatc

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TestAnomalyDetectorBasic tests basic functionality of the anomaly detector
func TestAnomalyDetectorBasic(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		common.HexToAddress("0x3333333333333333333333333333333333333333"),
	}

	detector := NewAnomalyDetector(DefaultAnomalyDetectionConfig(), signers)

	// Test with normal block pattern
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		header := &types.Header{
			Number:     big.NewInt(int64(i + 1)),
			Time:       uint64(baseTime.Add(time.Duration(i) * 15 * time.Second).Unix()),
			Difficulty: big.NewInt(1),
		}
		// Hash is computed automatically, we don't need to set it manually

		signer := signers[i%len(signers)]
		detector.AddBlock(header, signer)
	}

	anomalies := detector.DetectAnomalies()
	if len(anomalies) > 0 {
		t.Logf("Detected %d anomalies in normal pattern:", len(anomalies))
		for _, anomaly := range anomalies {
			t.Logf("  - %s: %s", anomaly.Severity, anomaly.Message)
		}
	} else {
		t.Log("No anomalies detected in normal pattern - good!")
	}
}

// TestAnomalyDetectorRapidSigning tests detection of rapid signing
func TestAnomalyDetectorRapidSigning(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}

	config := DefaultAnomalyDetectionConfig()
	config.MaxBlocksPerSigner = 3 // Very low threshold for testing
	detector := NewAnomalyDetector(config, signers)

	// Create blocks where one signer signs too many blocks
	baseTime := time.Now()
	for i := 0; i < 8; i++ {
		header := &types.Header{
			Number:     big.NewInt(int64(i + 1)),
			Time:       uint64(baseTime.Add(time.Duration(i) * 15 * time.Second).Unix()),
			Difficulty: big.NewInt(1),
		}
		// Hash is computed automatically, we don't need to set it manually

		// First signer signs 5 blocks, second signer signs 3 blocks
		var signer common.Address
		if i < 5 {
			signer = signers[0]
		} else {
			signer = signers[1]
		}
		detector.AddBlock(header, signer)
	}

	anomalies := detector.DetectAnomalies()

	// Should detect rapid signing anomaly
	rapidSigningFound := false
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyRapidSigning {
			rapidSigningFound = true
			t.Logf("✓ Detected rapid signing anomaly: %s", anomaly.Message)
			break
		}
	}

	if !rapidSigningFound {
		t.Error("Expected to detect rapid signing anomaly but none found")
	}

	t.Logf("Total anomalies detected: %d", len(anomalies))
}

// TestAnomalyDetectorSuspiciousPattern tests detection of suspicious patterns
func TestAnomalyDetectorSuspiciousPattern(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}

	config := DefaultAnomalyDetectionConfig()
	config.SuspiciousThreshold = 3 // Low threshold for testing
	detector := NewAnomalyDetector(config, signers)

	// Create blocks with consecutive signing pattern
	baseTime := time.Now()

	// Create some blocks by different signers first
	header1 := &types.Header{
		Number:     big.NewInt(1),
		Time:       uint64(baseTime.Unix()),
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header1, signers[1])

	header2 := &types.Header{
		Number:     big.NewInt(2),
		Time:       uint64(baseTime.Add(15 * time.Second).Unix()),
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header2, signers[1])

	// Now create 3 consecutive blocks by same signer at the end
	for i := 0; i < 3; i++ {
		header := &types.Header{
			Number:     big.NewInt(int64(i + 3)),
			Time:       uint64(baseTime.Add(time.Duration(i+2) * 15 * time.Second).Unix()),
			Difficulty: big.NewInt(1),
		}
		detector.AddBlock(header, signers[0])
	}

	anomalies := detector.DetectAnomalies()

	// Debug: print all anomalies
	t.Logf("All anomalies detected:")
	for i, anomaly := range anomalies {
		t.Logf("  %d. Type: %v, Severity: %s, Message: %s", i+1, anomaly.Type, anomaly.Severity, anomaly.Message)
	}

	// Should detect suspicious pattern
	suspiciousPatternFound := false
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalySuspiciousPattern {
			suspiciousPatternFound = true
			t.Logf("✓ Detected suspicious pattern: %s", anomaly.Message)
			break
		}
	}

	if !suspiciousPatternFound {
		t.Error("Expected to detect suspicious pattern but none found")
	}

	t.Logf("Total anomalies detected: %d", len(anomalies))
}

// TestAnomalyDetectorTimestampDrift tests detection of timestamp drift
func TestAnomalyDetectorTimestampDrift(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}

	config := DefaultAnomalyDetectionConfig()
	config.MaxTimestampDrift = 30 // 30 seconds max drift
	detector := NewAnomalyDetector(config, signers)

	// Create blocks with large timestamp drift
	baseTime := time.Now()

	// First block
	header1 := &types.Header{
		Number:     big.NewInt(1),
		Time:       uint64(baseTime.Unix()),
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header1, signers[0])

	// Second block with large drift
	header2 := &types.Header{
		Number:     big.NewInt(2),
		Time:       uint64(baseTime.Add(60 * time.Second).Unix()), // 60 seconds drift
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header2, signers[1])

	// Third block with even larger drift
	header3 := &types.Header{
		Number:     big.NewInt(3),
		Time:       uint64(baseTime.Add(120 * time.Second).Unix()), // 120 seconds drift
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header3, signers[0])

	// Add one more block to ensure we have enough for analysis
	header4 := &types.Header{
		Number:     big.NewInt(4),
		Time:       uint64(baseTime.Add(180 * time.Second).Unix()),
		Difficulty: big.NewInt(1),
	}
	detector.AddBlock(header4, signers[1])

	anomalies := detector.DetectAnomalies()

	// Should detect timestamp drift
	timestampDriftFound := false
	for _, anomaly := range anomalies {
		if anomaly.Type == AnomalyTimestampDrift {
			timestampDriftFound = true
			t.Logf("✓ Detected timestamp drift: %s", anomaly.Message)
			break
		}
	}

	if !timestampDriftFound {
		t.Error("Expected to detect timestamp drift but none found")
	}

	t.Logf("Total anomalies detected: %d", len(anomalies))
}

// TestAnomalyDetectorStats tests the statistics functionality
func TestAnomalyDetectorStats(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}

	detector := NewAnomalyDetector(DefaultAnomalyDetectionConfig(), signers)

	// Add some blocks
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		header := &types.Header{
			Number:     big.NewInt(int64(i + 1)),
			Time:       uint64(baseTime.Add(time.Duration(i) * 15 * time.Second).Unix()),
			Difficulty: big.NewInt(1),
		}
		// Hash is computed automatically, we don't need to set it manually

		signer := signers[i%len(signers)]
		detector.AddBlock(header, signer)
	}

	stats := detector.GetAnomalyStats()

	t.Logf("Anomaly detector stats: %+v", stats)

	// Check that stats contain expected fields
	if stats["total_anomalies"] == nil {
		t.Error("Stats should contain total_anomalies field")
	}
	if stats["block_history_size"] == nil {
		t.Error("Stats should contain block_history_size field")
	}
}
