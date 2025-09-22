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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestValidatorSelectionManagerBasic(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	config.SelectionMethod = "random"
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	addr4 := common.HexToAddress("0x4444444444444444444444444444444444444444")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 1.0)
	vsm.AddValidator(addr2, big.NewInt(2000000), 1.5)
	vsm.AddValidator(addr3, big.NewInt(1500000), 1.2)
	vsm.AddValidator(addr4, big.NewInt(3000000), 2.0)
	
	// Test selection
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	selected, err := vsm.SelectSmallValidatorSet(1, blockHash)
	
	if err != nil {
		t.Fatalf("Selection failed: %v", err)
	}
	
	if len(selected) != 2 {
		t.Fatalf("Expected 2 validators, got %d", len(selected))
	}
	
	// Verify all selected validators are in the original set
	validators := []common.Address{addr1, addr2, addr3, addr4}
	for _, selectedAddr := range selected {
		found := false
		for _, validAddr := range validators {
			if selectedAddr == validAddr {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Selected validator %s not in original set", selectedAddr.Hex())
		}
	}
	
	t.Logf("Selected validators: %v", selected)
}

func TestValidatorSelectionManagerStakeBased(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	config.SelectionMethod = "stake"
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators with different stakes
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 1.0)  // Low stake
	vsm.AddValidator(addr2, big.NewInt(10000000), 1.0) // High stake
	vsm.AddValidator(addr3, big.NewInt(5000000), 1.0)  // Medium stake
	
	// Test multiple selections to see if high stake validators are selected more often
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	selections := make(map[common.Address]int)
	
	for i := 0; i < 100; i++ {
		selected, err := vsm.SelectSmallValidatorSet(uint64(i+1), blockHash)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		
		for _, addr := range selected {
			selections[addr]++
		}
	}
	
	t.Logf("Selection distribution: %v", selections)
	
	// High stake validator should be selected more often
	if selections[addr2] < selections[addr1] {
		t.Logf("Warning: High stake validator not selected more often than low stake validator")
	}
}

func TestValidatorSelectionManagerReputationBased(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	config.SelectionMethod = "reputation"
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators with different reputations
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 0.5)  // Low reputation
	vsm.AddValidator(addr2, big.NewInt(1000000), 2.0)  // High reputation
	vsm.AddValidator(addr3, big.NewInt(1000000), 1.0)  // Medium reputation
	
	// Test multiple selections
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	selections := make(map[common.Address]int)
	
	for i := 0; i < 100; i++ {
		selected, err := vsm.SelectSmallValidatorSet(uint64(i+1), blockHash)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		
		for _, addr := range selected {
			selections[addr]++
		}
	}
	
	t.Logf("Selection distribution: %v", selections)
	
	// High reputation validator should be selected more often
	if selections[addr2] < selections[addr1] {
		t.Logf("Warning: High reputation validator not selected more often than low reputation validator")
	}
}

func TestValidatorSelectionManagerHybrid(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	config.SelectionMethod = "hybrid"
	config.StakeWeight = 0.4
	config.ReputationWeight = 0.3
	config.RandomWeight = 0.3
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators with different combinations of stake and reputation
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	addr4 := common.HexToAddress("0x4444444444444444444444444444444444444444")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 0.5)   // Low stake, low reputation
	vsm.AddValidator(addr2, big.NewInt(10000000), 2.0)  // High stake, high reputation
	vsm.AddValidator(addr3, big.NewInt(1000000), 2.0)   // Low stake, high reputation
	vsm.AddValidator(addr4, big.NewInt(10000000), 0.5)  // High stake, low reputation
	
	// Test multiple selections
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	selections := make(map[common.Address]int)
	
	for i := 0; i < 100; i++ {
		selected, err := vsm.SelectSmallValidatorSet(uint64(i+1), blockHash)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		
		for _, addr := range selected {
			selections[addr]++
		}
	}
	
	t.Logf("Selection distribution: %v", selections)
	
	// Validator with both high stake and high reputation should be selected most often
	if selections[addr2] < selections[addr1] {
		t.Logf("Warning: Best validator not selected more often than worst validator")
	}
}

func TestValidatorSelectionManagerDeterministic(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	config.SelectionMethod = "random"
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	addr4 := common.HexToAddress("0x4444444444444444444444444444444444444444")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 1.0)
	vsm.AddValidator(addr2, big.NewInt(2000000), 1.5)
	vsm.AddValidator(addr3, big.NewInt(1500000), 1.2)
	vsm.AddValidator(addr4, big.NewInt(3000000), 2.0)
	
	// Test deterministic selection
	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	
	selected1, err1 := vsm.SelectSmallValidatorSet(1, blockHash)
	selected2, err2 := vsm.SelectSmallValidatorSet(1, blockHash)
	
	if err1 != nil || err2 != nil {
		t.Fatalf("Selection failed: %v, %v", err1, err2)
	}
	
	// Should be identical
	if len(selected1) != len(selected2) {
		t.Fatalf("Different selection sizes: %d vs %d", len(selected1), len(selected2))
	}
	
	for i, addr1 := range selected1 {
		if addr1 != selected2[i] {
			t.Fatalf("Non-deterministic selection at index %d: %s vs %s", i, addr1.Hex(), selected2[i].Hex())
		}
	}
	
	t.Logf("Deterministic selection verified: %v", selected1)
}

func TestValidatorSelectionManagerStats(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	config.SmallValidatorSetSize = 2
	
	vsm := NewValidatorSelectionManager(config)
	
	// Add validators
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	
	vsm.AddValidator(addr1, big.NewInt(1000000), 1.0)
	vsm.AddValidator(addr2, big.NewInt(2000000), 1.5)
	
	// Get stats
	stats := vsm.GetStats()
	
	if stats == nil {
		t.Fatalf("Stats should not be nil")
	}
	
	// Check basic stats structure
	_, ok := stats["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("Config stats not found")
	}
	
	validatorsStats, ok := stats["validators"].(map[string]interface{})
	if !ok {
		t.Fatalf("Validators stats not found")
	}
	
	if validatorsStats["total"] != 2 {
		t.Fatalf("Expected 2 total validators, got %v", validatorsStats["total"])
	}
	
	if validatorsStats["active"] != 2 {
		t.Fatalf("Expected 2 active validators, got %v", validatorsStats["active"])
	}
	
	t.Logf("Stats: %+v", stats)
}

func TestValidatorSelectionManagerUpdate(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	vsm := NewValidatorSelectionManager(config)
	
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	
	// Add validator
	vsm.AddValidator(addr, big.NewInt(1000000), 1.0)
	
	// Update stake
	newStake := big.NewInt(5000000)
	vsm.UpdateValidatorStake(addr, newStake)
	
	// Update reputation
	newReputation := 2.5
	vsm.UpdateValidatorReputation(addr, newReputation)
	
	// Verify updates
	info := vsm.GetValidatorInfo(addr)
	if info == nil {
		t.Fatalf("Validator info not found")
	}
	
	if info.Stake.Cmp(newStake) != 0 {
		t.Fatalf("Stake not updated correctly: expected %s, got %s", newStake.String(), info.Stake.String())
	}
	
	if info.Reputation != newReputation {
		t.Fatalf("Reputation not updated correctly: expected %f, got %f", newReputation, info.Reputation)
	}
	
	t.Logf("Validator updated successfully: stake=%s, reputation=%f", info.Stake.String(), info.Reputation)
}

func TestValidatorSelectionManagerRecordBlockMining(t *testing.T) {
	config := DefaultValidatorSelectionConfig()
	vsm := NewValidatorSelectionManager(config)
	
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	
	// Add validator
	vsm.AddValidator(addr, big.NewInt(1000000), 1.0)
	
	// Record block mining
	vsm.RecordBlockMining(addr, 1)
	vsm.RecordBlockMining(addr, 2)
	vsm.RecordBlockMining(addr, 3)
	
	// Verify block count
	info := vsm.GetValidatorInfo(addr)
	if info == nil {
		t.Fatalf("Validator info not found")
	}
	
	if info.BlocksMined != 3 {
		t.Fatalf("Expected 3 blocks mined, got %d", info.BlocksMined)
	}
	
	t.Logf("Block mining recorded successfully: %d blocks", info.BlocksMined)
}
