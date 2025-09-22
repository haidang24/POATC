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
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// TimeDynamicConfig contains configuration for time dynamic mechanisms
type TimeDynamicConfig struct {
	// Dynamic Block Time
	EnableDynamicBlockTime bool          `json:"enable_dynamic_block_time"`
	BaseBlockTime         time.Duration `json:"base_block_time"`         // 15 seconds
	MinBlockTime          time.Duration `json:"min_block_time"`          // 5 seconds
	MaxBlockTime          time.Duration `json:"max_block_time"`          // 30 seconds
	TxThresholdHigh       int           `json:"tx_threshold_high"`       // High transaction threshold
	TxThresholdLow        int           `json:"tx_threshold_low"`        // Low transaction threshold
	
	// Dynamic Validator Selection
	EnableDynamicValidatorSelection bool          `json:"enable_dynamic_validator_selection"`
	ValidatorSelectionInterval      time.Duration `json:"validator_selection_interval"` // 10 minutes
	
	// Dynamic Reputation Decay
	EnableDynamicReputationDecay bool          `json:"enable_dynamic_reputation_decay"`
	ReputationDecayRate          float64       `json:"reputation_decay_rate"`          // Per hour decay rate
	ReputationUpdateInterval     time.Duration `json:"reputation_update_interval"`     // Real-time update interval
}

// DefaultTimeDynamicConfig returns a default configuration
func DefaultTimeDynamicConfig() *TimeDynamicConfig {
	return &TimeDynamicConfig{
		// Dynamic Block Time
		EnableDynamicBlockTime: true,
		BaseBlockTime:         15 * time.Second,
		MinBlockTime:          5 * time.Second,
		MaxBlockTime:          20 * time.Second, // Reduced from 30s to 20s for better block production
		TxThresholdHigh:       100, // High transaction count
		TxThresholdLow:        5,   // Reduced from 10 to 5 for better responsiveness
		
		// Dynamic Validator Selection
		EnableDynamicValidatorSelection: true,
		ValidatorSelectionInterval:      10 * time.Minute,
		
		// Dynamic Reputation Decay
		EnableDynamicReputationDecay: true,
		ReputationDecayRate:          0.05, // 5% decay per hour
		ReputationUpdateInterval:     1 * time.Minute,
	}
}

// TimeDynamicManager manages all time dynamic mechanisms
type TimeDynamicManager struct {
	config                  *TimeDynamicConfig
	mutex                   sync.RWMutex
	
	// Dynamic Block Time
	currentBlockTime        time.Duration
	recentTxCounts          []int
	lastBlockTimeUpdate     time.Time
	
	// Dynamic Validator Selection
	lastValidatorSelection  time.Time
	validatorSelectionCount int
	
	// Dynamic Reputation Decay
	lastReputationDecay     time.Time
	decayHistory            []DecayRecord
	
	// Integration components
	validatorSelectionManager *ValidatorSelectionManager
	reputationSystem          *ReputationSystem
	tracingSystem             *TracingSystem
}

// DecayRecord tracks reputation decay events
type DecayRecord struct {
	Timestamp   time.Time      `json:"timestamp"`
	Address     common.Address `json:"address"`
	OldScore    float64        `json:"old_score"`
	NewScore    float64        `json:"new_score"`
	DecayAmount float64        `json:"decay_amount"`
}

// TransactionMetrics tracks transaction metrics for dynamic block time
type TransactionMetrics struct {
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// NewTimeDynamicManager creates a new time dynamic manager
func NewTimeDynamicManager(config *TimeDynamicConfig) *TimeDynamicManager {
	if config == nil {
		config = DefaultTimeDynamicConfig()
	}
	
	return &TimeDynamicManager{
		config:                  config,
		currentBlockTime:        config.BaseBlockTime,
		recentTxCounts:          make([]int, 0),
		lastBlockTimeUpdate:     time.Now(),
		lastValidatorSelection:  time.Now(),
		validatorSelectionCount: 0,
		lastReputationDecay:     time.Now(),
		decayHistory:            make([]DecayRecord, 0),
	}
}

// SetIntegrationComponents sets the integration components
func (tdm *TimeDynamicManager) SetIntegrationComponents(
	vsm *ValidatorSelectionManager,
	rs *ReputationSystem,
	ts *TracingSystem,
) {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()
	
	tdm.validatorSelectionManager = vsm
	tdm.reputationSystem = rs
	tdm.tracingSystem = ts
}

// ===== Dynamic Block Time =====

// UpdateTransactionCount updates the transaction count for dynamic block time calculation
func (tdm *TimeDynamicManager) UpdateTransactionCount(txCount int) {
	if !tdm.config.EnableDynamicBlockTime {
		return
	}
	
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()
	
	// Add current transaction count
	tdm.recentTxCounts = append(tdm.recentTxCounts, txCount)
	
	// Keep only last 10 measurements
	if len(tdm.recentTxCounts) > 10 {
		tdm.recentTxCounts = tdm.recentTxCounts[len(tdm.recentTxCounts)-10:]
	}
	
	// Update block time if enough data
	if len(tdm.recentTxCounts) >= 3 {
		tdm.calculateDynamicBlockTime()
	}
}

// calculateDynamicBlockTime calculates the dynamic block time based on transaction volume
func (tdm *TimeDynamicManager) calculateDynamicBlockTime() {
	// Calculate average transaction count
	totalTx := 0
	for _, count := range tdm.recentTxCounts {
		totalTx += count
	}
	avgTxCount := float64(totalTx) / float64(len(tdm.recentTxCounts))
	
	var newBlockTime time.Duration
	
	if avgTxCount >= float64(tdm.config.TxThresholdHigh) {
		// High transaction volume -> shorter block time
		ratio := math.Min(avgTxCount/float64(tdm.config.TxThresholdHigh), 3.0) // Max 3x speedup
		newBlockTime = time.Duration(float64(tdm.config.BaseBlockTime) / ratio)
		if newBlockTime < tdm.config.MinBlockTime {
			newBlockTime = tdm.config.MinBlockTime
		}
	} else if avgTxCount <= float64(tdm.config.TxThresholdLow) {
		// Low/No transaction volume -> slightly longer block time, but not too much
		if avgTxCount == 0 {
			// No transactions: use base block time to ensure continuous block production
			newBlockTime = tdm.config.BaseBlockTime
		} else {
			// Very low transactions: moderate increase
			ratio := math.Min(float64(tdm.config.TxThresholdLow)/avgTxCount, 1.33) // Max 33% increase
			newBlockTime = time.Duration(float64(tdm.config.BaseBlockTime) * ratio)
			if newBlockTime > tdm.config.MaxBlockTime {
				newBlockTime = tdm.config.MaxBlockTime
			}
		}
	} else {
		// Normal transaction volume -> base block time with smooth transition
		highRatio := (avgTxCount - float64(tdm.config.TxThresholdLow)) / 
					 (float64(tdm.config.TxThresholdHigh) - float64(tdm.config.TxThresholdLow))
		newBlockTime = time.Duration(float64(tdm.config.BaseBlockTime) * (1.2 - 0.2*highRatio)) // Reduced range for stability
	}
	
	// Update if changed significantly
	if math.Abs(float64(newBlockTime-tdm.currentBlockTime)) > float64(time.Second) {
		oldBlockTime := tdm.currentBlockTime
		tdm.currentBlockTime = newBlockTime
		tdm.lastBlockTimeUpdate = time.Now()
		
		log.Info("Dynamic block time updated",
			"old_time", oldBlockTime,
			"new_time", newBlockTime,
			"avg_tx_count", avgTxCount,
			"threshold_high", tdm.config.TxThresholdHigh,
			"threshold_low", tdm.config.TxThresholdLow)
		
		// Trace the event
		if tdm.tracingSystem != nil {
			tdm.tracingSystem.Trace(TraceEventTimeDynamic, TraceLevelDetailed, 0, common.Address{},
				"Dynamic block time updated",
				map[string]interface{}{
					"old_block_time":    oldBlockTime.Seconds(),
					"new_block_time":    newBlockTime.Seconds(),
					"avg_tx_count":      avgTxCount,
					"threshold_high":    tdm.config.TxThresholdHigh,
					"threshold_low":     tdm.config.TxThresholdLow,
					"change_reason":     tdm.getBlockTimeChangeReason(avgTxCount),
				})
		}
	}
}

// getBlockTimeChangeReason returns the reason for block time change
func (tdm *TimeDynamicManager) getBlockTimeChangeReason(avgTxCount float64) string {
	if avgTxCount >= float64(tdm.config.TxThresholdHigh) {
		return "high_transaction_volume"
	} else if avgTxCount == 0 {
		return "no_transactions"
	} else if avgTxCount <= float64(tdm.config.TxThresholdLow) {
		return "low_transaction_volume"
	}
	return "normal_transaction_volume"
}

// GetCurrentBlockTime returns the current dynamic block time
func (tdm *TimeDynamicManager) GetCurrentBlockTime() time.Duration {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	if !tdm.config.EnableDynamicBlockTime {
		return tdm.config.BaseBlockTime
	}
	
	return tdm.currentBlockTime
}

// ===== Dynamic Validator Selection =====

// ShouldUpdateValidatorSelection checks if it's time to update validator selection
func (tdm *TimeDynamicManager) ShouldUpdateValidatorSelection() bool {
	if !tdm.config.EnableDynamicValidatorSelection {
		return false
	}
	
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	return time.Since(tdm.lastValidatorSelection) >= tdm.config.ValidatorSelectionInterval
}

// UpdateValidatorSelection triggers a validator selection update
func (tdm *TimeDynamicManager) UpdateValidatorSelection(blockNumber uint64, blockHash common.Hash) error {
	if !tdm.config.EnableDynamicValidatorSelection || tdm.validatorSelectionManager == nil {
		return nil
	}
	
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()
	
	// Trigger validator selection update
	_, err := tdm.validatorSelectionManager.SelectSmallValidatorSet(blockNumber, blockHash)
	if err != nil {
		log.Error("Failed to update validator selection", "error", err)
		return err
	}
	
	tdm.lastValidatorSelection = time.Now()
	tdm.validatorSelectionCount++
	
	log.Info("Dynamic validator selection updated",
		"block_number", blockNumber,
		"selection_count", tdm.validatorSelectionCount,
		"interval", tdm.config.ValidatorSelectionInterval)
	
	// Trace the event
	if tdm.tracingSystem != nil {
		tdm.tracingSystem.Trace(TraceEventTimeDynamic, TraceLevelBasic, blockNumber, common.Address{},
			"Dynamic validator selection updated",
			map[string]interface{}{
				"selection_count": tdm.validatorSelectionCount,
				"interval":        tdm.config.ValidatorSelectionInterval.String(),
				"block_number":    blockNumber,
			})
	}
	
	return nil
}

// ===== Dynamic Reputation Decay =====

// ShouldApplyReputationDecay checks if it's time to apply reputation decay
func (tdm *TimeDynamicManager) ShouldApplyReputationDecay() bool {
	if !tdm.config.EnableDynamicReputationDecay {
		return false
	}
	
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	return time.Since(tdm.lastReputationDecay) >= tdm.config.ReputationUpdateInterval
}

// ApplyReputationDecay applies real-time reputation decay
func (tdm *TimeDynamicManager) ApplyReputationDecay() error {
	if !tdm.config.EnableDynamicReputationDecay || tdm.reputationSystem == nil {
		return nil
	}
	
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()
	
	now := time.Now()
	timeSinceLastDecay := now.Sub(tdm.lastReputationDecay)
	
	// Calculate decay factor based on time elapsed
	hoursElapsed := timeSinceLastDecay.Hours()
	decayFactor := 1.0 - (tdm.config.ReputationDecayRate * hoursElapsed)
	
	if decayFactor < 0.5 {
		decayFactor = 0.5 // Minimum 50% retention
	}
	
	// Apply decay to all validators
	validators := tdm.reputationSystem.GetAllValidators()
	decayCount := 0
	
	for _, validator := range validators {
		score := tdm.reputationSystem.GetReputationScore(validator)
		if score != nil && score.CurrentScore > 0 {
			oldScore := score.CurrentScore
			newScore := oldScore * decayFactor
			
			// Update the score
			tdm.reputationSystem.ApplyDecay(validator, 1.0-decayFactor)
			
			// Record decay
			decayRecord := DecayRecord{
				Timestamp:   now,
				Address:     validator,
				OldScore:    oldScore,
				NewScore:    newScore,
				DecayAmount: oldScore - newScore,
			}
			tdm.decayHistory = append(tdm.decayHistory, decayRecord)
			decayCount++
		}
	}
	
	// Keep only recent decay history (last 100 records)
	if len(tdm.decayHistory) > 100 {
		tdm.decayHistory = tdm.decayHistory[len(tdm.decayHistory)-100:]
	}
	
	tdm.lastReputationDecay = now
	
	log.Info("Dynamic reputation decay applied",
		"decay_factor", decayFactor,
		"hours_elapsed", hoursElapsed,
		"validators_affected", decayCount,
		"decay_rate", tdm.config.ReputationDecayRate)
	
	// Trace the event
	if tdm.tracingSystem != nil {
		tdm.tracingSystem.Trace(TraceEventTimeDynamic, TraceLevelDetailed, 0, common.Address{},
			"Dynamic reputation decay applied",
			map[string]interface{}{
				"decay_factor":        decayFactor,
				"hours_elapsed":       hoursElapsed,
				"validators_affected": decayCount,
				"decay_rate":          tdm.config.ReputationDecayRate,
			})
	}
	
	return nil
}

// ===== Statistics and Monitoring =====

// GetTimeDynamicStats returns statistics about time dynamic mechanisms
func (tdm *TimeDynamicManager) GetTimeDynamicStats() map[string]interface{} {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"config": map[string]interface{}{
			"enable_dynamic_block_time":            tdm.config.EnableDynamicBlockTime,
			"enable_dynamic_validator_selection":   tdm.config.EnableDynamicValidatorSelection,
			"enable_dynamic_reputation_decay":      tdm.config.EnableDynamicReputationDecay,
		},
		"dynamic_block_time": map[string]interface{}{
			"current_block_time":    tdm.currentBlockTime.Seconds(),
			"base_block_time":       tdm.config.BaseBlockTime.Seconds(),
			"min_block_time":        tdm.config.MinBlockTime.Seconds(),
			"max_block_time":        tdm.config.MaxBlockTime.Seconds(),
			"last_update":           tdm.lastBlockTimeUpdate,
			"recent_tx_count":       len(tdm.recentTxCounts),
		},
		"dynamic_validator_selection": map[string]interface{}{
			"selection_count":     tdm.validatorSelectionCount,
			"last_selection":      tdm.lastValidatorSelection,
			"selection_interval":  tdm.config.ValidatorSelectionInterval.String(),
		},
		"dynamic_reputation_decay": map[string]interface{}{
			"last_decay":          tdm.lastReputationDecay,
			"decay_rate":          tdm.config.ReputationDecayRate,
			"update_interval":     tdm.config.ReputationUpdateInterval.String(),
			"decay_history_count": len(tdm.decayHistory),
		},
	}
	
	// Add recent transaction counts if available
	if len(tdm.recentTxCounts) > 0 {
		totalTx := 0
		for _, count := range tdm.recentTxCounts {
			totalTx += count
		}
		stats["dynamic_block_time"].(map[string]interface{})["avg_tx_count"] = float64(totalTx) / float64(len(tdm.recentTxCounts))
		stats["dynamic_block_time"].(map[string]interface{})["recent_tx_counts"] = tdm.recentTxCounts
	}
	
	return stats
}

// GetDecayHistory returns the recent decay history
func (tdm *TimeDynamicManager) GetDecayHistory(limit int) []DecayRecord {
	tdm.mutex.RLock()
	defer tdm.mutex.RUnlock()
	
	if limit <= 0 || limit > len(tdm.decayHistory) {
		limit = len(tdm.decayHistory)
	}
	
	start := len(tdm.decayHistory) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]DecayRecord, limit)
	copy(result, tdm.decayHistory[start:])
	return result
}

// UpdateConfig updates the time dynamic configuration
func (tdm *TimeDynamicManager) UpdateConfig(config *TimeDynamicConfig) {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()
	
	if config != nil {
		tdm.config = config
		log.Info("Time dynamic configuration updated",
			"enable_dynamic_block_time", config.EnableDynamicBlockTime,
			"enable_dynamic_validator_selection", config.EnableDynamicValidatorSelection,
			"enable_dynamic_reputation_decay", config.EnableDynamicReputationDecay)
	}
}
