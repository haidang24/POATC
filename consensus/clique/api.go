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

package clique

import (
	"encoding/json"
	"fmt"
	"math/big"

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
	chain  consensus.ChainHeaderReader
	clique *Clique
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
	return api.clique.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSnapshotAtHash retrieves the state snapshot at a given block.
func (api *API) GetSnapshotAtHash(hash common.Hash) (*Snapshot, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.clique.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
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
	snap, err := api.clique.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
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
	snap, err := api.clique.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.signers(), nil
}

// Proposals returns the current proposals the node tries to uphold and vote on.
func (api *API) Proposals() map[common.Address]bool {
	api.clique.lock.RLock()
	defer api.clique.lock.RUnlock()

	proposals := make(map[common.Address]bool)
	for address, auth := range api.clique.proposals {
		proposals[address] = auth
	}
	return proposals
}

// Propose injects a new authorization proposal that the signer will attempt to
// push through.
func (api *API) Propose(address common.Address, auth bool) {
	api.clique.lock.Lock()
	defer api.clique.lock.Unlock()

	api.clique.proposals[address] = auth
}

// Discard drops a currently running proposal, stopping the signer from casting
// further votes (either for or against).
func (api *API) Discard(address common.Address) {
	api.clique.lock.Lock()
	defer api.clique.lock.Unlock()

	delete(api.clique.proposals, address)
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
	snap, err := api.clique.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
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
		sealer, err := api.clique.Author(h)
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

// GetSigner returns the signer for a specific clique block.
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
		return api.clique.Author(header)
	}
	block := new(types.Block)
	if err := rlp.DecodeBytes(rlpOrBlockNr.RLP, block); err == nil {
		return api.clique.Author(block.Header())
	}
	header := new(types.Header)
	if err := rlp.DecodeBytes(rlpOrBlockNr.RLP, header); err != nil {
		return common.Address{}, err
	}
	return api.clique.Author(header)
}

// GetAnomalyStats returns statistics about detected anomalies
func (api *API) GetAnomalyStats() (map[string]interface{}, error) {
	if api.clique.anomalyDetector == nil {
		return map[string]interface{}{
			"error": "anomaly detector not initialized",
		}, nil
	}
	
	return api.clique.anomalyDetector.GetAnomalyStats(), nil
}

// DetectAnomalies manually triggers anomaly detection and returns results
func (api *API) DetectAnomalies() ([]AnomalyResult, error) {
	if api.clique.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}
	
	return api.clique.anomalyDetector.DetectAnomalies(), nil
}

// GetAnomalyConfig returns the current anomaly detection configuration
func (api *API) GetAnomalyConfig() (*AnomalyDetectionConfig, error) {
	if api.clique.anomalyDetector == nil {
		return nil, fmt.Errorf("anomaly detector not initialized")
	}
	
	return api.clique.anomalyDetector.config, nil
}

// Whitelist/Blacklist API endpoints

// GetWhitelistBlacklistStats returns statistics about whitelist and blacklist
func (api *API) GetWhitelistBlacklistStats() (map[string]interface{}, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return map[string]interface{}{
			"error": "whitelist/blacklist manager not initialized",
		}, nil
	}
	
	return api.clique.whitelistBlacklistManager.GetStats(), nil
}

// GetWhitelist returns the current whitelist
func (api *API) GetWhitelist() (map[string]WhitelistEntry, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return nil, fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	whitelist := api.clique.whitelistBlacklistManager.GetWhitelist()
	result := make(map[string]WhitelistEntry)
	for addr, entry := range whitelist {
		result[addr.Hex()] = entry
	}
	return result, nil
}

// GetBlacklist returns the current blacklist
func (api *API) GetBlacklist() (map[string]BlacklistEntry, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return nil, fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	blacklist := api.clique.whitelistBlacklistManager.GetBlacklist()
	result := make(map[string]BlacklistEntry)
	for addr, entry := range blacklist {
		result[addr.Hex()] = entry
	}
	return result, nil
}

// AddToWhitelist adds an address to the whitelist
func (api *API) AddToWhitelist(address common.Address, addedBy common.Address, reason string) error {
	if api.clique.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.AddToWhitelist(address, addedBy, reason, nil)
}

// RemoveFromWhitelist removes an address from the whitelist
func (api *API) RemoveFromWhitelist(address common.Address) error {
	if api.clique.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.RemoveFromWhitelist(address)
}

// AddToBlacklist adds an address to the blacklist
func (api *API) AddToBlacklist(address common.Address, addedBy common.Address, reason string) error {
	if api.clique.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.AddToBlacklist(address, addedBy, reason, nil)
}

// RemoveFromBlacklist removes an address from the blacklist
func (api *API) RemoveFromBlacklist(address common.Address) error {
	if api.clique.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.RemoveFromBlacklist(address)
}

// IsWhitelisted checks if an address is in the whitelist
func (api *API) IsWhitelisted(address common.Address) (bool, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return false, fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.IsWhitelisted(address), nil
}

// IsBlacklisted checks if an address is in the blacklist
func (api *API) IsBlacklisted(address common.Address) (bool, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return false, fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	return api.clique.whitelistBlacklistManager.IsBlacklisted(address), nil
}

// ValidateSigner validates if a signer is allowed to sign
func (api *API) ValidateSigner(address common.Address) (bool, string, error) {
	if api.clique.whitelistBlacklistManager == nil {
		return false, "", fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	valid, reason := api.clique.whitelistBlacklistManager.ValidateSigner(address)
	return valid, reason, nil
}

// CleanupExpiredEntries removes expired entries from whitelist and blacklist
func (api *API) CleanupExpiredEntries() error {
	if api.clique.whitelistBlacklistManager == nil {
		return fmt.Errorf("whitelist/blacklist manager not initialized")
	}
	
	api.clique.whitelistBlacklistManager.CleanupExpiredEntries()
	return nil
}

// ===== Validator Selection API =====

// GetValidatorSelectionStats returns statistics about validator selection
func (api *API) GetValidatorSelectionStats() (map[string]interface{}, error) {
	if api.clique.validatorSelectionManager == nil {
		return map[string]interface{}{
			"error": "validator selection manager not initialized",
		}, nil
	}
	
	return api.clique.validatorSelectionManager.GetStats(), nil
}

// GetSmallValidatorSet returns the current small validator set
func (api *API) GetSmallValidatorSet() ([]common.Address, error) {
	if api.clique.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}
	
	return api.clique.validatorSelectionManager.GetSmallValidatorSet(), nil
}

// GetValidatorInfo returns information about a specific validator
func (api *API) GetValidatorInfo(address common.Address) (*ValidatorInfo, error) {
	if api.clique.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}
	
	info := api.clique.validatorSelectionManager.GetValidatorInfo(address)
	if info == nil {
		return nil, fmt.Errorf("validator not found")
	}
	
	return info, nil
}

// AddValidator adds a new validator to the selection system
func (api *API) AddValidator(address common.Address, stake *big.Int, reputation float64) error {
	if api.clique.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}
	
	api.clique.validatorSelectionManager.AddValidator(address, stake, reputation)
	return nil
}

// UpdateValidatorStake updates a validator's stake
func (api *API) UpdateValidatorStake(address common.Address, stake *big.Int) error {
	if api.clique.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}
	
	api.clique.validatorSelectionManager.UpdateValidatorStake(address, stake)
	return nil
}

// UpdateValidatorReputation updates a validator's reputation
func (api *API) UpdateValidatorReputation(address common.Address, reputation float64) error {
	if api.clique.validatorSelectionManager == nil {
		return fmt.Errorf("validator selection manager not initialized")
	}
	
	api.clique.validatorSelectionManager.UpdateValidatorReputation(address, reputation)
	return nil
}

// GetSelectionHistory returns the validator selection history
func (api *API) GetSelectionHistory() ([]ValidatorSelectionRecord, error) {
	if api.clique.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}
	
	return api.clique.validatorSelectionManager.GetSelectionHistory(), nil
}

// ForceValidatorSelection forces a new validator selection
func (api *API) ForceValidatorSelection(blockNumber uint64, blockHash common.Hash) ([]common.Address, error) {
	if api.clique.validatorSelectionManager == nil {
		return nil, fmt.Errorf("validator selection manager not initialized")
	}
	
	return api.clique.validatorSelectionManager.SelectSmallValidatorSet(blockNumber, blockHash)
}
