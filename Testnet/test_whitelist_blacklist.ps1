# Script to test Whitelist/Blacklist functionality
Write-Host "=== Testing Whitelist/Blacklist Functionality ===" -ForegroundColor Magenta
Write-Host ""

# Function to make RPC calls
function Invoke-RPC {
    param(
        [string]$Url,
        [string]$Method,
        [array]$Params = @(),
        [int]$Id = 1
    )
    
    $body = @{
        jsonrpc = "2.0"
        method = $Method
        params = $Params
        id = $Id
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method Post -Body $body -ContentType "application/json"
        return $response
    } catch {
        Write-Host "RPC Error: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Test both nodes
$node1Url = "http://localhost:8547"
$node2Url = "http://localhost:8549"

Write-Host "Testing Node1 (Port 8547)..." -ForegroundColor Green

# Test 1: Get Whitelist/Blacklist Stats
Write-Host "`n1. Testing Whitelist/Blacklist Stats..." -ForegroundColor Yellow
$stats = Invoke-RPC -Url $node1Url -Method "clique_getWhitelistBlacklistStats"
if ($stats) {
    Write-Host "Whitelist/Blacklist Stats:" -ForegroundColor Cyan
    $stats.result | ConvertTo-Json -Depth 3 | Write-Host
} else {
    Write-Host "Failed to get whitelist/blacklist stats" -ForegroundColor Red
}

# Test 2: Get Current Whitelist
Write-Host "`n2. Testing Get Whitelist..." -ForegroundColor Yellow
$whitelist = Invoke-RPC -Url $node1Url -Method "clique_getWhitelist"
if ($whitelist) {
    Write-Host "Current Whitelist:" -ForegroundColor Cyan
    $whitelist.result | ConvertTo-Json -Depth 3 | Write-Host
} else {
    Write-Host "Failed to get whitelist" -ForegroundColor Red
}

# Test 3: Get Current Blacklist
Write-Host "`n3. Testing Get Blacklist..." -ForegroundColor Yellow
$blacklist = Invoke-RPC -Url $node1Url -Method "clique_getBlacklist"
if ($blacklist) {
    Write-Host "Current Blacklist:" -ForegroundColor Cyan
    $blacklist.result | ConvertTo-Json -Depth 3 | Write-Host
} else {
    Write-Host "Failed to get blacklist" -ForegroundColor Red
}

# Test 4: Add to Whitelist
Write-Host "`n4. Testing Add to Whitelist..." -ForegroundColor Yellow
$testAddress = "0x1234567890123456789012345678901234567890"
$adminAddress = "0x6519B747fC2c4DD4393843855Bef77f28875B07C"
$addToWhitelist = Invoke-RPC -Url $node1Url -Method "clique_addToWhitelist" -Params @($testAddress, $adminAddress, "Test whitelist entry")
if ($addToWhitelist) {
    if ($addToWhitelist.error) {
        Write-Host "Error adding to whitelist: $($addToWhitelist.error.message)" -ForegroundColor Red
    } else {
        Write-Host "Successfully added $testAddress to whitelist" -ForegroundColor Green
    }
} else {
    Write-Host "Failed to add to whitelist" -ForegroundColor Red
}

# Test 5: Check if Address is Whitelisted
Write-Host "`n5. Testing Is Whitelisted..." -ForegroundColor Yellow
$isWhitelisted = Invoke-RPC -Url $node1Url -Method "clique_isWhitelisted" -Params @($testAddress)
if ($isWhitelisted) {
    if ($isWhitelisted.result) {
        Write-Host "$testAddress is whitelisted" -ForegroundColor Green
    } else {
        Write-Host "$testAddress is not whitelisted" -ForegroundColor Yellow
    }
} else {
    Write-Host "Failed to check whitelist status" -ForegroundColor Red
}

# Test 6: Add to Blacklist
Write-Host "`n6. Testing Add to Blacklist..." -ForegroundColor Yellow
$blacklistAddress = "0x2345678901234567890123456789012345678901"
$addToBlacklist = Invoke-RPC -Url $node1Url -Method "clique_addToBlacklist" -Params @($blacklistAddress, $adminAddress, "Test blacklist entry")
if ($addToBlacklist) {
    if ($addToBlacklist.error) {
        Write-Host "Error adding to blacklist: $($addToBlacklist.error.message)" -ForegroundColor Red
    } else {
        Write-Host "Successfully added $blacklistAddress to blacklist" -ForegroundColor Green
    }
} else {
    Write-Host "Failed to add to blacklist" -ForegroundColor Red
}

# Test 7: Check if Address is Blacklisted
Write-Host "`n7. Testing Is Blacklisted..." -ForegroundColor Yellow
$isBlacklisted = Invoke-RPC -Url $node1Url -Method "clique_isBlacklisted" -Params @($blacklistAddress)
if ($isBlacklisted) {
    if ($isBlacklisted.result) {
        Write-Host "$blacklistAddress is blacklisted" -ForegroundColor Green
    } else {
        Write-Host "$blacklistAddress is not blacklisted" -ForegroundColor Yellow
    }
} else {
    Write-Host "Failed to check blacklist status" -ForegroundColor Red
}

# Test 8: Validate Signer
Write-Host "`n8. Testing Validate Signer..." -ForegroundColor Yellow
$validateSigner = Invoke-RPC -Url $node1Url -Method "clique_validateSigner" -Params @($testAddress)
if ($validateSigner) {
    if ($validateSigner.result) {
        Write-Host "Signer $testAddress is valid" -ForegroundColor Green
    } else {
        Write-Host "Signer $testAddress is not valid: $($validateSigner.result)" -ForegroundColor Red
    }
} else {
    Write-Host "Failed to validate signer" -ForegroundColor Red
}

# Test 9: Get Updated Stats
Write-Host "`n9. Testing Updated Stats..." -ForegroundColor Yellow
$updatedStats = Invoke-RPC -Url $node1Url -Method "clique_getWhitelistBlacklistStats"
if ($updatedStats) {
    Write-Host "Updated Whitelist/Blacklist Stats:" -ForegroundColor Cyan
    $updatedStats.result | ConvertTo-Json -Depth 3 | Write-Host
} else {
    Write-Host "Failed to get updated stats" -ForegroundColor Red
}

# Test 10: Cleanup Expired Entries
Write-Host "`n10. Testing Cleanup Expired Entries..." -ForegroundColor Yellow
$cleanup = Invoke-RPC -Url $node1Url -Method "clique_cleanupExpiredEntries"
if ($cleanup) {
    if ($cleanup.error) {
        Write-Host "Error during cleanup: $($cleanup.error.message)" -ForegroundColor Red
    } else {
        Write-Host "Successfully cleaned up expired entries" -ForegroundColor Green
    }
} else {
    Write-Host "Failed to cleanup expired entries" -ForegroundColor Red
}

Write-Host "`n=== Whitelist/Blacklist Testing Completed ===" -ForegroundColor Magenta
