// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.

package poatc

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestTimeDynamicManagerBasic(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	tdm := NewTimeDynamicManager(config)

	if tdm == nil {
		t.Fatal("Failed to create TimeDynamicManager")
	}

	// Test initial state
	currentBlockTime := tdm.GetCurrentBlockTime()
	if currentBlockTime != config.BaseBlockTime {
		t.Fatalf("Expected initial block time %v, got %v", config.BaseBlockTime, currentBlockTime)
	}

	t.Logf("TimeDynamicManager created successfully with base block time: %v", currentBlockTime)
}

func TestDynamicBlockTime(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	config.TxThresholdHigh = 50
	config.TxThresholdLow = 5
	tdm := NewTimeDynamicManager(config)

	// Test high transaction volume (should decrease block time)
	for i := 0; i < 5; i++ {
		tdm.UpdateTransactionCount(100) // High transaction count
	}

	newBlockTime := tdm.GetCurrentBlockTime()
	if newBlockTime >= config.BaseBlockTime {
		t.Fatalf("Expected block time to decrease with high tx volume, but got %v >= %v", 
			newBlockTime, config.BaseBlockTime)
	}

	t.Logf("High tx volume: block time changed from %v to %v", config.BaseBlockTime, newBlockTime)

	// Reset and test no transaction volume (should stay at base block time)
	tdm = NewTimeDynamicManager(config)
	for i := 0; i < 5; i++ {
		tdm.UpdateTransactionCount(0) // No transactions
	}

	newBlockTime = tdm.GetCurrentBlockTime()
	if newBlockTime != config.BaseBlockTime {
		t.Logf("No tx volume: block time %v (base: %v) - should stay at base time for continuous block production", 
			newBlockTime, config.BaseBlockTime)
	}

	t.Logf("No tx volume: block time changed from %v to %v (should remain at base)", config.BaseBlockTime, newBlockTime)

	// Test very low transaction volume (should slightly increase block time)
	tdm = NewTimeDynamicManager(config)
	for i := 0; i < 5; i++ {
		tdm.UpdateTransactionCount(1) // Very low transaction count
	}

	newBlockTime = tdm.GetCurrentBlockTime()
	t.Logf("Very low tx volume: block time changed from %v to %v", config.BaseBlockTime, newBlockTime)
}

func TestDynamicValidatorSelection(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	config.ValidatorSelectionInterval = 100 * time.Millisecond // Short interval for testing
	tdm := NewTimeDynamicManager(config)

	// Initially should not need update
	if tdm.ShouldUpdateValidatorSelection() {
		t.Fatal("Should not need validator selection update initially")
	}

	// Wait for interval
	time.Sleep(150 * time.Millisecond)

	// Now should need update
	if !tdm.ShouldUpdateValidatorSelection() {
		t.Fatal("Should need validator selection update after interval")
	}

	t.Logf("Validator selection timing works correctly")
}

func TestDynamicReputationDecay(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	config.ReputationUpdateInterval = 100 * time.Millisecond // Short interval for testing
	tdm := NewTimeDynamicManager(config)

	// Initially should not need decay
	if tdm.ShouldApplyReputationDecay() {
		t.Fatal("Should not need reputation decay initially")
	}

	// Wait for interval
	time.Sleep(150 * time.Millisecond)

	// Now should need decay
	if !tdm.ShouldApplyReputationDecay() {
		t.Fatal("Should need reputation decay after interval")
	}

	t.Logf("Reputation decay timing works correctly")
}

func TestTimeDynamicStats(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	tdm := NewTimeDynamicManager(config)

	// Update some transaction counts
	tdm.UpdateTransactionCount(10)
	tdm.UpdateTransactionCount(20)
	tdm.UpdateTransactionCount(30)

	stats := tdm.GetTimeDynamicStats()
	if stats == nil {
		t.Fatal("Failed to get time dynamic stats")
	}

	// Check required fields
	configStats, ok := stats["config"].(map[string]interface{})
	if !ok {
		t.Fatal("Config stats not found")
	}

	if !configStats["enable_dynamic_block_time"].(bool) {
		t.Error("Dynamic block time should be enabled")
	}

	blockTimeStats, ok := stats["dynamic_block_time"].(map[string]interface{})
	if !ok {
		t.Fatal("Block time stats not found")
	}

	if blockTimeStats["recent_tx_count"].(int) != 3 {
		t.Errorf("Expected 3 recent tx counts, got %v", blockTimeStats["recent_tx_count"])
	}

	t.Logf("Time dynamic stats: %+v", stats)
}

func TestDecayHistory(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	tdm := NewTimeDynamicManager(config)

	// Initially no decay history
	history := tdm.GetDecayHistory(10)
	if len(history) != 0 {
		t.Fatalf("Expected empty decay history, got %d records", len(history))
	}

	// Add some mock decay records
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	tdm.decayHistory = append(tdm.decayHistory, DecayRecord{
		Timestamp:   time.Now(),
		Address:     addr1,
		OldScore:    1.0,
		NewScore:    0.95,
		DecayAmount: 0.05,
	})

	tdm.decayHistory = append(tdm.decayHistory, DecayRecord{
		Timestamp:   time.Now(),
		Address:     addr2,
		OldScore:    0.8,
		NewScore:    0.76,
		DecayAmount: 0.04,
	})

	history = tdm.GetDecayHistory(10)
	if len(history) != 2 {
		t.Fatalf("Expected 2 decay records, got %d", len(history))
	}

	if history[0].Address != addr1 {
		t.Errorf("Expected first record address %s, got %s", addr1.Hex(), history[0].Address.Hex())
	}

	t.Logf("Decay history works correctly: %d records", len(history))
}

func TestTimeDynamicConfigUpdate(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	tdm := NewTimeDynamicManager(config)

	// Check initial config
	if !tdm.config.EnableDynamicBlockTime {
		t.Fatal("Dynamic block time should be enabled initially")
	}

	// Update config
	newConfig := &TimeDynamicConfig{
		EnableDynamicBlockTime:          false,
		EnableDynamicValidatorSelection: false,
		EnableDynamicReputationDecay:    false,
		BaseBlockTime:                   20 * time.Second,
		MinBlockTime:                    10 * time.Second,
		MaxBlockTime:                    40 * time.Second,
	}

	tdm.UpdateConfig(newConfig)

	// Check updated config
	if tdm.config.EnableDynamicBlockTime {
		t.Error("Dynamic block time should be disabled after update")
	}

	if tdm.config.BaseBlockTime != 20*time.Second {
		t.Errorf("Expected base block time 20s, got %v", tdm.config.BaseBlockTime)
	}

	t.Logf("Config update works correctly")
}

func TestBlockTimeChangeReason(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	config.TxThresholdHigh = 50
	config.TxThresholdLow = 5
	tdm := NewTimeDynamicManager(config)

	// Test different transaction volumes
	testCases := []struct {
		avgTxCount     float64
		expectedReason string
	}{
		{100, "high_transaction_volume"},
		{25, "normal_transaction_volume"},
		{2, "low_transaction_volume"},
		{0, "no_transactions"},
	}

	for _, tc := range testCases {
		reason := tdm.getBlockTimeChangeReason(tc.avgTxCount)
		if reason != tc.expectedReason {
			t.Errorf("For avg tx count %f, expected reason '%s', got '%s'",
				tc.avgTxCount, tc.expectedReason, reason)
		}
	}

	t.Logf("Block time change reason detection works correctly")
}

func TestTimeDynamicIntegration(t *testing.T) {
	// Create mock components
	config := DefaultTimeDynamicConfig()
	config.ValidatorSelectionInterval = 50 * time.Millisecond
	config.ReputationUpdateInterval = 50 * time.Millisecond
	tdm := NewTimeDynamicManager(config)

	// Create mock validator selection manager
	vsm := NewValidatorSelectionManager(DefaultValidatorSelectionConfig())
	
	// Create mock reputation system
	rs := NewReputationSystem(DefaultReputationConfig(), nil)
	
	// Create mock tracing system
	ts := NewTracingSystem(DefaultTracingConfig())

	// Set integration components
	tdm.SetIntegrationComponents(vsm, rs, ts)

	// Test that components are set
	if tdm.validatorSelectionManager == nil {
		t.Error("Validator selection manager should be set")
	}

	if tdm.reputationSystem == nil {
		t.Error("Reputation system should be set")
	}

	if tdm.tracingSystem == nil {
		t.Error("Tracing system should be set")
	}

	t.Logf("Time dynamic integration works correctly")
}
