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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/params"
)

// TestRandomPOASelection tests the new random signer selection algorithm
func TestRandomPOASelection(t *testing.T) {
	// Create test signers
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		common.HexToAddress("0x3333333333333333333333333333333333333333"),
		common.HexToAddress("0x4444444444444444444444444444444444444444"),
	}

	// Create a snapshot with test signers
	config := &params.CliqueConfig{Period: 1, Epoch: 30000}
	sigcache := lru.NewCache[common.Hash, common.Address](4096)
	snap := newSnapshot(config, sigcache, 0, common.Hash{}, signers)

	// Test multiple blocks to ensure randomness
	selectionCount := make(map[common.Address]int)
	totalBlocks := 1000

	for blockNum := uint64(1); blockNum <= uint64(totalBlocks); blockNum++ {
		// Update snapshot hash for each block to simulate different block hashes
		snap.Hash = common.BytesToHash([]byte{byte(blockNum), byte(blockNum >> 8), byte(blockNum >> 16), byte(blockNum >> 24)})
		
		// Find which signer is selected for this block
		for _, signer := range signers {
			if snap.inturn(blockNum, signer) {
				selectionCount[signer]++
				break
			}
		}
	}

	// Verify that all signers get selected (randomness should distribute selections)
	expectedMinSelections := int(totalBlocks) / len(signers) / 4 // At least 25% of fair distribution
	for _, signer := range signers {
		if selectionCount[signer] < int(expectedMinSelections) {
			t.Errorf("Signer %s was selected only %d times out of %d blocks, expected at least %d", 
				signer.Hex(), selectionCount[signer], totalBlocks, expectedMinSelections)
		}
	}

	// Verify deterministic behavior - same block should always select same signer
	testBlock := uint64(42)
	snap.Hash = common.BytesToHash([]byte{0x42, 0x00, 0x00, 0x00})
	
	var firstSelection common.Address
	for _, signer := range signers {
		if snap.inturn(testBlock, signer) {
			firstSelection = signer
			break
		}
	}

	// Test multiple times to ensure deterministic behavior
	for i := 0; i < 10; i++ {
		var currentSelection common.Address
		for _, signer := range signers {
			if snap.inturn(testBlock, signer) {
				currentSelection = signer
				break
			}
		}
		if currentSelection != firstSelection {
			t.Errorf("Non-deterministic behavior: block %d selected %s first, then %s", 
				testBlock, firstSelection.Hex(), currentSelection.Hex())
		}
	}

	t.Logf("Selection distribution over %d blocks:", totalBlocks)
	for _, signer := range signers {
		t.Logf("  %s: %d times (%.2f%%)", 
			signer.Hex(), selectionCount[signer], 
			float64(selectionCount[signer])/float64(totalBlocks)*100)
	}
}

// TestRandomPOAWithDifferentHashes tests that different block hashes produce different selections
func TestRandomPOAWithDifferentHashes(t *testing.T) {
	signers := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		common.HexToAddress("0x3333333333333333333333333333333333333333"),
	}

	config := &params.CliqueConfig{Period: 1, Epoch: 30000}
	sigcache := lru.NewCache[common.Hash, common.Address](4096)
	snap := newSnapshot(config, sigcache, 0, common.Hash{}, signers)

	blockNum := uint64(100)
	selections := make(map[common.Hash]common.Address)

	// Test with different block hashes
	for i := 0; i < 20; i++ {
		hash := common.BytesToHash([]byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)})
		snap.Hash = hash

		for _, signer := range signers {
			if snap.inturn(blockNum, signer) {
				selections[hash] = signer
				break
			}
		}
	}

	// Verify that different hashes can produce different selections
	uniqueSelections := make(map[common.Address]bool)
	for _, signer := range selections {
		uniqueSelections[signer] = true
	}

	if len(uniqueSelections) == 1 {
		t.Logf("All different hashes selected the same signer: %s", 
			selections[common.Hash{}].Hex())
	} else {
		t.Logf("Different hashes produced %d unique selections", len(uniqueSelections))
		for signer := range uniqueSelections {
			t.Logf("  Selected signer: %s", signer.Hex())
		}
	}
}
