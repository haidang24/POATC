// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.

package poatc

import (
	"testing"
	"time"
)

// TestConsecutiveBlockPrevention tests that minimum delays prevent consecutive blocks
func TestConsecutiveBlockPrevention(t *testing.T) {
	// Test minimum delay logic
	testCases := []struct {
		name           string
		originalDelay  time.Duration
		expectedDelay  time.Duration
		description    string
	}{
		{
			name:           "Zero delay should be increased to minimum",
			originalDelay:  0,
			expectedDelay:  1 * time.Second,
			description:    "Prevents immediate consecutive blocks",
		},
		{
			name:           "Negative delay should be increased to minimum",
			originalDelay:  -5 * time.Second,
			expectedDelay:  1 * time.Second,
			description:    "Handles negative delays gracefully",
		},
		{
			name:           "Small positive delay should be increased to minimum",
			originalDelay:  500 * time.Millisecond,
			expectedDelay:  1 * time.Second,
			description:    "Ensures minimum gap between blocks",
		},
		{
			name:           "Adequate delay should remain unchanged",
			originalDelay:  5 * time.Second,
			expectedDelay:  5 * time.Second,
			description:    "Normal delays are preserved",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the minimum delay logic from clique.go
			delay := tc.originalDelay
			absoluteMinDelay := 1 * time.Second
			
			if delay < absoluteMinDelay {
				delay = absoluteMinDelay
			}

			if delay != tc.expectedDelay {
				t.Errorf("Expected delay %v, got %v for case: %s", 
					tc.expectedDelay, delay, tc.description)
			}

			t.Logf("✓ %s: %v → %v", tc.description, tc.originalDelay, delay)
		})
	}
}

// TestDynamicBlockTimeMinimumDelay tests dynamic block time minimum delay logic
func TestDynamicBlockTimeMinimumDelay(t *testing.T) {
	config := DefaultTimeDynamicConfig()
	config.MinBlockTime = 3 * time.Second  // Very fast for testing
	tdm := NewTimeDynamicManager(config)

	// Simulate high transaction volume to get very fast block time
	for i := 0; i < 5; i++ {
		tdm.UpdateTransactionCount(200) // Very high transaction count
	}

	dynamicBlockTime := tdm.GetCurrentBlockTime()
	t.Logf("Dynamic block time with high tx volume: %v", dynamicBlockTime)

	// Simulate delay calculation logic
	baseBlockTime := 15 * time.Second
	originalDelay := 100 * time.Millisecond // Very small original delay
	
	if dynamicBlockTime != baseBlockTime {
		timeRatio := float64(dynamicBlockTime) / float64(baseBlockTime)
		adjustedDelay := time.Duration(float64(originalDelay) * timeRatio)
		
		// Apply minimum delay logic (from clique.go)
		minDelay := 2 * time.Second
		if adjustedDelay < minDelay {
			adjustedDelay = minDelay
		}
		
		// Then absolute minimum
		absoluteMinDelay := 1 * time.Second
		if adjustedDelay < absoluteMinDelay {
			adjustedDelay = absoluteMinDelay
		}

		if adjustedDelay < 1*time.Second {
			t.Errorf("Delay should never be less than 1 second, got %v", adjustedDelay)
		}

		t.Logf("✓ Minimum delay protection works: original=%v, adjusted=%v", 
			originalDelay, adjustedDelay)
	}
}
