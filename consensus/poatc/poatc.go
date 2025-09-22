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

// Package poatc implements the proof-of-authority consensus engine with AI tracing.
package poatc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	lru "github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

const (
	checkpointInterval = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	wiggleTime = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

// POATC proof-of-authority protocol constants.
var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes

	extraVanity = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	diffInTurn = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 20 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errWrongDifficulty is returned if the difficulty of a block doesn't match the
	// turn of the signer.
	errWrongDifficulty = errors.New("wrong difficulty")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned = errors.New("recently signed")
)

// SignerFn hashes and signs the data to be signed by a backing account.
type SignerFn func(signer accounts.Account, mimeType string, message []byte) ([]byte, error)

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, sigcache *sigLRU) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address, nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	sigcache.Add(hash, signer)
	return signer, nil
}

// POATC (Proof of Authority with AI Tracing) is the proof-of-authority consensus engine
// enhanced with AI-powered tracing, reputation system, and dynamic mechanisms.
type POATC struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.Cache[common.Hash, *Snapshot] // Snapshots for recent block to speed up reorgs
	signatures *sigLRU                            // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	signer common.Address // Ethereum address of the signing key
	signFn SignerFn       // Signer function to authorize hashes with
	lock   sync.RWMutex   // Protects the signer and proposals fields

	// Anomaly detection
	anomalyDetector *AnomalyDetector // Anomaly detection system

	// Whitelist/Blacklist management
	whitelistBlacklistManager *WhitelistBlacklistManager // Whitelist/blacklist management system

	// Validator selection management
	validatorSelectionManager *ValidatorSelectionManager // Validator selection system for 2-tier selection

	// Reputation system
	reputationSystem *ReputationSystem // On-chain reputation scoring system

	// Tracing system
	tracingSystem *TracingSystem // Tracing system with Merkle Tree support

	// Time dynamic system
	timeDynamicManager *TimeDynamicManager // Time dynamic mechanisms

	// The fields below are for testing only
	fakeDiff bool // Skip difficulty verifications
}

// New creates a POATC proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.CliqueConfig, db ethdb.Database) *POATC {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents := lru.NewCache[common.Hash, *Snapshot](inmemorySnapshots)
	signatures := lru.NewCache[common.Hash, common.Address](inmemorySignatures)

	return &POATC{
		config:     &conf,
		db:         db,
		recents:    recents,
		signatures: signatures,
		proposals:  make(map[common.Address]bool),
		// anomalyDetector will be initialized when signers are available
		// timeDynamicManager will be initialized when needed
	}
}

// initializeAnomalyDetector initializes the anomaly detector with current signers
func (c *POATC) initializeAnomalyDetector(signers []common.Address) {
	if c.anomalyDetector == nil {
		c.anomalyDetector = NewAnomalyDetector(DefaultAnomalyDetectionConfig(), signers)
		log.Info("Anomaly detector initialized", "signers", len(signers))
	}
}

// initializeWhitelistBlacklistManager initializes the whitelist/blacklist manager
func (c *POATC) initializeWhitelistBlacklistManager() {
	if c.whitelistBlacklistManager == nil {
		c.whitelistBlacklistManager = NewWhitelistBlacklistManager(DefaultWhitelistBlacklistConfig())
		log.Info("Whitelist/Blacklist manager initialized")
	}
}

// initializeValidatorSelectionManager initializes the validator selection manager
func (c *POATC) initializeValidatorSelectionManager(signers []common.Address) {
	if c.validatorSelectionManager == nil {
		c.validatorSelectionManager = NewValidatorSelectionManager(DefaultValidatorSelectionConfig())

		// Add all signers to the validator selection manager
		for _, signer := range signers {
			// Initialize with default stake and reputation
			defaultStake := big.NewInt(1000000) // 1M wei default stake
			defaultReputation := 1.0            // Default reputation
			c.validatorSelectionManager.AddValidator(signer, defaultStake, defaultReputation)
		}

		log.Info("Validator selection manager initialized", "signers", len(signers))
	}
}

// initializeReputationSystem initializes the reputation system
func (c *POATC) initializeReputationSystem(signers []common.Address) {
	if c.reputationSystem == nil {
		c.reputationSystem = NewReputationSystem(DefaultReputationConfig(), c.db)

		// Add all signers to the reputation system
		for _, signer := range signers {
			c.reputationSystem.AddValidator(signer)
		}

		log.Info("Reputation system initialized", "signers", len(signers))
	}
}

// initializeTracingSystem initializes the tracing system
func (c *POATC) initializeTracingSystem() {
	if c.tracingSystem == nil {
		c.tracingSystem = NewTracingSystem(DefaultTracingConfig())
		log.Info("Tracing system initialized with Merkle Tree support")
	}
}

// initializeTimeDynamicManager initializes the time dynamic manager
func (c *POATC) initializeTimeDynamicManager() {
	if c.timeDynamicManager == nil {
		c.timeDynamicManager = NewTimeDynamicManager(DefaultTimeDynamicConfig())

		// Set integration components
		c.timeDynamicManager.SetIntegrationComponents(
			c.validatorSelectionManager,
			c.reputationSystem,
			c.tracingSystem,
		)

		log.Info("Time dynamic manager initialized",
			"dynamic_block_time", c.timeDynamicManager.config.EnableDynamicBlockTime,
			"dynamic_validator_selection", c.timeDynamicManager.config.EnableDynamicValidatorSelection,
			"dynamic_reputation_decay", c.timeDynamicManager.config.EnableDynamicReputationDecay)
	}
}

// manageWhitelistBlacklistByReputation automatically manages whitelist/blacklist based on reputation
func (c *POATC) manageWhitelistBlacklistByReputation(signer common.Address, blockNumber uint64) {
	if c.reputationSystem == nil || c.whitelistBlacklistManager == nil {
		return
	}

	score := c.reputationSystem.GetReputationScore(signer)
	if score == nil {
		return
	}

	config := c.reputationSystem.config

	// Check if reputation is too low (should be blacklisted)
	if score.CurrentScore < config.LowReputationThreshold {
		if !c.whitelistBlacklistManager.IsBlacklisted(signer) {
			// Auto-blacklist validator with low reputation
			expiresAt := time.Now().Add(24 * time.Hour) // Auto-remove after 24 hours
			err := c.whitelistBlacklistManager.AddToBlacklist(
				signer,
				common.Address{}, // System address
				fmt.Sprintf("Auto-blacklisted due to low reputation: %.2f", score.CurrentScore),
				&expiresAt,
			)
			if err == nil {
				log.Warn("Validator auto-blacklisted due to low reputation",
					"address", signer.Hex(),
					"reputation", score.CurrentScore,
					"threshold", config.LowReputationThreshold)
			}
		}
	}

	// Check if reputation is high enough (should be whitelisted)
	if score.CurrentScore >= config.HighReputationThreshold {
		if !c.whitelistBlacklistManager.IsWhitelisted(signer) {
			// Auto-whitelist validator with high reputation
			err := c.whitelistBlacklistManager.AddToWhitelist(
				signer,
				common.Address{}, // System address
				fmt.Sprintf("Auto-whitelisted due to high reputation: %.2f", score.CurrentScore),
				nil, // No expiration
			)
			if err == nil {
				log.Info("Validator auto-whitelisted due to high reputation",
					"address", signer.Hex(),
					"reputation", score.CurrentScore,
					"threshold", config.HighReputationThreshold)
			}
		}
	}

	// Check if blacklisted validator has improved reputation
	if c.whitelistBlacklistManager.IsBlacklisted(signer) && score.CurrentScore >= config.HighReputationThreshold {
		// Remove from blacklist if reputation has improved
		err := c.whitelistBlacklistManager.RemoveFromBlacklist(signer)
		if err == nil {
			log.Info("Validator removed from blacklist due to improved reputation",
				"address", signer.Hex(),
				"reputation", score.CurrentScore)
		}
	}
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *POATC) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *POATC) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *POATC) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *POATC) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}
	// Checkpoint blocks need to enforce zero beneficiary
	checkpoint := (number % c.config.Epoch) == 0
	if checkpoint && header.Coinbase != (common.Address{}) {
		return errInvalidCheckpointBeneficiary
	}
	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}
	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}
	if checkpoint && signersBytes%common.AddressLength != 0 {
		return errInvalidCheckpointSigners
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return errInvalidDifficulty
		}
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}
	if chain.Config().IsShanghai(header.Number, header.Time) {
		return errors.New("poatc does not support shanghai fork")
	}
	// Verify the non-existence of withdrawalsHash.
	if header.WithdrawalsHash != nil {
		return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
	}
	if chain.Config().IsCancun(header.Number, header.Time) {
		return errors.New("poatc does not support cancun fork")
	}
	// Verify the non-existence of cancun-specific header fields
	switch {
	case header.ExcessBlobGas != nil:
		return fmt.Errorf("invalid excessBlobGas: have %d, expected nil", header.ExcessBlobGas)
	case header.BlobGasUsed != nil:
		return fmt.Errorf("invalid blobGasUsed: have %d, expected nil", header.BlobGasUsed)
	case header.ParentBeaconRoot != nil:
		return fmt.Errorf("invalid parentBeaconRoot, have %#x, expected nil", header.ParentBeaconRoot)
	}
	// All basic checks passed, verify cascading fields
	return c.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *POATC) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := eip1559.VerifyEIP1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}
	// If the block is a checkpoint block, verify the signer list
	if number%c.config.Epoch == 0 {
		signers := make([]byte, len(snap.Signers)*common.AddressLength)
		for i, signer := range snap.signers() {
			copy(signers[i*common.AddressLength:], signer[:])
		}
		extraSuffix := len(header.Extra) - extraSeal
		if !bytes.Equal(header.Extra[extraVanity:extraSuffix], signers) {
			return errMismatchingCheckpointSigners
		}
	}
	// All basic checks passed, verify the seal and return
	return c.verifySeal(snap, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *POATC) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%checkpointInterval == 0 {
			if s, err := loadSnapshot(c.config, c.signatures, c.db, hash); err == nil {
				log.Trace("Loaded voting snapshot from disk", "number", number, "hash", hash)
				snap = s
				break
			}
		}
		// If we're at the genesis, snapshot the initial state. Alternatively if we're
		// at a checkpoint block without a parent (light client CHT), or we have piled
		// up more headers than allowed to be reorged (chain reinit from a freezer),
		// consider the checkpoint trusted and snapshot it.
		if number == 0 || (number%c.config.Epoch == 0 && (len(headers) > params.FullImmutabilityThreshold || chain.GetHeaderByNumber(number-1) == nil)) {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()

				signers := make([]common.Address, (len(checkpoint.Extra)-extraVanity-extraSeal)/common.AddressLength)
				for i := 0; i < len(signers); i++ {
					copy(signers[i][:], checkpoint.Extra[extraVanity+i*common.AddressLength:])
				}
				snap = newSnapshot(c.config, c.signatures, number, hash, signers)
				if err := snap.store(c.db); err != nil {
					return nil, err
				}
				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}
	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = snap.store(c.db); err != nil {
			return nil, err
		}
		log.Trace("Stored voting snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}

	// Initialize validator selection manager if not already done
	if c.validatorSelectionManager == nil {
		signers := snap.signers()
		c.initializeValidatorSelectionManager(signers)
	}

	// Initialize reputation system if not already done
	if c.reputationSystem == nil {
		signers := snap.signers()
		c.initializeReputationSystem(signers)
	}

	// Initialize tracing system if not already done
	if c.tracingSystem == nil {
		c.initializeTracingSystem()
	}

	// Initialize time dynamic manager if not already done
	if c.timeDynamicManager == nil {
		c.initializeTimeDynamicManager()
	}

	// Set validator selection manager for this snapshot
	if c.validatorSelectionManager != nil {
		snap.setValidatorSelectionManager(c.validatorSelectionManager)
		// Set tracing system for validator selection manager
		if c.tracingSystem != nil {
			c.validatorSelectionManager.SetTracingSystem(c.tracingSystem)
		}
	}

	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *POATC) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *POATC) verifySeal(snap *Snapshot, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// Resolve the authorization key and check against signers
	signer, err := ecrecover(header, c.signatures)
	if err != nil {
		return err
	}

	// Initialize tracing system if not already done
	if c.tracingSystem == nil {
		c.initializeTracingSystem()
	}

	// Set current round for tracing
	if c.tracingSystem != nil {
		c.tracingSystem.SetCurrentRound(number)
	}

	if _, ok := snap.Signers[signer]; !ok {
		// Trace unauthorized signer attempt
		if c.tracingSystem != nil {
			c.tracingSystem.Trace(TraceEventBlockValidation, TraceLevelBasic, number, signer,
				"Unauthorized signer attempt", map[string]interface{}{
					"error":  "unauthorized_signer",
					"signer": signer.Hex(),
				})
		}
		return errUnauthorizedSigner
	}
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only fail if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
				// Trace recently signed error
				if c.tracingSystem != nil {
					c.tracingSystem.Trace(TraceEventBlockValidation, TraceLevelBasic, number, signer,
						"Recently signed error", map[string]interface{}{
							"error":     "recently_signed",
							"signer":    signer.Hex(),
							"last_seen": seen,
							"limit":     limit,
						})
				}
				return errRecentlySigned
			}
		}
	}
	// Ensure that the difficulty corresponds to the turn-ness of the signer
	if !c.fakeDiff {
		inturn := snap.inturn(header.Number.Uint64(), signer)
		if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
			return errWrongDifficulty
		}
		if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
			return errWrongDifficulty
		}
	}

	// Initialize whitelist/blacklist manager if not already done
	if c.whitelistBlacklistManager == nil {
		c.initializeWhitelistBlacklistManager()
	}

	// Validate signer against whitelist/blacklist
	if c.whitelistBlacklistManager != nil {
		valid, reason := c.whitelistBlacklistManager.ValidateSigner(signer)
		if !valid {
			// Trace whitelist/blacklist validation failure
			if c.tracingSystem != nil {
				c.tracingSystem.Trace(TraceEventWhitelistBlacklist, TraceLevelBasic, number, signer,
					"Whitelist/Blacklist validation failed", map[string]interface{}{
						"error":  "whitelist_blacklist_validation_failed",
						"signer": signer.Hex(),
						"reason": reason,
					})
			}
			log.Error("Whitelist/Blacklist validation failed", "signer", signer.Hex(), "reason", reason)
			return fmt.Errorf("whitelist/blacklist validation failed: %s", reason)
		}
	}

	// Initialize anomaly detector if not already done
	if c.anomalyDetector == nil {
		signers := make([]common.Address, 0, len(snap.Signers))
		for signer := range snap.Signers {
			signers = append(signers, signer)
		}
		c.initializeAnomalyDetector(signers)
	}

	// Add block to anomaly detector and check for anomalies
	if c.anomalyDetector != nil {
		c.anomalyDetector.AddBlock(header, signer)
		anomalies := c.anomalyDetector.DetectAnomalies()
		c.anomalyDetector.LogAnomalies(anomalies)

		// Trace anomaly detection results
		if c.tracingSystem != nil && len(anomalies) > 0 {
			for _, anomaly := range anomalies {
				// Convert AnomalyType to string
				anomalyTypeStr := ""
				switch anomaly.Type {
				case AnomalyRapidSigning:
					anomalyTypeStr = "RapidSigning"
				case AnomalySuspiciousPattern:
					anomalyTypeStr = "SuspiciousPattern"
				case AnomalyHighFrequency:
					anomalyTypeStr = "HighFrequency"
				case AnomalyMissingSigner:
					anomalyTypeStr = "MissingSigner"
				case AnomalyTimestampDrift:
					anomalyTypeStr = "TimestampDrift"
				default:
					anomalyTypeStr = "Unknown"
				}

				c.tracingSystem.TraceAnomalyDetection(
					anomalyTypeStr,
					anomaly.Signer,
					number,
					"medium", // severity
					map[string]interface{}{
						"anomaly_type": anomalyTypeStr,
						"message":      anomaly.Message,
						"timestamp":    anomaly.Timestamp,
						"severity":     anomaly.Severity,
					},
				)
			}
		}

		// Record violations in reputation system
		if c.reputationSystem != nil {
			for _, anomaly := range anomalies {
				if anomaly.Type == AnomalyRapidSigning || anomaly.Type == AnomalySuspiciousPattern ||
					anomaly.Type == AnomalyTimestampDrift || anomaly.Type == AnomalyMissingSigner {
					// Convert AnomalyType to string for RecordViolation
					violationType := ""
					switch anomaly.Type {
					case AnomalyRapidSigning:
						violationType = "RapidSigning"
					case AnomalySuspiciousPattern:
						violationType = "SuspiciousPattern"
					case AnomalyTimestampDrift:
						violationType = "TimestampDrift"
					case AnomalyMissingSigner:
						violationType = "MissingSigner"
					}
					c.reputationSystem.RecordViolation(signer, number, violationType, anomaly.Message)
				}
			}
		}
	}

	// Record block mining in reputation system
	if c.reputationSystem != nil {
		c.reputationSystem.RecordBlockMining(signer, number)

		// Update validator selection manager with new reputation
		if c.validatorSelectionManager != nil {
			if score := c.reputationSystem.GetReputationScore(signer); score != nil {
				c.validatorSelectionManager.UpdateValidatorReputation(signer, score.CurrentScore)
			}
		}

		// Auto-manage whitelist/blacklist based on reputation
		if c.whitelistBlacklistManager != nil {
			c.manageWhitelistBlacklistByReputation(signer, number)
		}

		// Trace reputation update
		if c.tracingSystem != nil {
			if score := c.reputationSystem.GetReputationScore(signer); score != nil {
				c.tracingSystem.TraceReputation(
					"block_mined",
					signer,
					number,
					score.PreviousScore,
					score.CurrentScore,
					map[string]interface{}{
						"block_mining_score": score.BlockMiningScore,
						"uptime_score":       score.UptimeScore,
						"consistency_score":  score.ConsistencyScore,
						"penalty_score":      score.PenaltyScore,
						"total_blocks_mined": score.TotalBlocksMined,
					},
				)
			}
		}
	}

	// Handle time dynamic mechanisms
	if c.timeDynamicManager != nil {
		// Check and trigger dynamic validator selection
		if c.timeDynamicManager.ShouldUpdateValidatorSelection() {
			err := c.timeDynamicManager.UpdateValidatorSelection(number, header.Hash())
			if err != nil {
				log.Error("Failed to update validator selection", "error", err)
			}
		}

		// Check and apply dynamic reputation decay
		if c.timeDynamicManager.ShouldApplyReputationDecay() {
			err := c.timeDynamicManager.ApplyReputationDecay()
			if err != nil {
				log.Error("Failed to apply reputation decay", "error", err)
			}
		}
	}

	// Trace successful block validation
	if c.tracingSystem != nil {
		c.tracingSystem.Trace(TraceEventBlockValidation, TraceLevelBasic, number, signer,
			"Block validation successful", map[string]interface{}{
				"signer":     signer.Hex(),
				"difficulty": header.Difficulty.String(),
				"timestamp":  header.Time,
			})
	}

	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *POATC) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// If the block isn't a checkpoint, cast a random vote (good enough for now)
	header.Coinbase = common.Address{}
	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()
	// Assemble the voting snapshot to check which votes make sense
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	c.lock.RLock()
	if number%c.config.Epoch != 0 {
		// Gather all the proposals that make sense voting on
		addresses := make([]common.Address, 0, len(c.proposals))
		for address, authorize := range c.proposals {
			if snap.validVote(address, authorize) {
				addresses = append(addresses, address)
			}
		}
		// If there's pending proposals, cast a vote on them
		if len(addresses) > 0 {
			header.Coinbase = addresses[rand.Intn(len(addresses))]
			if c.proposals[header.Coinbase] {
				copy(header.Nonce[:], nonceAuthVote)
			} else {
				copy(header.Nonce[:], nonceDropVote)
			}
		}
	}

	// Copy signer protected by mutex to avoid race condition
	signer := c.signer
	c.lock.RUnlock()

	// Set the correct difficulty
	header.Difficulty = calcDifficulty(snap, signer)

	// Ensure the extra data has all its components
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]

	if number%c.config.Epoch == 0 {
		for _, signer := range snap.signers() {
			header.Extra = append(header.Extra, signer[:]...)
		}
	}
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)

	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}

	// Ensure the timestamp has the correct delay
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Time = parent.Time + c.config.Period
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}
	return nil
}

// Finalize implements consensus.Engine. There is no post-transaction
// consensus rules in poatc, do nothing here.
func (c *POATC) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, withdrawals []*types.Withdrawal) {
	// No block rewards in PoA, so the state remains as is
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (c *POATC) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if len(withdrawals) > 0 {
		return nil, errors.New("clique does not support withdrawals")
	}
	// Finalize block
	c.Finalize(chain, header, state, txs, uncles, nil)

	// Assign the final state root to header.
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Assemble and return the final block for sealing.
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *POATC) Authorize(signer common.Address, signFn SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.signer = signer
	c.signFn = signFn
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *POATC) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := block.Header()

	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// For 0-period chains, refuse to seal empty blocks (no reward but would spin sealing)
	if c.config.Period == 0 && len(block.Transactions()) == 0 {
		return errors.New("sealing paused while waiting for transactions")
	}
	// Don't hold the signer fields for the entire sealing procedure
	c.lock.RLock()
	signer, signFn := c.signer, c.signFn
	c.lock.RUnlock()

	// Bail out if we're unauthorized to sign a block
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	if _, authorized := snap.Signers[signer]; !authorized {
		return errUnauthorizedSigner
	}
	// If we're amongst the recent signers, wait for the next block
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); number < limit || seen > number-limit {
				return errors.New("signed recently, must wait for others")
			}
		}
	}
	// Update transaction count for dynamic block time
	if c.timeDynamicManager != nil {
		txCount := len(block.Transactions())
		c.timeDynamicManager.UpdateTransactionCount(txCount)
	}

	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(int64(header.Time), 0).Sub(time.Now()) // nolint: gosimple

	// Apply dynamic block time adjustment
	if c.timeDynamicManager != nil {
		dynamicBlockTime := c.timeDynamicManager.GetCurrentBlockTime()
		baseBlockTime := time.Duration(c.config.Period) * time.Second

		// Adjust delay based on dynamic block time
		if dynamicBlockTime != baseBlockTime {
			timeRatio := float64(dynamicBlockTime) / float64(baseBlockTime)
			adjustedDelay := time.Duration(float64(delay) * timeRatio)

			// Ensure minimum delay to prevent consecutive blocks
			minDelay := 2 * time.Second // Minimum 2 seconds between blocks
			if adjustedDelay < minDelay {
				adjustedDelay = minDelay
			}

			// Ensure delay is not too large
			maxDelay := dynamicBlockTime + 5*time.Second // Max delay is dynamic block time + 5s buffer
			if adjustedDelay > maxDelay {
				adjustedDelay = maxDelay
			}

			delay = adjustedDelay

			log.Debug("Dynamic block time applied",
				"base_block_time", baseBlockTime,
				"dynamic_block_time", dynamicBlockTime,
				"time_ratio", timeRatio,
				"original_delay", time.Unix(int64(header.Time), 0).Sub(time.Now()),
				"adjusted_delay", delay,
				"tx_count", len(block.Transactions()))
		}
	}

	// Ensure absolute minimum delay to prevent consecutive blocks
	absoluteMinDelay := 1 * time.Second
	if delay < absoluteMinDelay {
		delay = absoluteMinDelay
		log.Debug("Applied absolute minimum delay", "delay", delay)
	}

	if header.Difficulty.Cmp(diffNoTurn) == 0 {
		// It's not our turn explicitly to sign, delay it a bit
		wiggle := time.Duration(len(snap.Signers)/2+1) * wiggleTime
		delay += time.Duration(rand.Int63n(int64(wiggle)))

		log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
	}
	// Sign all the things!
	sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeClique, CliqueRLP(header))
	if err != nil {
		return err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	// Wait until sealing is terminated or delay timeout.
	log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))
	go func() {
		select {
		case <-stop:
			return
		case <-time.After(delay):
		}

		select {
		case results <- block.WithSeal(header):
		default:
			log.Warn("Sealing result is not read by miner", "sealhash", SealHash(header))
		}
	}()

	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have:
// * DIFF_NOTURN(2) if BLOCK_NUMBER % SIGNER_COUNT != SIGNER_INDEX
// * DIFF_INTURN(1) if BLOCK_NUMBER % SIGNER_COUNT == SIGNER_INDEX
func (c *POATC) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	snap, err := c.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
	if err != nil {
		return nil
	}
	c.lock.RLock()
	signer := c.signer
	c.lock.RUnlock()
	return calcDifficulty(snap, signer)
}

func calcDifficulty(snap *Snapshot, signer common.Address) *big.Int {
	if snap.inturn(snap.Number+1, signer) {
		return new(big.Int).Set(diffInTurn)
	}
	return new(big.Int).Set(diffNoTurn)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *POATC) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *POATC) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *POATC) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	// Expose the API under both historical "clique" and new "poatc" namespaces
	// to preserve backward compatibility while presenting the new branding
	// (POATC: Proof-of-Authority with AI Tracing).
	return []rpc.API{
		{
			Namespace: "clique",
			Service:   &API{chain: chain, poatc: c},
		},
		{
			Namespace: "poatc",
			Service:   &API{chain: chain, poatc: c},
		},
	}
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

// CliqueRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func CliqueRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if header.WithdrawalsHash != nil {
		panic("unexpected withdrawal hash value in clique")
	}
	if header.ExcessBlobGas != nil {
		panic("unexpected excess blob gas value in clique")
	}
	if header.BlobGasUsed != nil {
		panic("unexpected blob gas used value in clique")
	}
	if header.ParentBeaconRoot != nil {
		panic("unexpected parent beacon root value in clique")
	}
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}
