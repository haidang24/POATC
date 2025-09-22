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
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestWhitelistBlacklistManagerBasic(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false, // Monitoring mode
		PersistencePath: tempDir + "/whitelist_blacklist.json",
	}

	manager := NewWhitelistBlacklistManager(config)

	// Test addresses
	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Test adding to whitelist
	err := manager.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	// Test checking whitelist
	if !manager.IsWhitelisted(addr1) {
		t.Error("Address should be whitelisted")
	}

	if manager.IsWhitelisted(addr2) {
		t.Error("Address should not be whitelisted")
	}

	// Test adding to blacklist
	err = manager.AddToBlacklist(addr2, admin, "Test blacklist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to blacklist: %v", err)
	}

	// Test checking blacklist
	if !manager.IsBlacklisted(addr2) {
		t.Error("Address should be blacklisted")
	}

	if manager.IsBlacklisted(addr1) {
		t.Error("Address should not be blacklisted")
	}

	// Test validation
	valid, reason := manager.ValidateSigner(addr1)
	if !valid {
		t.Errorf("Whitelisted address should be valid: %s", reason)
	}

	valid, reason = manager.ValidateSigner(addr2)
	if valid {
		t.Error("Blacklisted address should not be valid")
	}
	if reason == "" {
		t.Error("Should provide reason for blacklisted address")
	}
}

func TestWhitelistBlacklistManagerStrictMode(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: false,
		WhitelistMode:   true, // Strict mode
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Add addr1 to whitelist
	err := manager.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	// Test validation in strict mode
	valid, reason := manager.ValidateSigner(addr1)
	if !valid {
		t.Errorf("Whitelisted address should be valid: %s", reason)
	}

	valid, reason = manager.ValidateSigner(addr2)
	if valid {
		t.Error("Non-whitelisted address should not be valid in strict mode")
	}
	if reason == "" {
		t.Error("Should provide reason for non-whitelisted address")
	}
}

func TestWhitelistBlacklistManagerExpiration(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Add with expiration in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	err := manager.AddToWhitelist(addr1, admin, "Expired entry", &pastTime)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	// Should not be considered whitelisted due to expiration
	if manager.IsWhitelisted(addr1) {
		t.Error("Expired address should not be whitelisted")
	}

	// Add with expiration in the future
	futureTime := time.Now().Add(1 * time.Hour)
	err = manager.AddToWhitelist(addr1, admin, "Future entry", &futureTime)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	// Should be considered whitelisted
	if !manager.IsWhitelisted(addr1) {
		t.Error("Non-expired address should be whitelisted")
	}
}

func TestWhitelistBlacklistManagerRemoval(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Add to whitelist
	err := manager.AddToWhitelist(addr1, admin, "Test entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	if !manager.IsWhitelisted(addr1) {
		t.Error("Address should be whitelisted")
	}

	// Remove from whitelist
	err = manager.RemoveFromWhitelist(addr1)
	if err != nil {
		t.Fatalf("Failed to remove from whitelist: %v", err)
	}

	if manager.IsWhitelisted(addr1) {
		t.Error("Address should not be whitelisted after removal")
	}

	// Add to blacklist
	err = manager.AddToBlacklist(addr1, admin, "Test blacklist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to blacklist: %v", err)
	}

	if !manager.IsBlacklisted(addr1) {
		t.Error("Address should be blacklisted")
	}

	// Remove from blacklist
	err = manager.RemoveFromBlacklist(addr1)
	if err != nil {
		t.Fatalf("Failed to remove from blacklist: %v", err)
	}

	if manager.IsBlacklisted(addr1) {
		t.Error("Address should not be blacklisted after removal")
	}
}

func TestWhitelistBlacklistManagerBlacklistOverridesWhitelist(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Add to whitelist first
	err := manager.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	if !manager.IsWhitelisted(addr1) {
		t.Error("Address should be whitelisted")
	}

	// Add to blacklist (should remove from whitelist)
	err = manager.AddToBlacklist(addr1, admin, "Test blacklist entry", nil)
	if err != nil {
		t.Fatalf("Failed to add to blacklist: %v", err)
	}

	// Should be blacklisted and not whitelisted
	if !manager.IsBlacklisted(addr1) {
		t.Error("Address should be blacklisted")
	}

	if manager.IsWhitelisted(addr1) {
		t.Error("Address should not be whitelisted after being blacklisted")
	}

	// Validation should fail
	valid, reason := manager.ValidateSigner(addr1)
	if valid {
		t.Error("Blacklisted address should not be valid")
	}
	if reason == "" {
		t.Error("Should provide reason for blacklisted address")
	}
}

func TestWhitelistBlacklistManagerStats(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Add entries
	manager.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	manager.AddToBlacklist(addr2, admin, "Test blacklist entry", nil)

	// Get stats
	stats := manager.GetStats()

	// Check config
	configData, ok := stats["config"].(map[string]interface{})
	if !ok {
		t.Fatal("Stats should contain config")
	}

	if configData["enable_whitelist"] != true {
		t.Error("Config should show whitelist enabled")
	}

	if configData["enable_blacklist"] != true {
		t.Error("Config should show blacklist enabled")
	}

	// Check whitelist stats
	whitelistData, ok := stats["whitelist"].(map[string]interface{})
	if !ok {
		t.Fatal("Stats should contain whitelist data")
	}

	if whitelistData["total"] != 1 {
		t.Errorf("Expected 1 whitelist entry, got %v", whitelistData["total"])
	}

	if whitelistData["active"] != 1 {
		t.Errorf("Expected 1 active whitelist entry, got %v", whitelistData["active"])
	}

	// Check blacklist stats
	blacklistData, ok := stats["blacklist"].(map[string]interface{})
	if !ok {
		t.Fatal("Stats should contain blacklist data")
	}

	if blacklistData["total"] != 1 {
		t.Errorf("Expected 1 blacklist entry, got %v", blacklistData["total"])
	}

	if blacklistData["active"] != 1 {
		t.Errorf("Expected 1 active blacklist entry, got %v", blacklistData["active"])
	}
}

func TestWhitelistBlacklistManagerPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: tempDir + "/whitelist_blacklist.json",
	}

	// Create first manager and add some entries
	manager1 := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	manager1.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	manager1.AddToBlacklist(addr2, admin, "Test blacklist entry", nil)

	// Create second manager with same config (should load from persistence)
	manager2 := NewWhitelistBlacklistManager(config)

	// Check if data was loaded
	if !manager2.IsWhitelisted(addr1) {
		t.Error("Address should be whitelisted after loading from persistence")
	}

	if !manager2.IsBlacklisted(addr2) {
		t.Error("Address should be blacklisted after loading from persistence")
	}

	// Check stats
	stats := manager2.GetStats()
	whitelistData := stats["whitelist"].(map[string]interface{})
	blacklistData := stats["blacklist"].(map[string]interface{})

	if whitelistData["total"] != 1 {
		t.Errorf("Expected 1 whitelist entry after loading, got %v", whitelistData["total"])
	}

	if blacklistData["total"] != 1 {
		t.Errorf("Expected 1 blacklist entry after loading, got %v", blacklistData["total"])
	}
}

func TestWhitelistBlacklistManagerCleanupExpired(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")
	addr3 := common.HexToAddress("0x3456789012345678901234567890123456789012")
	addr4 := common.HexToAddress("0x4567890123456789012345678901234567890123")
	admin := common.HexToAddress("0x5678901234567890123456789012345678901234")

	// Add expired entries
	pastTime := time.Now().Add(-1 * time.Hour)
	manager.AddToWhitelist(addr1, admin, "Expired whitelist", &pastTime)
	manager.AddToBlacklist(addr2, admin, "Expired blacklist", &pastTime)

	// Add non-expired entries
	futureTime := time.Now().Add(1 * time.Hour)
	manager.AddToWhitelist(addr3, admin, "Future whitelist", &futureTime)
	manager.AddToBlacklist(addr4, admin, "Future blacklist", &futureTime)

	// Check stats before cleanup
	stats := manager.GetStats()
	whitelistData := stats["whitelist"].(map[string]interface{})
	blacklistData := stats["blacklist"].(map[string]interface{})

	if whitelistData["total"] != 2 {
		t.Errorf("Expected 2 whitelist entries before cleanup, got %v", whitelistData["total"])
	}

	if blacklistData["total"] != 2 {
		t.Errorf("Expected 2 blacklist entries before cleanup, got %v", blacklistData["total"])
	}

	// Cleanup expired entries
	manager.CleanupExpiredEntries()

	// Check stats after cleanup
	stats = manager.GetStats()
	whitelistData = stats["whitelist"].(map[string]interface{})
	blacklistData = stats["blacklist"].(map[string]interface{})

	if whitelistData["total"] != 1 {
		t.Errorf("Expected 1 whitelist entry after cleanup, got %v", whitelistData["total"])
	}

	if blacklistData["total"] != 1 {
		t.Errorf("Expected 1 blacklist entry after cleanup, got %v", blacklistData["total"])
	}
}

func TestWhitelistBlacklistManagerErrorHandling(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Test removing non-existent entry
	err := manager.RemoveFromWhitelist(addr1)
	if err == nil {
		t.Error("Should return error when removing non-existent whitelist entry")
	}

	err = manager.RemoveFromBlacklist(addr1)
	if err == nil {
		t.Error("Should return error when removing non-existent blacklist entry")
	}

	// Test adding to whitelist when already in blacklist
	manager.AddToBlacklist(addr1, admin, "Test blacklist entry", nil)
	err = manager.AddToWhitelist(addr1, admin, "Test whitelist entry", nil)
	if err == nil {
		t.Error("Should return error when adding blacklisted address to whitelist")
	}
}

func TestWhitelistBlacklistManagerConcurrentAccess(t *testing.T) {
	config := &WhitelistBlacklistConfig{
		EnableWhitelist: true,
		EnableBlacklist: true,
		WhitelistMode:   false,
		PersistencePath: "",
	}

	manager := NewWhitelistBlacklistManager(config)

	admin := common.HexToAddress("0x3456789012345678901234567890123456789012")

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			addr := common.HexToAddress("0x" + string(rune('0'+i)) + "234567890123456789012345678901234567890")
			
			// Add to whitelist
			manager.AddToWhitelist(addr, admin, "Concurrent test", nil)
			
			// Check if whitelisted
			manager.IsWhitelisted(addr)
			
			// Remove from whitelist
			manager.RemoveFromWhitelist(addr)
			
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Final stats should be clean
	stats := manager.GetStats()
	whitelistData := stats["whitelist"].(map[string]interface{})
	blacklistData := stats["blacklist"].(map[string]interface{})

	if whitelistData["total"] != 0 {
		t.Errorf("Expected 0 whitelist entries after concurrent test, got %v", whitelistData["total"])
	}

	if blacklistData["total"] != 0 {
		t.Errorf("Expected 0 blacklist entries after concurrent test, got %v", blacklistData["total"])
	}
}
