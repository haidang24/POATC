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
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// ValidatorSelectionConfig contains configuration for validator selection
type ValidatorSelectionConfig struct {
	EnableValidatorSelection bool          // Enable validator selection mechanism
	SmallValidatorSetSize    int           // Size of small validator set
	SelectionWindow          time.Duration // Time window for validator selection
	SelectionMethod          string        // "random", "stake", "reputation", "hybrid"
	StakeWeight              float64       // Weight for stake-based selection (0.0-1.0)
	ReputationWeight         float64       // Weight for reputation-based selection (0.0-1.0)
	RandomWeight             float64       // Weight for random selection (0.0-1.0)
}

// DefaultValidatorSelectionConfig returns a default configuration
func DefaultValidatorSelectionConfig() *ValidatorSelectionConfig {
	return &ValidatorSelectionConfig{
		EnableValidatorSelection: true,
		SmallValidatorSetSize:    3, // Select 3 validators from all validators
		SelectionWindow:          1 * time.Hour,
		SelectionMethod:          "hybrid", // Use hybrid selection
		StakeWeight:              0.4,      // 40% stake weight
		ReputationWeight:         0.3,      // 30% reputation weight
		RandomWeight:             0.3,      // 30% random weight
	}
}

// ValidatorInfo contains information about a validator
type ValidatorInfo struct {
	Address     common.Address `json:"address"`
	Stake       *big.Int       `json:"stake"`
	Reputation  float64        `json:"reputation"`
	LastActive  time.Time      `json:"last_active"`
	BlocksMined int            `json:"blocks_mined"`
	IsActive    bool           `json:"is_active"`
}

// ValidatorSelectionManager handles validator selection logic
type ValidatorSelectionManager struct {
	config           *ValidatorSelectionConfig
	allValidators    map[common.Address]*ValidatorInfo
	smallValidatorSet []common.Address
	lastSelection    time.Time
	selectionHistory []ValidatorSelectionRecord
	tracingSystem    *TracingSystem
}

// ValidatorSelectionRecord records a validator selection event
type ValidatorSelectionRecord struct {
	BlockNumber      uint64            `json:"block_number"`
	Timestamp        time.Time         `json:"timestamp"`
	SelectedValidators []common.Address `json:"selected_validators"`
	SelectionMethod  string            `json:"selection_method"`
	SelectionSeed    []byte            `json:"selection_seed"`
}

// NewValidatorSelectionManager creates a new validator selection manager
func NewValidatorSelectionManager(config *ValidatorSelectionConfig) *ValidatorSelectionManager {
	if config == nil {
		config = DefaultValidatorSelectionConfig()
	}

	return &ValidatorSelectionManager{
		config:           config,
		allValidators:    make(map[common.Address]*ValidatorInfo),
		smallValidatorSet: make([]common.Address, 0),
		selectionHistory: make([]ValidatorSelectionRecord, 0),
	}
}

// AddValidator adds a validator to the system
func (vsm *ValidatorSelectionManager) AddValidator(address common.Address, stake *big.Int, reputation float64) {
	vsm.allValidators[address] = &ValidatorInfo{
		Address:     address,
		Stake:       stake,
		Reputation:  reputation,
		LastActive:  time.Now(),
		BlocksMined: 0,
		IsActive:    true,
	}
	log.Info("Validator added to selection system", "address", address.Hex(), "stake", stake, "reputation", reputation)
}

// UpdateValidatorStake updates a validator's stake
func (vsm *ValidatorSelectionManager) UpdateValidatorStake(address common.Address, stake *big.Int) {
	if validator, exists := vsm.allValidators[address]; exists {
		validator.Stake = stake
		log.Debug("Validator stake updated", "address", address.Hex(), "stake", stake)
	}
}

// UpdateValidatorReputation updates a validator's reputation
func (vsm *ValidatorSelectionManager) UpdateValidatorReputation(address common.Address, reputation float64) {
	if validator, exists := vsm.allValidators[address]; exists {
		validator.Reputation = reputation
		log.Debug("Validator reputation updated", "address", address.Hex(), "reputation", reputation)
	}
}

// RecordBlockMining records that a validator mined a block
func (vsm *ValidatorSelectionManager) RecordBlockMining(address common.Address, blockNumber uint64) {
	if validator, exists := vsm.allValidators[address]; exists {
		validator.BlocksMined++
		validator.LastActive = time.Now()
		log.Debug("Block mining recorded", "address", address.Hex(), "block", blockNumber, "total_blocks", validator.BlocksMined)
	}
}

// SelectSmallValidatorSet selects a small set of validators from all validators
func (vsm *ValidatorSelectionManager) SelectSmallValidatorSet(blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if !vsm.config.EnableValidatorSelection {
		// If validator selection is disabled, return all validators
		allValidators := make([]common.Address, 0, len(vsm.allValidators))
		for addr := range vsm.allValidators {
			allValidators = append(allValidators, addr)
		}
		return allValidators, nil
	}

	// Check if we need to reselect (based on time window or block interval)
	if time.Since(vsm.lastSelection) < vsm.config.SelectionWindow && len(vsm.smallValidatorSet) > 0 {
		log.Debug("Using existing small validator set", "size", len(vsm.smallValidatorSet))
		return vsm.smallValidatorSet, nil
	}

	// Get active validators
	activeValidators := vsm.getActiveValidators()
	if len(activeValidators) == 0 {
		return nil, fmt.Errorf("no active validators available")
	}

	// Determine selection size
	selectionSize := vsm.config.SmallValidatorSetSize
	if selectionSize > len(activeValidators) {
		selectionSize = len(activeValidators)
	}

	// Select validators based on configured method
	var selectedValidators []common.Address
	var err error

	switch vsm.config.SelectionMethod {
	case "random":
		selectedValidators, err = vsm.selectRandomValidators(activeValidators, selectionSize, blockNumber, blockHash)
	case "stake":
		selectedValidators, err = vsm.selectStakeBasedValidators(activeValidators, selectionSize, blockNumber, blockHash)
	case "reputation":
		selectedValidators, err = vsm.selectReputationBasedValidators(activeValidators, selectionSize, blockNumber, blockHash)
	case "hybrid":
		selectedValidators, err = vsm.selectHybridValidators(activeValidators, selectionSize, blockNumber, blockHash)
	default:
		return nil, fmt.Errorf("unknown selection method: %s", vsm.config.SelectionMethod)
	}

	if err != nil {
		return nil, err
	}

	// Update small validator set
	vsm.smallValidatorSet = selectedValidators
	vsm.lastSelection = time.Now()

	// Record selection
	selectionRecord := ValidatorSelectionRecord{
		BlockNumber:       blockNumber,
		Timestamp:         time.Now(),
		SelectedValidators: selectedValidators,
		SelectionMethod:   vsm.config.SelectionMethod,
		SelectionSeed:     vsm.generateSelectionSeed(blockNumber, blockHash),
	}
	vsm.selectionHistory = append(vsm.selectionHistory, selectionRecord)

	// Keep only recent history
	if len(vsm.selectionHistory) > 100 {
		vsm.selectionHistory = vsm.selectionHistory[len(vsm.selectionHistory)-100:]
	}

	log.Info("Small validator set selected", 
		"method", vsm.config.SelectionMethod,
		"size", len(selectedValidators),
		"block", blockNumber,
		"validators", selectedValidators)

	return selectedValidators, nil
}

// getActiveValidators returns list of active validators
func (vsm *ValidatorSelectionManager) getActiveValidators() []common.Address {
	var activeValidators []common.Address
	for addr, validator := range vsm.allValidators {
		if validator.IsActive {
			activeValidators = append(activeValidators, addr)
		}
	}
	return activeValidators
}

// selectRandomValidators selects validators randomly
func (vsm *ValidatorSelectionManager) selectRandomValidators(validators []common.Address, count int, blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if count >= len(validators) {
		return validators, nil
	}

	// Use block hash and number as seed for deterministic randomness
	seed := vsm.generateSelectionSeed(blockNumber, blockHash)
	
	// Simple random selection using the seed
	selected := make([]common.Address, 0, count)
	used := make(map[common.Address]bool)
	
	for len(selected) < count {
		// Use seed to select index
		index := int(seed[len(selected)%len(seed)]) % len(validators)
		validator := validators[index]
		
		if !used[validator] {
			selected = append(selected, validator)
			used[validator] = true
		}
	}

	return selected, nil
}

// selectStakeBasedValidators selects validators based on stake
func (vsm *ValidatorSelectionManager) selectStakeBasedValidators(validators []common.Address, count int, blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if count >= len(validators) {
		return validators, nil
	}

	// Calculate total stake
	totalStake := big.NewInt(0)
	for _, addr := range validators {
		if validator, exists := vsm.allValidators[addr]; exists {
			totalStake.Add(totalStake, validator.Stake)
		}
	}

	if totalStake.Cmp(big.NewInt(0)) == 0 {
		// If no stake, fall back to random selection
		return vsm.selectRandomValidators(validators, count, blockNumber, blockHash)
	}

	// Weighted selection based on stake
	selected := make([]common.Address, 0, count)
	used := make(map[common.Address]bool)
	seed := vsm.generateSelectionSeed(blockNumber, blockHash)

	for len(selected) < count {
		// Calculate cumulative weights
		cumulative := big.NewInt(0)
		target := new(big.Int).SetBytes(seed[len(selected)%len(seed):])
		target.Mod(target, totalStake)

		for _, addr := range validators {
			if used[addr] {
				continue
			}

			if validator, exists := vsm.allValidators[addr]; exists {
				cumulative.Add(cumulative, validator.Stake)
				if cumulative.Cmp(target) >= 0 {
					selected = append(selected, addr)
					used[addr] = true
					break
				}
			}
		}
	}

	return selected, nil
}

// selectReputationBasedValidators selects validators based on reputation
func (vsm *ValidatorSelectionManager) selectReputationBasedValidators(validators []common.Address, count int, blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if count >= len(validators) {
		return validators, nil
	}

	// Calculate total reputation
	totalReputation := 0.0
	for _, addr := range validators {
		if validator, exists := vsm.allValidators[addr]; exists {
			totalReputation += validator.Reputation
		}
	}

	if totalReputation == 0 {
		// If no reputation, fall back to random selection
		return vsm.selectRandomValidators(validators, count, blockNumber, blockHash)
	}

	// Weighted selection based on reputation
	selected := make([]common.Address, 0, count)
	used := make(map[common.Address]bool)
	seed := vsm.generateSelectionSeed(blockNumber, blockHash)

	for len(selected) < count {
		// Calculate cumulative weights
		cumulative := 0.0
		target := float64(seed[len(selected)%len(seed)]) / 255.0 * totalReputation

		for _, addr := range validators {
			if used[addr] {
				continue
			}

			if validator, exists := vsm.allValidators[addr]; exists {
				cumulative += validator.Reputation
				if cumulative >= target {
					selected = append(selected, addr)
					used[addr] = true
					break
				}
			}
		}
	}

	return selected, nil
}

// selectHybridValidators selects validators using hybrid method (stake + reputation + random)
func (vsm *ValidatorSelectionManager) selectHybridValidators(validators []common.Address, count int, blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if count >= len(validators) {
		return validators, nil
	}

	// Calculate hybrid scores for each validator
	scores := make(map[common.Address]float64)
	
	// Normalize stake and reputation
	maxStake := big.NewInt(0)
	maxReputation := 0.0
	
	for _, addr := range validators {
		if validator, exists := vsm.allValidators[addr]; exists {
			if validator.Stake.Cmp(maxStake) > 0 {
				maxStake.Set(validator.Stake)
			}
			if validator.Reputation > maxReputation {
				maxReputation = validator.Reputation
			}
		}
	}

	// Calculate hybrid scores
	for _, addr := range validators {
		if validator, exists := vsm.allValidators[addr]; exists {
			stakeScore := 0.0
			if maxStake.Cmp(big.NewInt(0)) > 0 {
				stakeScore = float64(validator.Stake.Uint64()) / float64(maxStake.Uint64())
			}
			
			reputationScore := 0.0
			if maxReputation > 0 {
				reputationScore = validator.Reputation / maxReputation
			}
			
			// Calculate hybrid score
			hybridScore := vsm.config.StakeWeight*stakeScore + 
						   vsm.config.ReputationWeight*reputationScore + 
						   vsm.config.RandomWeight*0.5 // Random component
			
			scores[addr] = hybridScore
		}
	}

	// Weighted selection based on hybrid scores
	selected := make([]common.Address, 0, count)
	used := make(map[common.Address]bool)
	seed := vsm.generateSelectionSeed(blockNumber, blockHash)

	// Calculate total score
	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	if totalScore == 0 {
		// If no scores, fall back to random selection
		return vsm.selectRandomValidators(validators, count, blockNumber, blockHash)
	}

	for len(selected) < count {
		// Calculate cumulative weights
		cumulative := 0.0
		target := float64(seed[len(selected)%len(seed)]) / 255.0 * totalScore
		
		selectedInThisRound := false

		for _, addr := range validators {
			if used[addr] {
				continue
			}

			if score, exists := scores[addr]; exists {
				cumulative += score
				if cumulative >= target {
					selected = append(selected, addr)
					used[addr] = true
					selectedInThisRound = true
					break
				}
			}
		}
		
		// If no validator was selected in this round, break to avoid infinite loop
		if !selectedInThisRound {
			break
		}
	}

	return selected, nil
}

// generateSelectionSeed generates a deterministic seed for selection
func (vsm *ValidatorSelectionManager) generateSelectionSeed(blockNumber uint64, blockHash common.Hash) []byte {
	// Combine block number and hash for seed
	data := make([]byte, 40)
	
	// Add block number (8 bytes)
	for i := 0; i < 8; i++ {
		data[i] = byte(blockNumber >> (i * 8))
	}
	
	// Add block hash (32 bytes)
	copy(data[8:], blockHash[:])
	
	// Hash the combined data
	hash := sha256.Sum256(data)
	return hash[:]
}

// GetSmallValidatorSet returns the current small validator set
func (vsm *ValidatorSelectionManager) GetSmallValidatorSet() []common.Address {
	return vsm.smallValidatorSet
}

// GetValidatorInfo returns information about a validator
func (vsm *ValidatorSelectionManager) GetValidatorInfo(address common.Address) *ValidatorInfo {
	return vsm.allValidators[address]
}

// GetSelectionHistory returns the selection history
func (vsm *ValidatorSelectionManager) GetSelectionHistory() []ValidatorSelectionRecord {
	return vsm.selectionHistory
}

// GetStats returns statistics about validator selection
func (vsm *ValidatorSelectionManager) GetStats() map[string]interface{} {
	activeCount := 0
	totalStake := big.NewInt(0)
	totalReputation := 0.0
	totalBlocks := 0

	for _, validator := range vsm.allValidators {
		if validator.IsActive {
			activeCount++
			totalStake.Add(totalStake, validator.Stake)
			totalReputation += validator.Reputation
			totalBlocks += validator.BlocksMined
		}
	}

	return map[string]interface{}{
		"config": map[string]interface{}{
			"enable_validator_selection": vsm.config.EnableValidatorSelection,
			"small_validator_set_size":   vsm.config.SmallValidatorSetSize,
			"selection_method":           vsm.config.SelectionMethod,
			"selection_window":           vsm.config.SelectionWindow.String(),
		},
		"validators": map[string]interface{}{
			"total":          len(vsm.allValidators),
			"active":         activeCount,
			"small_set_size": len(vsm.smallValidatorSet),
		},
		"totals": map[string]interface{}{
			"stake":      totalStake.String(),
			"reputation": totalReputation,
			"blocks":     totalBlocks,
		},
		"selection": map[string]interface{}{
			"last_selection":    vsm.lastSelection,
			"history_count":     len(vsm.selectionHistory),
			"current_set":       vsm.smallValidatorSet,
		},
	}
}

// UpdateConfig updates the validator selection configuration
func (vsm *ValidatorSelectionManager) UpdateConfig(config *ValidatorSelectionConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	vsm.config = config
	log.Info("Validator selection configuration updated", 
		"enable", config.EnableValidatorSelection,
		"method", config.SelectionMethod,
		"set_size", config.SmallValidatorSetSize)

	return nil
}

// SetTracingSystem sets the tracing system for the validator selection manager
func (vsm *ValidatorSelectionManager) SetTracingSystem(tracingSystem *TracingSystem) {
	vsm.tracingSystem = tracingSystem
}
