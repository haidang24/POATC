package poatc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// TraceLevel represents the level of tracing detail
type TraceLevel int

const (
	TraceLevelOff TraceLevel = iota
	TraceLevelBasic
	TraceLevelDetailed
	TraceLevelVerbose
)

// TraceEventType represents the type of trace event
type TraceEventType string

const (
	TraceEventRandomPOA          TraceEventType = "random_poa"
	TraceEventLeaderSelection    TraceEventType = "leader_selection"
	TraceEventBlockSigning       TraceEventType = "block_signing"
	TraceEventBlockValidation    TraceEventType = "block_validation"
	TraceEventTimeout            TraceEventType = "timeout"
	TraceEventAccusation         TraceEventType = "accusation"
	TraceEventAIGateEvaluation   TraceEventType = "ai_gate_evaluation"
	TraceEventReputation         TraceEventType = "reputation"
	TraceEventReputationUpdate   TraceEventType = "reputation_update"
	TraceEventAnomalyDetection   TraceEventType = "anomaly_detection"
	TraceEventWhitelistBlacklist TraceEventType = "whitelist_blacklist"
	TraceEventValidatorSelection TraceEventType = "validator_selection"
	TraceEventMerkleRoot         TraceEventType = "merkle_root"
	TraceEventTimeDynamic        TraceEventType = "time_dynamic"
)

// TraceEvent represents a single trace event with Merkle Tree support
type TraceEvent struct {
	ID          string                 `json:"id"`
	Type        TraceEventType         `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	BlockNumber uint64                 `json:"block_number"`
	Round       uint64                 `json:"round"`
	Address     common.Address         `json:"address"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Level       TraceLevel             `json:"level"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Hash        common.Hash            `json:"hash"`        // Hash of this event
	MerklePath  []common.Hash          `json:"merkle_path"` // Path to Merkle root
}

// MerkleNode represents a node in the Merkle Tree
type MerkleNode struct {
	Hash  common.Hash `json:"hash"`
	Left  *MerkleNode `json:"left,omitempty"`
	Right *MerkleNode `json:"right,omitempty"`
	Data  []byte      `json:"data,omitempty"`
}

// MerkleTree represents the Merkle Tree for trace events
type MerkleTree struct {
	Root   *MerkleNode `json:"root"`
	Leaves []common.Hash `json:"leaves"`
	Events []TraceEvent `json:"events"`
}

// TracingConfig contains configuration for the tracing system
type TracingConfig struct {
	EnableTracing     bool        `json:"enable_tracing"`
	TraceLevel        TraceLevel  `json:"trace_level"`
	MaxTraceEvents    int         `json:"max_trace_events"`
	TraceRetention    time.Duration `json:"trace_retention"`
	EnableMerkleTree  bool        `json:"enable_merkle_tree"`
	EnablePersistence bool        `json:"enable_persistence"`
	EnableMetrics     bool        `json:"enable_metrics"`
	MerkleRootInBlock bool        `json:"merkle_root_in_block"`
}

// DefaultTracingConfig returns a default configuration
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		EnableTracing:     true,
		TraceLevel:        TraceLevelDetailed,
		MaxTraceEvents:    10000,
		TraceRetention:    24 * time.Hour,
		EnableMerkleTree:  true,
		EnablePersistence: true,
		EnableMetrics:     true,
		MerkleRootInBlock: true,
	}
}

// TracingSystem manages all tracing operations with Merkle Tree support
type TracingSystem struct {
	config      *TracingConfig
	events      []TraceEvent
	merkleTree  *MerkleTree
	metrics     map[string]interface{}
	mutex       sync.RWMutex
	eventCount  int64
	startTime   time.Time
	currentRound uint64
}

// NewTracingSystem creates a new tracing system
func NewTracingSystem(config *TracingConfig) *TracingSystem {
	if config == nil {
		config = DefaultTracingConfig()
	}

	ts := &TracingSystem{
		config:  config,
		events:  make([]TraceEvent, 0, config.MaxTraceEvents),
		merkleTree: &MerkleTree{
			Leaves: make([]common.Hash, 0),
			Events: make([]TraceEvent, 0),
		},
		metrics: make(map[string]interface{}),
		startTime: time.Now(),
	}

	// Initialize metrics
	ts.initializeMetrics()

	return ts
}

// initializeMetrics initializes the metrics map
func (ts *TracingSystem) initializeMetrics() {
	ts.metrics = map[string]interface{}{
		"total_events":       0,
		"events_by_type":     make(map[string]int),
		"events_by_level":    make(map[string]int),
		"merkle_trees_built": 0,
		"merkle_roots_generated": 0,
		"average_duration":   0.0,
		"max_duration":       0.0,
		"min_duration":       0.0,
		"system_uptime":      time.Since(ts.startTime),
		"last_event_time":    time.Time{},
		"events_per_minute":  0.0,
		"current_round":      0,
	}
}

// Trace records a trace event and adds it to Merkle Tree
func (ts *TracingSystem) Trace(eventType TraceEventType, level TraceLevel, blockNumber uint64, address common.Address, message string, data map[string]interface{}) {
	if !ts.config.EnableTracing || level > ts.config.TraceLevel {
		return
	}

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	event := TraceEvent{
		ID:          fmt.Sprintf("%d-%d-%d", blockNumber, ts.currentRound, ts.eventCount),
		Type:        eventType,
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		Round:       ts.currentRound,
		Address:     address,
		Message:     message,
		Data:        data,
		Level:       level,
	}

	// Calculate hash for this event
	event.Hash = ts.calculateEventHash(event)

	// Add to events list
	ts.events = append(ts.events, event)
	ts.eventCount++

	// Add to Merkle Tree if enabled
	if ts.config.EnableMerkleTree {
		ts.addToMerkleTree(event)
	}

	// Maintain max events limit
	if len(ts.events) > ts.config.MaxTraceEvents {
		ts.events = ts.events[1:]
	}

	// Update metrics
	ts.updateMetrics(event)

	// Log to console if verbose
	if level == TraceLevelVerbose {
		log.Debug("Tracing Event", 
			"type", eventType,
			"block", blockNumber,
			"round", ts.currentRound,
			"address", address.Hex(),
			"message", message,
			"hash", event.Hash.Hex(),
			"data", data)
	}
}

// TraceWithDuration records a trace event with duration
func (ts *TracingSystem) TraceWithDuration(eventType TraceEventType, level TraceLevel, blockNumber uint64, address common.Address, message string, data map[string]interface{}, duration time.Duration) {
	if !ts.config.EnableTracing || level > ts.config.TraceLevel {
		return
	}

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	event := TraceEvent{
		ID:          fmt.Sprintf("%d-%d-%d", blockNumber, ts.currentRound, ts.eventCount),
		Type:        eventType,
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		Round:       ts.currentRound,
		Address:     address,
		Message:     message,
		Data:        data,
		Level:       level,
		Duration:    duration,
	}

	// Calculate hash for this event
	event.Hash = ts.calculateEventHash(event)

	// Add to events list
	ts.events = append(ts.events, event)
	ts.eventCount++

	// Add to Merkle Tree if enabled
	if ts.config.EnableMerkleTree {
		ts.addToMerkleTree(event)
	}

	// Maintain max events limit
	if len(ts.events) > ts.config.MaxTraceEvents {
		ts.events = ts.events[1:]
	}

	// Update metrics
	ts.updateMetrics(event)

	// Log to console if verbose
	if level == TraceLevelVerbose {
		log.Debug("Tracing Event", 
			"type", eventType,
			"block", blockNumber,
			"round", ts.currentRound,
			"address", address.Hex(),
			"message", message,
			"duration", duration,
			"hash", event.Hash.Hex(),
			"data", data)
	}
}

// calculateEventHash calculates SHA256 hash of an event
func (ts *TracingSystem) calculateEventHash(event TraceEvent) common.Hash {
	// Create a deterministic representation of the event
	eventData := map[string]interface{}{
		"id":           event.ID,
		"type":         event.Type,
		"timestamp":    event.Timestamp.UnixNano(),
		"block_number": event.BlockNumber,
		"round":        event.Round,
		"address":      event.Address.Hex(),
		"message":      event.Message,
		"data":         event.Data,
		"duration":     event.Duration.Nanoseconds(),
	}

	// Convert to JSON for consistent hashing
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		log.Error("Failed to marshal event for hashing", "error", err)
		return common.Hash{}
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(jsonData)
	return common.BytesToHash(hash[:])
}

// addToMerkleTree adds an event to the Merkle Tree
func (ts *TracingSystem) addToMerkleTree(event TraceEvent) {
	// Add event to Merkle Tree events
	ts.merkleTree.Events = append(ts.merkleTree.Events, event)
	ts.merkleTree.Leaves = append(ts.merkleTree.Leaves, event.Hash)

	// Rebuild Merkle Tree
	ts.rebuildMerkleTree()
}

// rebuildMerkleTree rebuilds the Merkle Tree from current events
func (ts *TracingSystem) rebuildMerkleTree() {
	if len(ts.merkleTree.Leaves) == 0 {
		ts.merkleTree.Root = nil
		return
	}

	// Sort leaves for consistent ordering
	leaves := make([]common.Hash, len(ts.merkleTree.Leaves))
	copy(leaves, ts.merkleTree.Leaves)
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].Hex() < leaves[j].Hex()
	})

	// Build Merkle Tree
	ts.merkleTree.Root = ts.buildMerkleTree(leaves)

	// Update metrics
	ts.metrics["merkle_trees_built"] = ts.metrics["merkle_trees_built"].(int) + 1
	ts.metrics["merkle_roots_generated"] = ts.metrics["merkle_roots_generated"].(int) + 1

	log.Debug("Merkle Tree rebuilt", 
		"events", len(ts.merkleTree.Events),
		"root", ts.merkleTree.Root.Hash.Hex())
}

// buildMerkleTree builds a Merkle Tree from sorted leaves
func (ts *TracingSystem) buildMerkleTree(leaves []common.Hash) *MerkleNode {
	if len(leaves) == 0 {
		return nil
	}

	if len(leaves) == 1 {
		return &MerkleNode{
			Hash: leaves[0],
			Data: leaves[0].Bytes(),
		}
	}

	// If odd number of leaves, duplicate the last one
	if len(leaves)%2 == 1 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}

	// Build next level
	var nextLevel []common.Hash
	for i := 0; i < len(leaves); i += 2 {
		left := leaves[i]
		right := leaves[i+1]
		
		// Combine left and right hashes
		combined := append(left.Bytes(), right.Bytes()...)
		hash := sha256.Sum256(combined)
		nextLevel = append(nextLevel, common.BytesToHash(hash[:]))
	}

	// Recursively build the tree
	root := ts.buildMerkleTree(nextLevel)
	
	// Set left and right children for visualization
	if len(leaves) >= 2 {
		root.Left = &MerkleNode{Hash: leaves[0]}
		root.Right = &MerkleNode{Hash: leaves[1]}
	}

	return root
}

// GetMerkleRoot returns the current Merkle root
func (ts *TracingSystem) GetMerkleRoot() common.Hash {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if ts.merkleTree.Root == nil {
		return common.Hash{}
	}

	return ts.merkleTree.Root.Hash
}

// VerifyEventInMerkleTree verifies if an event is in the Merkle Tree
func (ts *TracingSystem) VerifyEventInMerkleTree(event TraceEvent) bool {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	// Calculate event hash
	eventHash := ts.calculateEventHash(event)

	// Check if event exists in leaves
	for _, leaf := range ts.merkleTree.Leaves {
		if leaf == eventHash {
			return true
		}
	}

	return false
}

// GetMerkleProof returns the Merkle proof for an event
func (ts *TracingSystem) GetMerkleProof(event TraceEvent) ([]common.Hash, bool) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	eventHash := ts.calculateEventHash(event)
	
	// Find the event in leaves
	index := -1
	for i, leaf := range ts.merkleTree.Leaves {
		if leaf == eventHash {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, false
	}

	// Generate Merkle proof
	proof := ts.generateMerkleProof(index, ts.merkleTree.Leaves)
	return proof, true
}

// generateMerkleProof generates a Merkle proof for a given index
func (ts *TracingSystem) generateMerkleProof(index int, leaves []common.Hash) []common.Hash {
	var proof []common.Hash
	
	// Sort leaves for consistent ordering
	sortedLeaves := make([]common.Hash, len(leaves))
	copy(sortedLeaves, leaves)
	sort.Slice(sortedLeaves, func(i, j int) bool {
		return sortedLeaves[i].Hex() < sortedLeaves[j].Hex()
	})

	// Find the index in sorted leaves
	sortedIndex := -1
	for i, leaf := range sortedLeaves {
		if leaf == leaves[index] {
			sortedIndex = i
			break
		}
	}

	if sortedIndex == -1 {
		return proof
	}

	// Generate proof by traversing the tree
	currentLeaves := sortedLeaves
	currentIndex := sortedIndex

	for len(currentLeaves) > 1 {
		// If odd number of leaves, duplicate the last one
		if len(currentLeaves)%2 == 1 {
			currentLeaves = append(currentLeaves, currentLeaves[len(currentLeaves)-1])
		}

		// Determine sibling index
		siblingIndex := currentIndex ^ 1
		if siblingIndex < len(currentLeaves) {
			proof = append(proof, currentLeaves[siblingIndex])
		}

		// Move to parent level
		currentIndex = currentIndex / 2
		var nextLevel []common.Hash
		for i := 0; i < len(currentLeaves); i += 2 {
			left := currentLeaves[i]
			right := currentLeaves[i+1]
			combined := append(left.Bytes(), right.Bytes()...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, common.BytesToHash(hash[:]))
		}
		currentLeaves = nextLevel
	}

	return proof
}

// SetCurrentRound sets the current round for tracing
func (ts *TracingSystem) SetCurrentRound(round uint64) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.currentRound = round
	ts.metrics["current_round"] = round
}

// TraceRandomPOA traces random POA algorithm events
func (ts *TracingSystem) TraceRandomPOA(blockNumber uint64, signer common.Address, selectedSigner common.Address, seed int64, signers []common.Address) {
	data := map[string]interface{}{
		"block_number":     blockNumber,
		"signer":          signer.Hex(),
		"selected_signer": selectedSigner.Hex(),
		"seed":            seed,
		"signers_count":   len(signers),
		"is_selected":     signer == selectedSigner,
	}

	message := fmt.Sprintf("Random POA selection: %s selected from %d signers", 
		selectedSigner.Hex(), len(signers))

	ts.Trace(TraceEventRandomPOA, TraceLevelDetailed, blockNumber, signer, message, data)
}

// TraceLeaderSelection traces leader selection events
func (ts *TracingSystem) TraceLeaderSelection(blockNumber uint64, selectedLeader common.Address, allValidators []common.Address, selectionMethod string) {
	data := map[string]interface{}{
		"selected_leader":    selectedLeader.Hex(),
		"all_validators":     make([]string, len(allValidators)),
		"selection_method":   selectionMethod,
		"validator_count":    len(allValidators),
	}

	for i, addr := range allValidators {
		data["all_validators"].([]string)[i] = addr.Hex()
	}

	message := fmt.Sprintf("Leader selected: %s using %s method from %d validators", 
		selectedLeader.Hex(), selectionMethod, len(allValidators))

	ts.Trace(TraceEventLeaderSelection, TraceLevelBasic, blockNumber, selectedLeader, message, data)
}

// TraceBlockSigning traces block signing events
func (ts *TracingSystem) TraceBlockSigning(blockNumber uint64, signer common.Address, success bool, duration time.Duration, errorMsg string) {
	data := map[string]interface{}{
		"success":    success,
		"duration_ms": duration.Milliseconds(),
		"error":      errorMsg,
	}

	message := fmt.Sprintf("Block signing: %s %s block #%d in %v", 
		signer.Hex(), map[bool]string{true: "successfully signed", false: "failed to sign"}[success], blockNumber, duration)

	ts.TraceWithDuration(TraceEventBlockSigning, TraceLevelBasic, blockNumber, signer, message, data, duration)
}

// TraceTimeout traces timeout events
func (ts *TracingSystem) TraceTimeout(blockNumber uint64, validator common.Address, timeoutType string, expectedTime, actualTime time.Duration) {
	data := map[string]interface{}{
		"timeout_type":    timeoutType,
		"expected_time":   expectedTime.Milliseconds(),
		"actual_time":     actualTime.Milliseconds(),
		"delay":           (actualTime - expectedTime).Milliseconds(),
	}

	message := fmt.Sprintf("Timeout: %s %s timeout - expected %v, actual %v", 
		validator.Hex(), timeoutType, expectedTime, actualTime)

	ts.Trace(TraceEventTimeout, TraceLevelBasic, blockNumber, validator, message, data)
}

// TraceAccusation traces accusation events
func (ts *TracingSystem) TraceAccusation(blockNumber uint64, accuser, accused common.Address, accusationType string, evidence map[string]interface{}) {
	data := map[string]interface{}{
		"accuser":         accuser.Hex(),
		"accused":         accused.Hex(),
		"accusation_type": accusationType,
		"evidence":        evidence,
	}

	message := fmt.Sprintf("Accusation: %s accuses %s of %s", 
		accuser.Hex(), accused.Hex(), accusationType)

	ts.Trace(TraceEventAccusation, TraceLevelBasic, blockNumber, accuser, message, data)
}

// TraceAIGateEvaluation traces AI gate evaluation events
func (ts *TracingSystem) TraceAIGateEvaluation(blockNumber uint64, validator common.Address, evaluationResult map[string]interface{}, confidence float64) {
	data := map[string]interface{}{
		"evaluation_result": evaluationResult,
		"confidence":        confidence,
		"ai_model":          "validator_behavior_analyzer",
	}

	message := fmt.Sprintf("AI Gate Evaluation: %s evaluated with confidence %.2f", 
		validator.Hex(), confidence)

	ts.Trace(TraceEventAIGateEvaluation, TraceLevelDetailed, blockNumber, validator, message, data)
}

// GetTraceEvents returns trace events with optional filtering
func (ts *TracingSystem) GetTraceEvents(eventType TraceEventType, level TraceLevel, limit int) []TraceEvent {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	var filtered []TraceEvent
	for _, event := range ts.events {
		// Apply filters
		if eventType != "" && event.Type != eventType {
			continue
		}
		if level != TraceLevelOff && event.Level > level {
			continue
		}

		filtered = append(filtered, event)

		// Apply limit
		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// GetTraceMetrics returns current trace metrics
func (ts *TracingSystem) GetTraceMetrics() map[string]interface{} {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := make(map[string]interface{})
	for k, v := range ts.metrics {
		metricsCopy[k] = v
	}

	return metricsCopy
}

// GetTraceStats returns trace statistics
func (ts *TracingSystem) GetTraceStats() map[string]interface{} {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	stats := map[string]interface{}{
		"config": map[string]interface{}{
			"enable_tracing":        ts.config.EnableTracing,
			"trace_level":           ts.config.TraceLevel,
			"max_trace_events":      ts.config.MaxTraceEvents,
			"trace_retention":       ts.config.TraceRetention.String(),
			"enable_merkle_tree":    ts.config.EnableMerkleTree,
			"enable_persistence":    ts.config.EnablePersistence,
			"enable_metrics":        ts.config.EnableMetrics,
			"merkle_root_in_block":  ts.config.MerkleRootInBlock,
		},
		"current_events":     len(ts.events),
		"total_events":       ts.eventCount,
		"system_uptime":      time.Since(ts.startTime).String(),
		"current_round":      ts.currentRound,
		"merkle_root":        ts.GetMerkleRoot().Hex(),
		"merkle_tree_events": len(ts.merkleTree.Events),
		"metrics":            ts.metrics,
	}

	return stats
}

// updateMetrics updates the metrics based on the new event
func (ts *TracingSystem) updateMetrics(event TraceEvent) {
	// Update total events
	ts.metrics["total_events"] = ts.eventCount

	// Update events by type
	eventsByType := ts.metrics["events_by_type"].(map[string]int)
	eventsByType[string(event.Type)]++

	// Update events by level
	eventsByLevel := ts.metrics["events_by_level"].(map[string]int)
	levelName := fmt.Sprintf("level_%d", event.Level)
	eventsByLevel[levelName]++

	// Update duration metrics
	if event.Duration > 0 {
		currentMax := ts.metrics["max_duration"].(float64)
		currentMin := ts.metrics["min_duration"].(float64)
		
		durationMs := float64(event.Duration.Milliseconds())
		
		if durationMs > currentMax {
			ts.metrics["max_duration"] = durationMs
		}
		if currentMin == 0 || durationMs < currentMin {
			ts.metrics["min_duration"] = durationMs
		}

		// Update average duration
		totalEvents := float64(ts.eventCount)
		currentAvg := ts.metrics["average_duration"].(float64)
		ts.metrics["average_duration"] = (currentAvg*(totalEvents-1) + durationMs) / totalEvents
	}

	// Update system uptime
	ts.metrics["system_uptime"] = time.Since(ts.startTime)

	// Update last event time
	ts.metrics["last_event_time"] = event.Timestamp

	// Update events per minute
	uptimeMinutes := time.Since(ts.startTime).Minutes()
	if uptimeMinutes > 0 {
		ts.metrics["events_per_minute"] = float64(ts.eventCount) / uptimeMinutes
	}
}

// ClearTraceEvents clears all trace events
func (ts *TracingSystem) ClearTraceEvents() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.events = make([]TraceEvent, 0, ts.config.MaxTraceEvents)
	ts.merkleTree = &MerkleTree{
		Leaves: make([]common.Hash, 0),
		Events: make([]TraceEvent, 0),
	}
	ts.eventCount = 0
	ts.initializeMetrics()

	log.Info("Trace events and Merkle Tree cleared")
}

// ExportTraceEvents exports trace events and Merkle Tree to JSON
func (ts *TracingSystem) ExportTraceEvents() ([]byte, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	export := map[string]interface{}{
		"export_time":  time.Now(),
		"config":       ts.config,
		"events":       ts.events,
		"merkle_tree":  ts.merkleTree,
		"metrics":      ts.metrics,
		"merkle_root":  ts.GetMerkleRoot().Hex(),
	}

	return json.MarshalIndent(export, "", "  ")
}

// TraceAnomalyDetection traces anomaly detection events
func (ts *TracingSystem) TraceAnomalyDetection(anomalyType string, address common.Address, blockNumber uint64, severity string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["anomaly_type"] = anomalyType
	data["address"] = address.Hex()
	data["block_number"] = blockNumber
	data["severity"] = severity

	message := fmt.Sprintf("Anomaly detected: %s for %s at block %d (severity: %s)", 
		anomalyType, address.Hex(), blockNumber, severity)

	ts.Trace(TraceEventAnomalyDetection, TraceLevelBasic, blockNumber, address, message, data)
}

// TraceReputation traces reputation system events
func (ts *TracingSystem) TraceReputation(eventType string, address common.Address, blockNumber uint64, oldScore, newScore float64, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["event_type"] = eventType
	data["address"] = address.Hex()
	data["block_number"] = blockNumber
	data["old_score"] = oldScore
	data["new_score"] = newScore
	data["score_change"] = newScore - oldScore

	message := fmt.Sprintf("Reputation %s: %s score %.2f -> %.2f (change: %.2f)", 
		eventType, address.Hex(), oldScore, newScore, newScore-oldScore)

	ts.Trace(TraceEventReputation, TraceLevelDetailed, blockNumber, address, message, data)
}

// SetTraceLevel sets the trace level
func (ts *TracingSystem) SetTraceLevel(level TraceLevel) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.config.TraceLevel = level
	log.Info("Trace level updated", "level", level)
}

// EnableTracing enables or disables tracing
func (ts *TracingSystem) EnableTracing(enable bool) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.config.EnableTracing = enable
	log.Info("Tracing enabled/disabled", "enabled", enable)
}
