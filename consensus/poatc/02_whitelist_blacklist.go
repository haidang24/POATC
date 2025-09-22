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
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// WhitelistBlacklistConfig contains configuration for whitelist/blacklist management
type WhitelistBlacklistConfig struct {
	EnableWhitelist bool   `json:"enable_whitelist"` // Enable whitelist checking
	EnableBlacklist bool   `json:"enable_blacklist"` // Enable blacklist checking
	WhitelistMode   bool   `json:"whitelist_mode"`   // If true, only whitelisted signers can sign; if false, whitelist is just for monitoring
	PersistencePath string `json:"persistence_path"` // Path to store whitelist/blacklist data
}

// DefaultWhitelistBlacklistConfig returns a default configuration
func DefaultWhitelistBlacklistConfig() *WhitelistBlacklistConfig {
	return &WhitelistBlacklistConfig{
		EnableWhitelist: false, // Disabled by default
		EnableBlacklist: false, // Disabled by default
		WhitelistMode:   false, // Monitoring mode by default
		PersistencePath: "./whitelist_blacklist.json",
	}
}

// WhitelistBlacklistManager handles whitelist and blacklist management
type WhitelistBlacklistManager struct {
	config          *WhitelistBlacklistConfig
	whitelist       map[common.Address]WhitelistEntry
	blacklist       map[common.Address]BlacklistEntry
	mutex           sync.RWMutex
	persistencePath string
}

// WhitelistEntry represents an entry in the whitelist
type WhitelistEntry struct {
	Address   common.Address `json:"address"`
	AddedAt   time.Time      `json:"added_at"`
	AddedBy   common.Address `json:"added_by"`
	Reason    string         `json:"reason"`
	IsActive  bool           `json:"is_active"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

// BlacklistEntry represents an entry in the blacklist
type BlacklistEntry struct {
	Address   common.Address `json:"address"`
	AddedAt   time.Time      `json:"added_at"`
	AddedBy   common.Address `json:"added_by"`
	Reason    string         `json:"reason"`
	IsActive  bool           `json:"is_active"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

// PersistenceData represents the data structure for persistence
type PersistenceData struct {
	Whitelist   map[common.Address]WhitelistEntry `json:"whitelist"`
	Blacklist   map[common.Address]BlacklistEntry `json:"blacklist"`
	Config      *WhitelistBlacklistConfig         `json:"config"`
	LastUpdated time.Time                         `json:"last_updated"`
}

// NewWhitelistBlacklistManager creates a new whitelist/blacklist manager
func NewWhitelistBlacklistManager(config *WhitelistBlacklistConfig) *WhitelistBlacklistManager {
	if config == nil {
		config = DefaultWhitelistBlacklistConfig()
	}

	manager := &WhitelistBlacklistManager{
		config:          config,
		whitelist:       make(map[common.Address]WhitelistEntry),
		blacklist:       make(map[common.Address]BlacklistEntry),
		persistencePath: config.PersistencePath,
	}

	// Load existing data if available
	manager.loadFromPersistence()

	return manager
}

// AddToWhitelist adds an address to the whitelist
func (wbm *WhitelistBlacklistManager) AddToWhitelist(address common.Address, addedBy common.Address, reason string, expiresAt *time.Time) error {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	// Check if address is in blacklist
	if _, exists := wbm.blacklist[address]; exists {
		return fmt.Errorf("address %s is in blacklist, cannot add to whitelist", address.Hex())
	}

	entry := WhitelistEntry{
		Address:   address,
		AddedAt:   time.Now(),
		AddedBy:   addedBy,
		Reason:    reason,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	wbm.whitelist[address] = entry
	log.Info("Address added to whitelist", "address", address.Hex(), "added_by", addedBy.Hex(), "reason", reason)

	// Save to persistence
	return wbm.saveToPersistence()
}

// RemoveFromWhitelist removes an address from the whitelist
func (wbm *WhitelistBlacklistManager) RemoveFromWhitelist(address common.Address) error {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	if _, exists := wbm.whitelist[address]; !exists {
		return fmt.Errorf("address %s not found in whitelist", address.Hex())
	}

	delete(wbm.whitelist, address)
	log.Info("Address removed from whitelist", "address", address.Hex())

	// Save to persistence
	return wbm.saveToPersistence()
}

// AddToBlacklist adds an address to the blacklist
func (wbm *WhitelistBlacklistManager) AddToBlacklist(address common.Address, addedBy common.Address, reason string, expiresAt *time.Time) error {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	// Remove from whitelist if exists
	if _, exists := wbm.whitelist[address]; exists {
		delete(wbm.whitelist, address)
		log.Info("Address removed from whitelist due to blacklist addition", "address", address.Hex())
	}

	entry := BlacklistEntry{
		Address:   address,
		AddedAt:   time.Now(),
		AddedBy:   addedBy,
		Reason:    reason,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	wbm.blacklist[address] = entry
	log.Info("Address added to blacklist", "address", address.Hex(), "added_by", addedBy.Hex(), "reason", reason)

	// Save to persistence
	return wbm.saveToPersistence()
}

// RemoveFromBlacklist removes an address from the blacklist
func (wbm *WhitelistBlacklistManager) RemoveFromBlacklist(address common.Address) error {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	if _, exists := wbm.blacklist[address]; !exists {
		return fmt.Errorf("address %s not found in blacklist", address.Hex())
	}

	delete(wbm.blacklist, address)
	log.Info("Address removed from blacklist", "address", address.Hex())

	// Save to persistence
	return wbm.saveToPersistence()
}

// IsWhitelisted checks if an address is in the whitelist
func (wbm *WhitelistBlacklistManager) IsWhitelisted(address common.Address) bool {
	wbm.mutex.RLock()
	defer wbm.mutex.RUnlock()

	entry, exists := wbm.whitelist[address]
	if !exists || !entry.IsActive {
		return false
	}

	// Check expiration
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		return false
	}

	return true
}

// IsBlacklisted checks if an address is in the blacklist
func (wbm *WhitelistBlacklistManager) IsBlacklisted(address common.Address) bool {
	wbm.mutex.RLock()
	defer wbm.mutex.RUnlock()

	entry, exists := wbm.blacklist[address]
	if !exists || !entry.IsActive {
		return false
	}

	// Check expiration
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		return false
	}

	return true
}

// ValidateSigner validates if a signer is allowed to sign based on whitelist/blacklist rules
func (wbm *WhitelistBlacklistManager) ValidateSigner(address common.Address) (bool, string) {
	// Check blacklist first
	if wbm.config.EnableBlacklist && wbm.IsBlacklisted(address) {
		return false, fmt.Sprintf("signer %s is blacklisted", address.Hex())
	}

	// Check whitelist if enabled
	if wbm.config.EnableWhitelist {
		if wbm.config.WhitelistMode {
			// Strict mode: only whitelisted signers can sign
			if !wbm.IsWhitelisted(address) {
				return false, fmt.Sprintf("signer %s is not whitelisted", address.Hex())
			}
		} else {
			// Monitoring mode: log if not whitelisted but allow signing
			if !wbm.IsWhitelisted(address) {
				log.Warn("Non-whitelisted signer detected", "address", address.Hex())
			}
		}
	}

	return true, ""
}

// GetWhitelist returns a copy of the whitelist
func (wbm *WhitelistBlacklistManager) GetWhitelist() map[common.Address]WhitelistEntry {
	wbm.mutex.RLock()
	defer wbm.mutex.RUnlock()

	result := make(map[common.Address]WhitelistEntry)
	for addr, entry := range wbm.whitelist {
		result[addr] = entry
	}
	return result
}

// GetBlacklist returns a copy of the blacklist
func (wbm *WhitelistBlacklistManager) GetBlacklist() map[common.Address]BlacklistEntry {
	wbm.mutex.RLock()
	defer wbm.mutex.RUnlock()

	result := make(map[common.Address]BlacklistEntry)
	for addr, entry := range wbm.blacklist {
		result[addr] = entry
	}
	return result
}

// GetStats returns statistics about whitelist and blacklist
func (wbm *WhitelistBlacklistManager) GetStats() map[string]interface{} {
	wbm.mutex.RLock()
	defer wbm.mutex.RUnlock()

	activeWhitelist := 0
	activeBlacklist := 0
	expiredWhitelist := 0
	expiredBlacklist := 0

	now := time.Now()

	for _, entry := range wbm.whitelist {
		if entry.IsActive {
			if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
				expiredWhitelist++
			} else {
				activeWhitelist++
			}
		}
	}

	for _, entry := range wbm.blacklist {
		if entry.IsActive {
			if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
				expiredBlacklist++
			} else {
				activeBlacklist++
			}
		}
	}

	return map[string]interface{}{
		"config": map[string]interface{}{
			"enable_whitelist": wbm.config.EnableWhitelist,
			"enable_blacklist": wbm.config.EnableBlacklist,
			"whitelist_mode":   wbm.config.WhitelistMode,
		},
		"whitelist": map[string]interface{}{
			"total":   len(wbm.whitelist),
			"active":  activeWhitelist,
			"expired": expiredWhitelist,
		},
		"blacklist": map[string]interface{}{
			"total":   len(wbm.blacklist),
			"active":  activeBlacklist,
			"expired": expiredBlacklist,
		},
	}
}

// CleanupExpiredEntries removes expired entries from both lists
func (wbm *WhitelistBlacklistManager) CleanupExpiredEntries() {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	// Cleanup expired whitelist entries
	for addr, entry := range wbm.whitelist {
		if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			delete(wbm.whitelist, addr)
			cleaned++
		}
	}

	// Cleanup expired blacklist entries
	for addr, entry := range wbm.blacklist {
		if entry.ExpiresAt != nil && now.After(*entry.ExpiresAt) {
			delete(wbm.blacklist, addr)
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Info("Cleaned up expired whitelist/blacklist entries", "count", cleaned)
		wbm.saveToPersistence()
	}
}

// saveToPersistence saves the current state to disk
func (wbm *WhitelistBlacklistManager) saveToPersistence() error {
	if wbm.persistencePath == "" {
		return nil // No persistence configured
	}

	data := PersistenceData{
		Whitelist:   wbm.whitelist,
		Blacklist:   wbm.blacklist,
		Config:      wbm.config,
		LastUpdated: time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal whitelist/blacklist data: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(wbm.persistencePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write to temporary file first, then rename (atomic operation)
	tempPath := wbm.persistencePath + ".tmp"
	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write whitelist/blacklist data: %v", err)
	}

	if err := os.Rename(tempPath, wbm.persistencePath); err != nil {
		return fmt.Errorf("failed to rename whitelist/blacklist file: %v", err)
	}

	return nil
}

// loadFromPersistence loads the state from disk
func (wbm *WhitelistBlacklistManager) loadFromPersistence() error {
	if wbm.persistencePath == "" {
		return nil // No persistence configured
	}

	// Check if file exists
	if _, err := os.Stat(wbm.persistencePath); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty lists
	}

	jsonData, err := os.ReadFile(wbm.persistencePath)
	if err != nil {
		return fmt.Errorf("failed to read whitelist/blacklist data: %v", err)
	}

	var data PersistenceData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal whitelist/blacklist data: %v", err)
	}

	// Update the manager's data
	wbm.whitelist = data.Whitelist
	wbm.blacklist = data.Blacklist
	if data.Config != nil {
		wbm.config = data.Config
	}

	log.Info("Loaded whitelist/blacklist from persistence",
		"whitelist_count", len(wbm.whitelist),
		"blacklist_count", len(wbm.blacklist),
		"last_updated", data.LastUpdated)

	return nil
}

// UpdateConfig updates the configuration
func (wbm *WhitelistBlacklistManager) UpdateConfig(config *WhitelistBlacklistConfig) error {
	wbm.mutex.Lock()
	defer wbm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	wbm.config = config
	wbm.persistencePath = config.PersistencePath

	log.Info("Whitelist/Blacklist configuration updated",
		"enable_whitelist", config.EnableWhitelist,
		"enable_blacklist", config.EnableBlacklist,
		"whitelist_mode", config.WhitelistMode)

	return wbm.saveToPersistence()
}
