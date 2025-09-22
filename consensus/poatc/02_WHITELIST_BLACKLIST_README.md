# Whitelist/Blacklist Management for POA Consensus

This document describes the whitelist/blacklist management system integrated into the POA consensus engine.

## Overview

The whitelist/blacklist system provides fine-grained control over which addresses can participate in the consensus process. It operates alongside the existing POA signer management and provides additional security layers.

## Features

- **Whitelist Management**: Allow specific addresses to sign blocks
- **Blacklist Management**: Block specific addresses from signing blocks
- **Expiration Support**: Set expiration times for whitelist/blacklist entries
- **Persistence**: Automatic saving and loading of whitelist/blacklist data
- **Concurrent Access**: Thread-safe operations for multi-threaded environments
- **API Integration**: Full RPC API support for management operations

## Configuration

### WhitelistBlacklistConfig

```go
type WhitelistBlacklistConfig struct {
    EnableWhitelist bool   // Enable whitelist checking
    EnableBlacklist bool   // Enable blacklist checking
    WhitelistMode   bool   // Strict mode (true) or monitoring mode (false)
    PersistencePath string // Path to store whitelist/blacklist data
}
```

### Default Configuration

```go
func DefaultWhitelistBlacklistConfig() *WhitelistBlacklistConfig {
    return &WhitelistBlacklistConfig{
        EnableWhitelist: false, // Disabled by default
        EnableBlacklist: false, // Disabled by default
        WhitelistMode:   false, // Monitoring mode by default
        PersistencePath: "./whitelist_blacklist.json",
    }
}
```

## Modes

### 1. Monitoring Mode (WhitelistMode = false)
- Whitelist is used for monitoring and logging only
- Non-whitelisted signers can still sign blocks
- Logs warnings when non-whitelisted signers are detected

### 2. Strict Mode (WhitelistMode = true)
- Only whitelisted signers can sign blocks
- Non-whitelisted signers are rejected
- Provides strict access control

## API Endpoints

### Statistics
- `clique_getWhitelistBlacklistStats()` - Get statistics about whitelist and blacklist

### Whitelist Management
- `clique_getWhitelist()` - Get current whitelist
- `clique_addToWhitelist(address, addedBy, reason)` - Add address to whitelist
- `clique_removeFromWhitelist(address)` - Remove address from whitelist
- `clique_isWhitelisted(address)` - Check if address is whitelisted

### Blacklist Management
- `clique_getBlacklist()` - Get current blacklist
- `clique_addToBlacklist(address, addedBy, reason)` - Add address to blacklist
- `clique_removeFromBlacklist(address)` - Remove address from blacklist
- `clique_isBlacklisted(address)` - Check if address is blacklisted

### Validation
- `clique_validateSigner(address)` - Validate if signer is allowed to sign
- `clique_cleanupExpiredEntries()` - Remove expired entries

## Usage Examples

### Basic Operations

```bash
# Get whitelist/blacklist statistics
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_getWhitelistBlacklistStats","params":[],"id":1}' \
  http://localhost:8547

# Add address to whitelist
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_addToWhitelist","params":["0x1234567890123456789012345678901234567890","0x6519B747fC2c4DD4393843855Bef77f28875B07C","Test whitelist entry"],"id":1}' \
  http://localhost:8547

# Add address to blacklist
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_addToBlacklist","params":["0x2345678901234567890123456789012345678901","0x6519B747fC2c4DD4393843855Bef77f28875B07C","Test blacklist entry"],"id":1}' \
  http://localhost:8547

# Check if address is whitelisted
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_isWhitelisted","params":["0x1234567890123456789012345678901234567890"],"id":1}' \
  http://localhost:8547

# Validate signer
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"clique_validateSigner","params":["0x1234567890123456789012345678901234567890"],"id":1}' \
  http://localhost:8547
```

### PowerShell Testing

```powershell
# Run the whitelist/blacklist test script
.\test_whitelist_blacklist.ps1
```

## Data Structures

### WhitelistEntry

```go
type WhitelistEntry struct {
    Address     common.Address `json:"address"`
    AddedAt     time.Time      `json:"added_at"`
    AddedBy     common.Address `json:"added_by"`
    Reason      string         `json:"reason"`
    IsActive    bool           `json:"is_active"`
    ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}
```

### BlacklistEntry

```go
type BlacklistEntry struct {
    Address     common.Address `json:"address"`
    AddedAt     time.Time      `json:"added_at"`
    AddedBy     common.Address `json:"added_by"`
    Reason      string         `json:"reason"`
    IsActive    bool           `json:"is_active"`
    ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}
```

## Persistence

The whitelist/blacklist data is automatically saved to a JSON file specified in the configuration. The file includes:

- Whitelist entries
- Blacklist entries
- Configuration settings
- Last updated timestamp

### File Format

```json
{
  "whitelist": {
    "0x1234567890123456789012345678901234567890": {
      "address": "0x1234567890123456789012345678901234567890",
      "added_at": "2024-01-01T00:00:00Z",
      "added_by": "0x6519B747fC2c4DD4393843855Bef77f28875B07C",
      "reason": "Test whitelist entry",
      "is_active": true,
      "expires_at": null
    }
  },
  "blacklist": {},
  "config": {
    "enable_whitelist": true,
    "enable_blacklist": true,
    "whitelist_mode": false,
    "persistence_path": "./whitelist_blacklist.json"
  },
  "last_updated": "2024-01-01T00:00:00Z"
}
```

## Integration with POA

The whitelist/blacklist system is integrated into the POA consensus engine through the `verifySeal` function:

1. **Initialization**: The whitelist/blacklist manager is initialized when the first block is verified
2. **Validation**: Every signer is validated against the whitelist/blacklist before block acceptance
3. **Logging**: Validation failures are logged with appropriate severity levels
4. **Error Handling**: Invalid signers are rejected with descriptive error messages

## Security Considerations

1. **Blacklist Priority**: Blacklisted addresses are always rejected, even if whitelisted
2. **Expiration Handling**: Expired entries are automatically cleaned up
3. **Concurrent Access**: Thread-safe operations prevent race conditions
4. **Persistence Security**: File operations use atomic writes to prevent corruption

## Testing

The system includes comprehensive unit tests covering:

- Basic operations (add, remove, check)
- Expiration handling
- Concurrent access
- Error handling
- Persistence
- Integration with POA consensus

Run tests with:

```bash
go test ./consensus/clique -v -run TestWhitelistBlacklist
```

## Performance

- **Memory Usage**: Minimal overhead with efficient map-based storage
- **CPU Usage**: O(1) lookup time for validation operations
- **Disk Usage**: JSON persistence with automatic cleanup of expired entries
- **Network**: RPC API calls for remote management

## Troubleshooting

### Common Issues

1. **Manager Not Initialized**: Ensure the whitelist/blacklist manager is properly initialized
2. **Permission Denied**: Check file permissions for persistence path
3. **Invalid Addresses**: Verify address format (0x prefix, 40 hex characters)
4. **Expired Entries**: Use cleanup API to remove expired entries

### Debug Logging

Enable debug logging to see detailed whitelist/blacklist operations:

```bash
--verbosity 5
```

## Future Enhancements

- **Role-based Access Control**: Different permission levels for whitelist/blacklist management
- **Audit Logging**: Detailed logging of all whitelist/blacklist operations
- **Bulk Operations**: Support for adding/removing multiple addresses at once
- **Webhook Integration**: Notifications for whitelist/blacklist changes
- **Metrics Integration**: Prometheus metrics for monitoring

## License

This implementation follows the same licensing terms as the go-ethereum project.
