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
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

// ReputationConfig contains configuration for the reputation system
type ReputationConfig struct {
	EnableReputationSystem bool    // Enable reputation system
	InitialReputation      float64 // Initial reputation for new validators
	MaxReputation          float64 // Maximum reputation score
	MinReputation          float64 // Minimum reputation score

	// Scoring weights
	BlockMiningWeight float64 // Weight for successful block mining
	UptimeWeight      float64 // Weight for validator uptime
	ConsistencyWeight float64 // Weight for consistent performance
	PenaltyWeight     float64 // Weight for penalties

	// Scoring parameters
	BlockMiningReward float64 // Reward for mining a block
	UptimeReward      float64 // Reward per hour of uptime
	ConsistencyReward float64 // Reward for consistent performance
	PenaltyAmount     float64 // Penalty for violations

	// Time windows
	EvaluationWindow time.Duration // Window for reputation evaluation
	UpdateInterval   time.Duration // Interval for reputation updates
	DecayFactor      float64       // Factor for reputation decay over time

	// Fairness mechanisms
	MaxComponentScore float64       // Maximum score for each component (prevents accumulation)
	ResetInterval     time.Duration // Interval for partial score reset
	NewValidatorBoost float64       // Boost for new validators
	VeteranPenalty    float64       // Penalty for very old validators

	// Thresholds
	HighReputationThreshold float64 // Threshold for high reputation
	LowReputationThreshold  float64 // Threshold for low reputation
	PenaltyThreshold        int     // Number of violations before penalty
}

// DefaultReputationConfig returns a default configuration
func DefaultReputationConfig() *ReputationConfig {
	return &ReputationConfig{
		EnableReputationSystem: true,
		InitialReputation:      1.0,
		MaxReputation:          10.0,
		MinReputation:          0.1,

		BlockMiningWeight: 0.4, // 40%
		UptimeWeight:      0.3, // 30%
		ConsistencyWeight: 0.2, // 20%
		PenaltyWeight:     0.1, // 10%

		BlockMiningReward: 0.1,
		UptimeReward:      0.05,
		ConsistencyReward: 0.08,
		PenaltyAmount:     0.5,

		EvaluationWindow: 24 * time.Hour,
		UpdateInterval:   1 * time.Hour,
		DecayFactor:      0.95, // 5% decay per update (stronger decay)

		// Fairness mechanisms
		MaxComponentScore: 5.0,                // Maximum score for each component
		ResetInterval:     7 * 24 * time.Hour, // Weekly reset
		NewValidatorBoost: 0.5,                // Boost for new validators
		VeteranPenalty:    0.1,                // Penalty for very old validators

		HighReputationThreshold: 7.0,
		LowReputationThreshold:  3.0,
		PenaltyThreshold:        3,
	}
}

// ReputationScore represents a validator's reputation score
type ReputationScore struct {
	Address          common.Address `json:"address"`
	CurrentScore     float64        `json:"current_score"`
	PreviousScore    float64        `json:"previous_score"`
	BlockMiningScore float64        `json:"block_mining_score"`
	UptimeScore      float64        `json:"uptime_score"`
	ConsistencyScore float64        `json:"consistency_score"`
	PenaltyScore     float64        `json:"penalty_score"`
	LastUpdate       time.Time      `json:"last_update"`
	LastBlockMined   uint64         `json:"last_block_mined"`
	TotalBlocksMined int            `json:"total_blocks_mined"`
	UptimeHours      float64        `json:"uptime_hours"`
	ViolationCount   int            `json:"violation_count"`
	IsActive         bool           `json:"is_active"`

	// Fairness tracking
	JoinTime       time.Time `json:"join_time"`        // When validator joined
	LastReset      time.Time `json:"last_reset"`       // Last score reset
	IsNewValidator bool      `json:"is_new_validator"` // Is this a new validator
	VeteranPenalty float64   `json:"veteran_penalty"`  // Penalty for being too old
}

// ReputationEvent represents a reputation event
type ReputationEvent struct {
	Address     common.Address `json:"address"`
	EventType   string         `json:"event_type"` // "block_mined", "uptime", "violation", "penalty"
	ScoreChange float64        `json:"score_change"`
	BlockNumber uint64         `json:"block_number"`
	Timestamp   time.Time      `json:"timestamp"`
	Description string         `json:"description"`
}

// ReputationSystem manages validator reputation scores
type ReputationSystem struct {
	config     *ReputationConfig
	db         ethdb.Database
	scores     map[common.Address]*ReputationScore
	events     []ReputationEvent
	lastUpdate time.Time
	mutex      sync.RWMutex

	// Performance tracking
	blockTimes    map[common.Address][]time.Time // Track block mining times
	uptimeTracker map[common.Address]*UptimeTracker
}

// UptimeTracker tracks validator uptime
type UptimeTracker struct {
	Address       common.Address `json:"address"`
	LastSeen      time.Time      `json:"last_seen"`
	TotalUptime   time.Duration  `json:"total_uptime"`
	IsOnline      bool           `json:"is_online"`
	OnlinePeriods []OnlinePeriod `json:"online_periods"`
}

// OnlinePeriod represents a period when validator was online
type OnlinePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewReputationSystem creates a new reputation system
func NewReputationSystem(config *ReputationConfig, db ethdb.Database) *ReputationSystem {
	if config == nil {
		config = DefaultReputationConfig()
	}

	rs := &ReputationSystem{
		config:        config,
		db:            db,
		scores:        make(map[common.Address]*ReputationScore),
		events:        make([]ReputationEvent, 0),
		blockTimes:    make(map[common.Address][]time.Time),
		uptimeTracker: make(map[common.Address]*UptimeTracker),
		lastUpdate:    time.Now(),
	}

	// Load existing data from database
	rs.loadFromDatabase()

	return rs
}

// AddValidator adds a new validator to the reputation system
func (rs *ReputationSystem) AddValidator(address common.Address) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if _, exists := rs.scores[address]; !exists {
		now := time.Now()
		rs.scores[address] = &ReputationScore{
			Address:          address,
			CurrentScore:     rs.config.InitialReputation,
			PreviousScore:    rs.config.InitialReputation,
			BlockMiningScore: 0.0,
			UptimeScore:      0.0,
			ConsistencyScore: 0.0,
			PenaltyScore:     0.0,
			LastUpdate:       now,
			IsActive:         true,

			// Fairness tracking
			JoinTime:       now,
			LastReset:      now,
			IsNewValidator: true,
			VeteranPenalty: 0.0,
		}

		rs.uptimeTracker[address] = &UptimeTracker{
			Address:       address,
			LastSeen:      now,
			TotalUptime:   0,
			IsOnline:      true,
			OnlinePeriods: make([]OnlinePeriod, 0),
		}

		rs.blockTimes[address] = make([]time.Time, 0)

		log.Info("Validator added to reputation system", "address", address.Hex(), "initial_score", rs.config.InitialReputation)
		rs.saveToDatabase()
	}
}

// RecordBlockMining records a successful block mining event
func (rs *ReputationSystem) RecordBlockMining(address common.Address, blockNumber uint64) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if score, exists := rs.scores[address]; exists {
		// Update block mining score with cap
		newBlockMiningScore := score.BlockMiningScore + rs.config.BlockMiningReward
		if newBlockMiningScore > rs.config.MaxComponentScore {
			newBlockMiningScore = rs.config.MaxComponentScore
		}
		score.BlockMiningScore = newBlockMiningScore
		score.LastBlockMined = blockNumber
		score.TotalBlocksMined++
		score.LastUpdate = time.Now()

		// Track block mining time for consistency calculation
		now := time.Now()
		rs.blockTimes[address] = append(rs.blockTimes[address], now)

		// Keep only recent block times (last 100 blocks)
		if len(rs.blockTimes[address]) > 100 {
			rs.blockTimes[address] = rs.blockTimes[address][len(rs.blockTimes[address])-100:]
		}

		// Update uptime tracker
		if tracker, exists := rs.uptimeTracker[address]; exists {
			tracker.LastSeen = now
			if !tracker.IsOnline {
				tracker.IsOnline = true
				tracker.OnlinePeriods = append(tracker.OnlinePeriods, OnlinePeriod{
					Start: now,
				})
			}
		}

		// Record event
		event := ReputationEvent{
			Address:     address,
			EventType:   "block_mined",
			ScoreChange: rs.config.BlockMiningReward,
			BlockNumber: blockNumber,
			Timestamp:   now,
			Description: fmt.Sprintf("Successfully mined block %d", blockNumber),
		}
		rs.events = append(rs.events, event)

		// Recalculate total score
		rs.calculateTotalScore(address)

		log.Debug("Block mining recorded", "address", address.Hex(), "block", blockNumber, "score_change", rs.config.BlockMiningReward)
		rs.saveToDatabase()
	}
}

// RecordViolation records a violation by a validator
func (rs *ReputationSystem) RecordViolation(address common.Address, blockNumber uint64, violationType string, description string) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if score, exists := rs.scores[address]; exists {
		score.ViolationCount++
		score.LastUpdate = time.Now()

		// Apply penalty if threshold is reached
		if score.ViolationCount >= rs.config.PenaltyThreshold {
			score.PenaltyScore += rs.config.PenaltyAmount

			// Record penalty event
			event := ReputationEvent{
				Address:     address,
				EventType:   "penalty",
				ScoreChange: -rs.config.PenaltyAmount,
				BlockNumber: blockNumber,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("Penalty applied: %s", description),
			}
			rs.events = append(rs.events, event)
		} else {
			// Record violation event
			event := ReputationEvent{
				Address:     address,
				EventType:   "violation",
				ScoreChange: 0,
				BlockNumber: blockNumber,
				Timestamp:   time.Now(),
				Description: fmt.Sprintf("Violation: %s", description),
			}
			rs.events = append(rs.events, event)
		}

		// Recalculate total score
		rs.calculateTotalScore(address)

		log.Warn("Violation recorded", "address", address.Hex(), "type", violationType, "count", score.ViolationCount)
		rs.saveToDatabase()
	}
}

// UpdateUptime updates the uptime for a validator
func (rs *ReputationSystem) UpdateUptime(address common.Address) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if tracker, exists := rs.uptimeTracker[address]; exists {
		now := time.Now()

		if tracker.IsOnline {
			// Update total uptime
			timeDiff := now.Sub(tracker.LastSeen)
			tracker.TotalUptime += timeDiff

			// Update uptime score with cap
			if score, exists := rs.scores[address]; exists {
				hours := timeDiff.Hours()
				newUptimeScore := score.UptimeScore + hours*rs.config.UptimeReward
				if newUptimeScore > rs.config.MaxComponentScore {
					newUptimeScore = rs.config.MaxComponentScore
				}
				score.UptimeScore = newUptimeScore
				score.UptimeHours += hours
				score.LastUpdate = now

				// Recalculate total score
				rs.calculateTotalScore(address)
			}
		}

		tracker.LastSeen = now
	}
}

// MarkValidatorOffline marks a validator as offline
func (rs *ReputationSystem) MarkValidatorOffline(address common.Address) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	if tracker, exists := rs.uptimeTracker[address]; exists {
		if tracker.IsOnline {
			tracker.IsOnline = false

			// Close the current online period
			if len(tracker.OnlinePeriods) > 0 {
				lastPeriod := &tracker.OnlinePeriods[len(tracker.OnlinePeriods)-1]
				if lastPeriod.End.IsZero() {
					lastPeriod.End = time.Now()
				}
			}

			log.Debug("Validator marked offline", "address", address.Hex())
		}
	}
}

// calculateTotalScore calculates the total reputation score for a validator
func (rs *ReputationSystem) calculateTotalScore(address common.Address) {
	if score, exists := rs.scores[address]; exists {
		// Calculate consistency score based on block mining intervals
		rs.calculateConsistencyScore(address)

		// Apply fairness mechanisms
		rs.applyFairnessMechanisms(address)

		// Calculate weighted total score
		totalScore := rs.config.BlockMiningWeight*score.BlockMiningScore +
			rs.config.UptimeWeight*score.UptimeScore +
			rs.config.ConsistencyWeight*score.ConsistencyScore -
			rs.config.PenaltyWeight*score.PenaltyScore

		// Check for NaN or infinite values
		if math.IsNaN(totalScore) || math.IsInf(totalScore, 0) {
			totalScore = rs.config.InitialReputation
		}

		// Apply bounds
		if totalScore > rs.config.MaxReputation {
			totalScore = rs.config.MaxReputation
		}
		if totalScore < rs.config.MinReputation {
			totalScore = rs.config.MinReputation
		}

		score.PreviousScore = score.CurrentScore
		score.CurrentScore = totalScore
	}
}

// applyFairnessMechanisms applies fairness mechanisms to prevent score accumulation
func (rs *ReputationSystem) applyFairnessMechanisms(address common.Address) {
	if score, exists := rs.scores[address]; exists {
		now := time.Now()

		// 1. Check if it's time for a partial reset
		if now.Sub(score.LastReset) >= rs.config.ResetInterval {
			rs.performPartialReset(address)
		}

		// 2. Apply new validator boost
		if score.IsNewValidator && now.Sub(score.JoinTime) < 24*time.Hour {
			// Give new validators a boost for first 24 hours
			boost := rs.config.NewValidatorBoost
			score.BlockMiningScore = math.Min(score.BlockMiningScore+boost, rs.config.MaxComponentScore)
			score.UptimeScore = math.Min(score.UptimeScore+boost, rs.config.MaxComponentScore)
		} else if now.Sub(score.JoinTime) >= 24*time.Hour {
			score.IsNewValidator = false
		}

		// 3. Apply veteran penalty for very old validators
		if now.Sub(score.JoinTime) > 30*24*time.Hour { // 30 days
			penalty := rs.config.VeteranPenalty
			score.VeteranPenalty = penalty
			score.BlockMiningScore = math.Max(score.BlockMiningScore-penalty, 0)
			score.UptimeScore = math.Max(score.UptimeScore-penalty, 0)
		}

		// 4. Ensure component scores don't exceed maximum
		score.BlockMiningScore = math.Min(score.BlockMiningScore, rs.config.MaxComponentScore)
		score.UptimeScore = math.Min(score.UptimeScore, rs.config.MaxComponentScore)
		score.ConsistencyScore = math.Min(score.ConsistencyScore, rs.config.MaxComponentScore)
	}
}

// performPartialReset performs a partial reset of validator scores
func (rs *ReputationSystem) performPartialReset(address common.Address) {
	if score, exists := rs.scores[address]; exists {
		// Reset 50% of accumulated scores to prevent infinite accumulation
		resetFactor := 0.5

		score.BlockMiningScore *= resetFactor
		score.UptimeScore *= resetFactor
		score.ConsistencyScore *= resetFactor

		// Update reset time
		score.LastReset = time.Now()

		log.Info("Partial score reset applied",
			"address", address.Hex(),
			"block_mining", score.BlockMiningScore,
			"uptime", score.UptimeScore,
			"consistency", score.ConsistencyScore)
	}
}

// calculateConsistencyScore calculates consistency score based on block mining intervals
func (rs *ReputationSystem) calculateConsistencyScore(address common.Address) {
	if score, exists := rs.scores[address]; exists {
		blockTimes := rs.blockTimes[address]
		if len(blockTimes) < 2 {
			score.ConsistencyScore = 0
			return
		}

		// Calculate average interval between blocks
		var totalInterval time.Duration
		for i := 1; i < len(blockTimes); i++ {
			totalInterval += blockTimes[i].Sub(blockTimes[i-1])
		}
		avgInterval := totalInterval / time.Duration(len(blockTimes)-1)

		// Calculate variance
		var variance float64
		for i := 1; i < len(blockTimes); i++ {
			interval := blockTimes[i].Sub(blockTimes[i-1])
			diff := float64(interval - avgInterval)
			variance += diff * diff
		}
		variance /= float64(len(blockTimes) - 1)

		// Consistency score is inverse of variance (lower variance = higher consistency)
		// Avoid division by zero and ensure valid calculation
		if avgInterval > 0 {
			consistencyScore := rs.config.ConsistencyReward / (1.0 + math.Sqrt(variance)/float64(avgInterval))
			// Ensure score is valid (not NaN or infinite)
			if math.IsNaN(consistencyScore) || math.IsInf(consistencyScore, 0) {
				consistencyScore = 0
			}
			score.ConsistencyScore = consistencyScore
		} else {
			score.ConsistencyScore = 0
		}
	}
}

// UpdateReputation updates reputation scores for all validators
func (rs *ReputationSystem) UpdateReputation() {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	now := time.Now()

	// Apply decay factor and fairness mechanisms
	for address, score := range rs.scores {
		if now.Sub(score.LastUpdate) > rs.config.UpdateInterval {
			// Apply stronger decay to prevent score accumulation
			score.BlockMiningScore *= rs.config.DecayFactor
			score.UptimeScore *= rs.config.DecayFactor
			score.ConsistencyScore *= rs.config.DecayFactor
			score.CurrentScore *= rs.config.DecayFactor
			score.LastUpdate = now

			// Recalculate total score with fairness mechanisms
			rs.calculateTotalScore(address)
		}
	}

	rs.lastUpdate = now
	rs.saveToDatabase()

	log.Info("Reputation scores updated", "validators", len(rs.scores))
}

// GetReputationScore returns the reputation score for a validator
func (rs *ReputationSystem) GetReputationScore(address common.Address) *ReputationScore {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	if score, exists := rs.scores[address]; exists {
		// Return a copy to avoid race conditions
		scoreCopy := *score
		return &scoreCopy
	}
	return nil
}

// GetAllValidators returns all validator addresses
func (rs *ReputationSystem) GetAllValidators() []common.Address {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	validators := make([]common.Address, 0, len(rs.scores))
	for address := range rs.scores {
		validators = append(validators, address)
	}
	return validators
}

// GetTopValidators returns validators sorted by reputation score
func (rs *ReputationSystem) GetTopValidators(limit int) []*ReputationScore {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	scores := make([]*ReputationScore, 0, len(rs.scores))
	for _, score := range rs.scores {
		if score.IsActive {
			scoreCopy := *score
			scores = append(scores, &scoreCopy)
		}
	}

	// Sort by current score (descending)
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].CurrentScore < scores[j].CurrentScore {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	if limit > 0 && limit < len(scores) {
		scores = scores[:limit]
	}

	return scores
}

// GetReputationStats returns statistics about the reputation system
func (rs *ReputationSystem) GetReputationStats() map[string]interface{} {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	stats := map[string]interface{}{
		"config": map[string]interface{}{
			"enable_reputation_system": rs.config.EnableReputationSystem,
			"initial_reputation":       rs.config.InitialReputation,
			"max_reputation":           rs.config.MaxReputation,
			"min_reputation":           rs.config.MinReputation,
			"evaluation_window":        rs.config.EvaluationWindow.String(),
			"update_interval":          rs.config.UpdateInterval.String(),
		},
		"validators": map[string]interface{}{
			"total":  len(rs.scores),
			"active": 0,
		},
		"reputation": map[string]interface{}{
			"average": 0.0,
			"highest": 0.0,
			"lowest":  0.0,
		},
		"events": map[string]interface{}{
			"total": len(rs.events),
		},
		"last_update": rs.lastUpdate,
	}

	// Calculate statistics
	var totalScore float64
	var activeCount int
	var highestScore, lowestScore float64

	first := true
	for _, score := range rs.scores {
		if score.IsActive {
			activeCount++
			totalScore += score.CurrentScore

			if first {
				highestScore = score.CurrentScore
				lowestScore = score.CurrentScore
				first = false
			} else {
				if score.CurrentScore > highestScore {
					highestScore = score.CurrentScore
				}
				if score.CurrentScore < lowestScore {
					lowestScore = score.CurrentScore
				}
			}
		}
	}

	stats["validators"].(map[string]interface{})["active"] = activeCount

	if activeCount > 0 {
		stats["reputation"].(map[string]interface{})["average"] = totalScore / float64(activeCount)
	}
	stats["reputation"].(map[string]interface{})["highest"] = highestScore
	stats["reputation"].(map[string]interface{})["lowest"] = lowestScore

	return stats
}

// GetReputationEvents returns recent reputation events
func (rs *ReputationSystem) GetReputationEvents(limit int) []ReputationEvent {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	events := make([]ReputationEvent, len(rs.events))
	copy(events, rs.events)

	// Sort by timestamp (newest first)
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	if limit > 0 && limit < len(events) {
		events = events[:limit]
	}

	return events
}

// saveToDatabase saves reputation data to database
func (rs *ReputationSystem) saveToDatabase() {
	if rs.db == nil {
		return
	}

	// Save scores
	scoresData, err := json.Marshal(rs.scores)
	if err == nil {
		rs.db.Put(rawdb.ReputationScoresKey, scoresData)
	}

	// Save events (keep only recent events)
	recentEvents := rs.events
	if len(recentEvents) > 1000 {
		recentEvents = recentEvents[len(recentEvents)-1000:]
	}
	eventsData, err := json.Marshal(recentEvents)
	if err == nil {
		rs.db.Put(rawdb.ReputationEventsKey, eventsData)
	}
}

// loadFromDatabase loads reputation data from database
func (rs *ReputationSystem) loadFromDatabase() {
	if rs.db == nil {
		return
	}

	// Load scores
	if scoresData, err := rs.db.Get(rawdb.ReputationScoresKey); err == nil {
		json.Unmarshal(scoresData, &rs.scores)
	}

	// Load events
	if eventsData, err := rs.db.Get(rawdb.ReputationEventsKey); err == nil {
		json.Unmarshal(eventsData, &rs.events)
	}

	log.Info("Reputation data loaded from database", "scores", len(rs.scores), "events", len(rs.events))
}

// ApplyDecay applies decay to a validator's reputation score
func (rs *ReputationSystem) ApplyDecay(address common.Address, decayFactor float64) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	if score, exists := rs.scores[address]; exists {
		oldScore := score.CurrentScore
		
		// Apply decay to all component scores
		score.BlockMiningScore *= (1.0 - decayFactor)
		score.UptimeScore *= (1.0 - decayFactor)
		score.ConsistencyScore *= (1.0 - decayFactor)
		
		// Recalculate total score
		rs.calculateTotalScore(address)
		
		// Record decay event
		event := ReputationEvent{
			Address:     address,
			EventType:   "decay",
			ScoreChange: score.CurrentScore - oldScore,
			BlockNumber: 0, // Decay is not block-specific
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Reputation decay applied: factor %.4f", decayFactor),
		}
		rs.events = append(rs.events, event)
		
		// Keep only recent events (last 1000)
		if len(rs.events) > 1000 {
			rs.events = rs.events[len(rs.events)-1000:]
		}
		
		score.LastUpdate = time.Now()
		rs.saveToDatabase()
	}
}

