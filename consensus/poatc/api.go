// Copyright 2017 The go-ethereum Authors
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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// API is a user facing RPC API to allow controlling the signer and voting
// mechanisms of the proof-of-authority scheme.
type API struct {
	chain consensus.ChainHeaderReader
	poatc *POATC
}

// GetSnapshot retrieves the state snapshot at a given block.
func (api *API) GetSnapshot(number *rpc.BlockNumber) (*Snapshot, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	// Ensure we have an actually valid block and return its snapshot
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.poatc.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSnapshotAtHash retrieves the state snapshot at a given block.
func (api *API) GetSnapshotAtHash(hash common.Hash) (*Snapshot, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.poatc.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSigners retrieves the list of authorized signers at the specified block.
func (api *API) GetSigners(number *rpc.BlockNumber) ([]common.Address, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	// Ensure we have an actually valid block and return the signers from its snapshot
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.poatc.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.signers(), nil
}

// GetSignersAtHash retrieves the list of authorized signers at the specified block.
func (api *API) GetSignersAtHash(hash common.Hash) ([]common.Address, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.poatc.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.signers(), nil
}

// Proposals returns the current proposals the node tries to uphold and vote on.
func (api *API) Proposals() map[common.Address]bool {
	api.poatc.lock.RLock()
	defer api.poatc.lock.RUnlock()

	proposals := make(map[common.Address]bool)
	for address, auth := range api.poatc.proposals {
		proposals[address] = auth
	}
	return proposals
}

// Propose injects a new authorization proposal that the signer will attempt to
// push through.
func (api *API) Propose(address common.Address, auth bool) {
	api.poatc.lock.Lock()
	defer api.poatc.lock.Unlock()

	api.poatc.proposals[address] = auth
}

// Discard drops a currently running proposal, stopping the signer from casting
// further votes (either for or against).
func (api *API) Discard(address common.Address) {
	api.poatc.lock.Lock()
	defer api.poatc.lock.Unlock()

	delete(api.poatc.proposals, address)
}

type status struct {
	InturnPercent float64                `json:"inturnPercent"`
	SigningStatus map[common.Address]int `json:"sealerActivity"`
	NumBlocks     uint64                 `json:"numBlocks"`
}

// Status returns the status of the last N blocks,
// - the number of active signers,
// - the number of signers,
// - the percentage of in-turn blocks
func (api *API) Status() (*status, error) {
	var (
		numBlocks = uint64(64)
		header    = api.chain.CurrentHeader()
		diff      = uint64(0)
		optimals  = 0
	)
	snap, err := api.poatc.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	var (
		signers = snap.signers()
		end     = header.Number.Uint64()
		start   = end - numBlocks
	)
	if numBlocks > end {
		start = 1
		numBlocks = end - start
	}
	signStatus := make(map[common.Address]int)
	for _, s := range signers {
		signStatus[s] = 0
	}
	for n := start; n < end; n++ {
		h := api.chain.GetHeaderByNumber(n)
		if h == nil {
			return nil, fmt.Errorf("missing block %d", n)
		}
		if h.Difficulty.Cmp(diffInTurn) == 0 {
			optimals++
		}
		diff += h.Difficulty.Uint64()
		sealer, err := api.poatc.Author(h)
		if err != nil {
			return nil, err
		}
		signStatus[sealer]++
	}
	return &status{
		InturnPercent: float64(100*optimals) / float64(numBlocks),
		SigningStatus: signStatus,
		NumBlocks:     numBlocks,
	}, nil
}

type blockNumberOrHashOrRLP struct {
	*rpc.BlockNumberOrHash
	RLP hexutil.Bytes `json:"rlp,omitempty"`
}

func (sb *blockNumberOrHashOrRLP) UnmarshalJSON(data []byte) error {
	bnOrHash := new(rpc.BlockNumberOrHash)
	// Try to unmarshal bNrOrHash
	if err := bnOrHash.UnmarshalJSON(data); err == nil {
		sb.BlockNumberOrHash = bnOrHash
		return nil
	}
	// Try to unmarshal RLP
	var input string
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	blob, err := hexutil.Decode(input)
	if err != nil {
		return err
	}
	sb.RLP = blob
	return nil
}

// GetSigner returns the signer for a specific poatc block.
// Can be called with a block number, a block hash or a rlp encoded blob.
// The RLP encoded blob can either be a block or a header.
func (api *API) GetSigner(rlpOrBlockNr *blockNumberOrHashOrRLP) (common.Address, error) {
	if len(rlpOrBlockNr.RLP) == 0 {
		blockNrOrHash := rlpOrBlockNr.BlockNumberOrHash
		var header *types.Header
		if blockNrOrHash == nil {
			header = api.chain.CurrentHeader()
		} else if hash, ok := blockNrOrHash.Hash(); ok {
			header = api.chain.GetHeaderByHash(hash)
		} else if number, ok := blockNrOrHash.Number(); ok {
			header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
		}
		if header == nil {
			return common.Address{}, fmt.Errorf("missing block %v", blockNrOrHash.String())
		}
		return api.poatc.Author(header)
	}
	block := new(types.Block)
	if err := rlp.DecodeBytes(rlpOrBlockNr.RLP, block); err == nil {
		return api.poatc.Author(block.Header())
	}
	header := new(types.Header)
	if err := rlp.DecodeBytes(rlpOrBlockNr.RLP, header); err != nil {
		return common.Address{}, err
	}
	return api.poatc.Author(header)
}

// GetAnomalyStats returns statistics about detected anomalies
func (api *API) GetAnomalyStats() (map[string]interface{}, error) {
	if api.poatc.anomalyDetector == nil {
		return map[string]interface{}{
			"error": "anomaly detector not initialized",
		}, nil
	}

	return api.poatc.anomalyDetector.GetAnomalyStats(), nil
}

// DetectAnomalies manually triggers anomaly detection and returns results
func (api *API) DetectAnomalies() ([]AnomalyResult, error) {
	if api.poatc.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}

	return api.poatc.anomalyDetector.DetectAnomalies(), nil
}

// GetAnomalyConfig returns the current anomaly detection configuration
func (api *API) GetAnomalyConfig() (*AnomalyDetectionConfig, error) {
	if api.poatc.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}

	return api.poatc.anomalyDetector.config, nil
}

// Whitelist/Blacklist API endpoints

// GetWhitelistBlacklistStats returns statistics about whitelist and blacklist
func (api *API) GetWhitelistBlacklistStats() (map[string]interface{}, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return map[string]interface{}{
			"error": "whitelist/blacklist manager not initialized",
		}, nil
	}

	return api.poatc.whitelistBlacklistManager.GetStats(), nil
}

// GetWhitelist returns the current whitelist
func (api *API) GetWhitelist() (map[string]WhitelistEntry, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return nil, fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	whitelist := api.poatc.whitelistBlacklistManager.GetWhitelist()
	result := make(map[string]WhitelistEntry)
	for addr, entry := range whitelist {
		result[addr.Hex()] = entry
	}
	return result, nil
}

// GetBlacklist returns the current blacklist
func (api *API) GetBlacklist() (map[string]BlacklistEntry, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return nil, fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	blacklist := api.poatc.whitelistBlacklistManager.GetBlacklist()
	result := make(map[string]BlacklistEntry)
	for addr, entry := range blacklist {
		result[addr.Hex()] = entry
	}
	return result, nil
}

// AddToWhitelist adds an address to the whitelist
func (api *API) AddToWhitelist(address common.Address, addedBy common.Address, reason string) error {
	if api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.AddToWhitelist(address, addedBy, reason, nil)
}

// RemoveFromWhitelist removes an address from the whitelist
func (api *API) RemoveFromWhitelist(address common.Address) error {
	if api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.RemoveFromWhitelist(address)
}

// AddToBlacklist adds an address to the blacklist
func (api *API) AddToBlacklist(address common.Address, addedBy common.Address, reason string) error {
	if api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.AddToBlacklist(address, addedBy, reason, nil)
}

// RemoveFromBlacklist removes an address from the blacklist
func (api *API) RemoveFromBlacklist(address common.Address) error {
	if api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.RemoveFromBlacklist(address)
}

// IsWhitelisted checks if an address is in the whitelist
func (api *API) IsWhitelisted(address common.Address) (bool, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return false, fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.IsWhitelisted(address), nil
}

// IsBlacklisted checks if an address is in the blacklist
func (api *API) IsBlacklisted(address common.Address) (bool, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return false, fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	return api.poatc.whitelistBlacklistManager.IsBlacklisted(address), nil
}

// ValidateSigner validates if a signer is allowed to sign
func (api *API) ValidateSigner(address common.Address) (bool, string, error) {
	if api.poatc.whitelistBlacklistManager == nil {
		return false, "", fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	valid, reason := api.poatc.whitelistBlacklistManager.ValidateSigner(address)
	return valid, reason, nil
}

// CleanupExpiredEntries removes expired entries from whitelist and blacklist
func (api *API) CleanupExpiredEntries() error {
	if api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}

	api.poatc.whitelistBlacklistManager.CleanupExpiredEntries()
	return nil
}

// ===== Validator Selection API =====

// GetValidatorSelectionStats returns statistics about validator selection
func (api *API) GetValidatorSelectionStats() (map[string]interface{}, error) {
	if api.poatc.validatorSelectionManager == nil {
		return map[string]interface{}{
			"error": "validator selection manager not initialized",
		}, nil
	}

	return api.poatc.validatorSelectionManager.GetStats(), nil
}

// GetSmallValidatorSet returns the current small validator set
func (api *API) GetSmallValidatorSet() ([]common.Address, error) {
	if api.poatc.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}

	return api.poatc.validatorSelectionManager.GetSmallValidatorSet(), nil
}

// GetValidatorInfo returns information about a specific validator
func (api *API) GetValidatorInfo(address common.Address) (*ValidatorInfo, error) {
	if api.poatc.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}

	info := api.poatc.validatorSelectionManager.GetValidatorInfo(address)
	if info == nil {
		return nil, fmt.Errorf("validator not found")
	}

	return info, nil
}

// AddValidator adds a new validator to the selection system
func (api *API) AddValidator(address common.Address, stake *big.Int, reputation float64) error {
	if api.poatc.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}

	api.poatc.validatorSelectionManager.AddValidator(address, stake, reputation)
	return nil
}

// UpdateValidatorStake updates a validator's stake
func (api *API) UpdateValidatorStake(address common.Address, stake *big.Int) error {
	if api.poatc.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}

	api.poatc.validatorSelectionManager.UpdateValidatorStake(address, stake)
	return nil
}

// UpdateValidatorReputation updates a validator's reputation
func (api *API) UpdateValidatorReputation(address common.Address, reputation float64) error {
	if api.poatc.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}

	api.poatc.validatorSelectionManager.UpdateValidatorReputation(address, reputation)
	return nil
}

// GetSelectionHistory returns the validator selection history
func (api *API) GetSelectionHistory() ([]ValidatorSelectionRecord, error) {
	if api.poatc.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}

	return api.poatc.validatorSelectionManager.GetSelectionHistory(), nil
}

// ForceValidatorSelection forces a new validator selection
func (api *API) ForceValidatorSelection(blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if api.poatc.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}

	return api.poatc.validatorSelectionManager.SelectSmallValidatorSet(blockNumber, blockHash)
}

// ===== Reputation System API =====

// GetReputationStats returns statistics about the reputation system
func (api *API) GetReputationStats() (map[string]interface{}, error) {
	if api.poatc.reputationSystem == nil {
		return map[string]interface{}{
			"error": "reputation system not initialized",
		}, nil
	}

	return api.poatc.reputationSystem.GetReputationStats(), nil
}

// GetReputationScore returns the reputation score for a specific validator
func (api *API) GetReputationScore(address common.Address) (*ReputationScore, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not initialized")
	}

	score := api.poatc.reputationSystem.GetReputationScore(address)
	if score == nil {
		return nil, fmt.Errorf("validator not found in reputation system")
	}

	return score, nil
}

// GetTopValidators returns validators sorted by reputation score
func (api *API) GetTopValidators(limit int) ([]*ReputationScore, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not initialized")
	}

	return api.poatc.reputationSystem.GetTopValidators(limit), nil
}

// GetReputationEvents returns recent reputation events
func (api *API) GetReputationEvents(limit int) ([]ReputationEvent, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not initialized")
	}

	return api.poatc.reputationSystem.GetReputationEvents(limit), nil
}

// RecordViolation records a violation by a validator
func (api *API) RecordViolation(address common.Address, blockNumber uint64, violationType string, description string) error {
	if api.poatc.reputationSystem == nil {
		return fmt.Errorf("reputation system not initialized")
	}

	api.poatc.reputationSystem.RecordViolation(address, blockNumber, violationType, description)
	return nil
}

// UpdateReputation manually updates reputation scores
func (api *API) UpdateReputation() error {
	if api.poatc.reputationSystem == nil {
		return fmt.Errorf("reputation system not initialized")
	}

	api.poatc.reputationSystem.UpdateReputation()
	return nil
}

// MarkValidatorOffline marks a validator as offline
func (api *API) MarkValidatorOffline(address common.Address) error {
	if api.poatc.reputationSystem == nil {
		return fmt.Errorf("reputation system not initialized")
	}

	api.poatc.reputationSystem.MarkValidatorOffline(address)
	return nil
}

// UpdateValidatorUptime updates the uptime for a validator
func (api *API) UpdateValidatorUptime(address common.Address) error {
	if api.poatc.reputationSystem == nil {
		return fmt.Errorf("reputation system not initialized")
	}

	api.poatc.reputationSystem.UpdateUptime(address)
	return nil
}

// ===== Integration Management API =====

// GetIntegrationStatus returns the status of all system integrations
func (api *API) GetIntegrationStatus() (map[string]interface{}, error) {
	status := map[string]interface{}{
		"reputation_system": map[string]interface{}{
			"initialized": api.poatc.reputationSystem != nil,
			"enabled":     api.poatc.reputationSystem != nil && api.poatc.reputationSystem.config.EnableReputationSystem,
		},
		"validator_selection": map[string]interface{}{
			"initialized": api.poatc.validatorSelectionManager != nil,
		},
		"anomaly_detection": map[string]interface{}{
			"initialized": api.poatc.anomalyDetector != nil,
		},
		"whitelist_blacklist": map[string]interface{}{
			"initialized": api.poatc.whitelistBlacklistManager != nil,
		},
		"integrations": map[string]interface{}{
			"reputation_to_validator_selection": api.poatc.reputationSystem != nil && api.poatc.validatorSelectionManager != nil,
			"anomaly_detection_to_reputation":   api.poatc.anomalyDetector != nil && api.poatc.reputationSystem != nil,
			"reputation_to_whitelist_blacklist": api.poatc.reputationSystem != nil && api.poatc.whitelistBlacklistManager != nil,
		},
	}

	return status, nil
}

// ForceReputationBasedWhitelistBlacklist forces whitelist/blacklist management based on current reputation scores
func (api *API) ForceReputationBasedWhitelistBlacklist() error {
	if api.poatc.reputationSystem == nil || api.poatc.whitelistBlacklistManager == nil {
		return fmt.Errorf("reputation system or whitelist/blacklist manager not initialized")
	}

	// Get all validators and their reputation scores
	topValidators := api.poatc.reputationSystem.GetTopValidators(0) // Get all validators

	config := api.poatc.reputationSystem.config
	processed := 0

	for _, validator := range topValidators {
		// Check if should be blacklisted
		if validator.CurrentScore < config.LowReputationThreshold {
			if !api.poatc.whitelistBlacklistManager.IsBlacklisted(validator.Address) {
				expiresAt := time.Now().Add(24 * time.Hour)
				err := api.poatc.whitelistBlacklistManager.AddToBlacklist(
					validator.Address,
					common.Address{}, // System address
					fmt.Sprintf("Force blacklisted due to low reputation: %.2f", validator.CurrentScore),
					&expiresAt,
				)
				if err == nil {
					processed++
				}
			}
		}

		// Check if should be whitelisted
		if validator.CurrentScore >= config.HighReputationThreshold {
			if !api.poatc.whitelistBlacklistManager.IsWhitelisted(validator.Address) {
				err := api.poatc.whitelistBlacklistManager.AddToWhitelist(
					validator.Address,
					common.Address{}, // System address
					fmt.Sprintf("Force whitelisted due to high reputation: %.2f", validator.CurrentScore),
					nil, // No expiration
				)
				if err == nil {
					processed++
				}
			}
		}
	}

	return nil
}

// GetReputationBasedRecommendations returns recommendations for whitelist/blacklist based on reputation
func (api *API) GetReputationBasedRecommendations() (map[string]interface{}, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not initialized")
	}

	topValidators := api.poatc.reputationSystem.GetTopValidators(0)
	config := api.poatc.reputationSystem.config

	recommendations := map[string]interface{}{
		"thresholds": map[string]interface{}{
			"high_reputation": config.HighReputationThreshold,
			"low_reputation":  config.LowReputationThreshold,
		},
		"recommendations": map[string]interface{}{
			"whitelist": []map[string]interface{}{},
			"blacklist": []map[string]interface{}{},
		},
	}

	whitelistRecs := []map[string]interface{}{}
	blacklistRecs := []map[string]interface{}{}

	for _, validator := range topValidators {
		if validator.CurrentScore >= config.HighReputationThreshold {
			whitelistRecs = append(whitelistRecs, map[string]interface{}{
				"address":    validator.Address.Hex(),
				"reputation": validator.CurrentScore,
				"reason":     fmt.Sprintf("High reputation score: %.2f", validator.CurrentScore),
			})
		} else if validator.CurrentScore < config.LowReputationThreshold {
			blacklistRecs = append(blacklistRecs, map[string]interface{}{
				"address":    validator.Address.Hex(),
				"reputation": validator.CurrentScore,
				"reason":     fmt.Sprintf("Low reputation score: %.2f", validator.CurrentScore),
			})
		}
	}

	recommendations["recommendations"].(map[string]interface{})["whitelist"] = whitelistRecs
	recommendations["recommendations"].(map[string]interface{})["blacklist"] = blacklistRecs

	return recommendations, nil
}

// ===== Fairness Management API =====

// GetFairnessStats returns statistics about fairness mechanisms
func (api *API) GetFairnessStats() (map[string]interface{}, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not enabled")
	}

	stats := make(map[string]interface{})
	config := api.poatc.reputationSystem.config

	// Configuration
	stats["max_component_score"] = config.MaxComponentScore
	stats["reset_interval_hours"] = config.ResetInterval.Hours()
	stats["new_validator_boost"] = config.NewValidatorBoost
	stats["veteran_penalty"] = config.VeteranPenalty
	stats["decay_factor"] = config.DecayFactor

	// Validator statistics
	allValidators := api.poatc.reputationSystem.GetAllValidators()
	newValidators := 0
	veteranValidators := 0
	now := time.Now()

	for _, validator := range allValidators {
		score := api.poatc.reputationSystem.GetReputationScore(validator)
		if score != nil {
			if score.IsNewValidator && now.Sub(score.JoinTime) < 24*time.Hour {
				newValidators++
			}
			if now.Sub(score.JoinTime) > 30*24*time.Hour {
				veteranValidators++
			}
		}
	}

	stats["total_validators"] = len(allValidators)
	stats["new_validators"] = newValidators
	stats["veteran_validators"] = veteranValidators

	return stats, nil
}

// ForcePartialReset forces a partial reset for a specific validator
func (api *API) ForcePartialReset(address common.Address) error {
	if api.poatc.reputationSystem == nil {
		return fmt.Errorf("reputation system not enabled")
	}

	// This would need to be implemented in the reputation system
	// For now, we'll just trigger a reputation update
	api.poatc.reputationSystem.UpdateReputation()

	return nil
}

// GetValidatorFairnessInfo returns fairness information for a specific validator
func (api *API) GetValidatorFairnessInfo(address common.Address) (map[string]interface{}, error) {
	if api.poatc.reputationSystem == nil {
		return nil, fmt.Errorf("reputation system not enabled")
	}

	score := api.poatc.reputationSystem.GetReputationScore(address)
	if score == nil {
		return nil, fmt.Errorf("validator not found")
	}

	info := make(map[string]interface{})
	now := time.Now()

	info["address"] = address.Hex()
	info["join_time"] = score.JoinTime
	info["last_reset"] = score.LastReset
	info["is_new_validator"] = score.IsNewValidator
	info["veteran_penalty"] = score.VeteranPenalty

	// Calculate time-based info
	info["days_since_join"] = now.Sub(score.JoinTime).Hours() / 24
	info["hours_since_reset"] = now.Sub(score.LastReset).Hours()

	// Component scores
	info["block_mining_score"] = score.BlockMiningScore
	info["uptime_score"] = score.UptimeScore
	info["consistency_score"] = score.ConsistencyScore
	info["penalty_score"] = score.PenaltyScore
	info["current_score"] = score.CurrentScore

	// Fairness status
	config := api.poatc.reputationSystem.config
	info["is_at_max_component"] = score.BlockMiningScore >= config.MaxComponentScore ||
		score.UptimeScore >= config.MaxComponentScore ||
		score.ConsistencyScore >= config.MaxComponentScore

	info["needs_reset"] = now.Sub(score.LastReset) >= config.ResetInterval
	info["is_veteran"] = now.Sub(score.JoinTime) > 30*24*time.Hour

	return info, nil
}

// ===== Tracing System API =====

// GetTracingStats returns statistics about the tracing system
func (api *API) GetTracingStats() (map[string]interface{}, error) {
	if api.poatc.tracingSystem == nil {
		return nil, fmt.Errorf("tracing system not enabled")
	}

	return api.poatc.tracingSystem.GetTraceStats(), nil
}

// GetTraceEvents returns trace events with optional filtering
func (api *API) GetTraceEvents(eventType string, level int, limit int) ([]TraceEvent, error) {
	if api.poatc.tracingSystem == nil {
		return nil, fmt.Errorf("tracing system not enabled")
	}

	var traceEventType TraceEventType
	if eventType != "" {
		traceEventType = TraceEventType(eventType)
	}

	var traceLevel TraceLevel
	if level >= 0 && level <= 3 {
		traceLevel = TraceLevel(level)
	} else {
		traceLevel = TraceLevelOff
	}

	return api.poatc.tracingSystem.GetTraceEvents(traceEventType, traceLevel, limit), nil
}

// GetMerkleRoot returns the current Merkle root of the tracing system
func (api *API) GetMerkleRoot() (string, error) {
	if api.poatc.tracingSystem == nil {
		return "", fmt.Errorf("tracing system not enabled")
	}

	return api.poatc.tracingSystem.GetMerkleRoot().Hex(), nil
}

// VerifyEventInMerkleTree verifies if an event is in the Merkle Tree
func (api *API) VerifyEventInMerkleTree(event TraceEvent) (bool, error) {
	if api.poatc.tracingSystem == nil {
		return false, fmt.Errorf("tracing system not enabled")
	}

	return api.poatc.tracingSystem.VerifyEventInMerkleTree(event), nil
}

// GetMerkleProof returns the Merkle proof for an event
func (api *API) GetMerkleProof(event TraceEvent) ([]string, bool, error) {
	if api.poatc.tracingSystem == nil {
		return nil, false, fmt.Errorf("tracing system not enabled")
	}

	proof, found := api.poatc.tracingSystem.GetMerkleProof(event)
	if !found {
		return nil, false, nil
	}

	// Convert proof to string array
	proofStrings := make([]string, len(proof))
	for i, hash := range proof {
		proofStrings[i] = hash.Hex()
	}

	return proofStrings, true, nil
}

// ExportTraceEvents exports all trace events and Merkle Tree to JSON
func (api *API) ExportTraceEvents() (string, error) {
	if api.poatc.tracingSystem == nil {
		return "", fmt.Errorf("tracing system not enabled")
	}

	data, err := api.poatc.tracingSystem.ExportTraceEvents()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ClearTraceEvents clears all trace events and Merkle Tree
func (api *API) ClearTraceEvents() error {
	if api.poatc.tracingSystem == nil {
		return fmt.Errorf("tracing system not enabled")
	}

	api.poatc.tracingSystem.ClearTraceEvents()
	return nil
}

// SetTraceLevel sets the trace level
func (api *API) SetTraceLevel(level int) error {
	if api.poatc.tracingSystem == nil {
		return fmt.Errorf("tracing system not enabled")
	}

	if level < 0 || level > 3 {
		return fmt.Errorf("invalid trace level: %d (must be 0-3)", level)
	}

	api.poatc.tracingSystem.SetTraceLevel(TraceLevel(level))
	return nil
}

// EnableTracing enables or disables tracing
func (api *API) EnableTracing(enable bool) error {
	if api.poatc.tracingSystem == nil {
		return fmt.Errorf("tracing system not enabled")
	}

	api.poatc.tracingSystem.EnableTracing(enable)
	return nil
}

// GetTraceMetrics returns current trace metrics
func (api *API) GetTraceMetrics() (map[string]interface{}, error) {
	if api.poatc.tracingSystem == nil {
		return nil, fmt.Errorf("tracing system not enabled")
	}

	return api.poatc.tracingSystem.GetTraceMetrics(), nil
}

// ===== Time Dynamic System API =====

// GetTimeDynamicStats returns statistics about time dynamic mechanisms
func (api *API) GetTimeDynamicStats() (map[string]interface{}, error) {
	if api.poatc.timeDynamicManager == nil {
		return nil, fmt.Errorf("time dynamic system not enabled")
	}
	return api.poatc.timeDynamicManager.GetTimeDynamicStats(), nil
}

// GetCurrentBlockTime returns the current dynamic block time
func (api *API) GetCurrentBlockTime() (float64, error) {
	if api.poatc.timeDynamicManager == nil {
		return 0, fmt.Errorf("time dynamic system not enabled")
	}
	return api.poatc.timeDynamicManager.GetCurrentBlockTime().Seconds(), nil
}

// UpdateTransactionCount manually updates transaction count for testing
func (api *API) UpdateTransactionCount(txCount int) error {
	if api.poatc.timeDynamicManager == nil {
		return fmt.Errorf("time dynamic system not enabled")
	}
	api.poatc.timeDynamicManager.UpdateTransactionCount(txCount)
	return nil
}

// TriggerValidatorSelection manually triggers validator selection update
func (api *API) TriggerValidatorSelection(blockNumber uint64, blockHash string) error {
	if api.poatc.timeDynamicManager == nil {
		return fmt.Errorf("time dynamic system not enabled")
	}

	hash := common.HexToHash(blockHash)
	return api.poatc.timeDynamicManager.UpdateValidatorSelection(blockNumber, hash)
}

// TriggerReputationDecay manually triggers reputation decay
func (api *API) TriggerReputationDecay() error {
	if api.poatc.timeDynamicManager == nil {
		return fmt.Errorf("time dynamic system not enabled")
	}
	return api.poatc.timeDynamicManager.ApplyReputationDecay()
}

// GetDecayHistory returns the recent decay history
func (api *API) GetDecayHistory(limit int) ([]DecayRecord, error) {
	if api.poatc.timeDynamicManager == nil {
		return nil, fmt.Errorf("time dynamic system not enabled")
	}
	return api.poatc.timeDynamicManager.GetDecayHistory(limit), nil
}

// UpdateTimeDynamicConfig updates the time dynamic configuration
func (api *API) UpdateTimeDynamicConfig(config *TimeDynamicConfig) error {
	if api.poatc.timeDynamicManager == nil {
		return fmt.Errorf("time dynamic system not enabled")
	}
	api.poatc.timeDynamicManager.UpdateConfig(config)
	return nil
}

// GetTimeDynamicConfig returns the current time dynamic configuration
func (api *API) GetTimeDynamicConfig() (*TimeDynamicConfig, error) {
	if api.poatc.timeDynamicManager == nil {
		return nil, fmt.Errorf("time dynamic system not enabled")
	}
	return api.poatc.timeDynamicManager.config, nil
}
