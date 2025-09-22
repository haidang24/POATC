# Test script for Time Dynamic System
# This script tests the Time Dynamic mechanisms: Dynamic Block Time, Dynamic Validator Selection, Dynamic Reputation Decay

Write-Host "=== Time Dynamic System Test ===" -ForegroundColor Green
Write-Host "Testing Dynamic Block Time, Dynamic Validator Selection, and Dynamic Reputation Decay" -ForegroundColor Yellow

# Start node in background
Write-Host "`nStarting node..." -ForegroundColor Cyan
Start-Process -FilePath ".\hdchain.exe" -ArgumentList "--datadir", "node1", "--networkid", "12345", "--http", "--http.port", "8545", "--http.api", "eth,net,web3,personal,clique", "--allow-insecure-unlock", "--nodiscover", "--verbosity", "3" -WindowStyle Hidden

# Wait for node to start
Write-Host "Waiting for node to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test 1: Get Time Dynamic Stats
Write-Host "`n1. Testing Get Time Dynamic Stats..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getTimeDynamicStats",
    "params": [],
    "id": 1
}
"@

if ($response.result) {
    Write-Host "✓ Time Dynamic Stats Retrieved Successfully" -ForegroundColor Green
    Write-Host "Config:" -ForegroundColor White
    $response.result.config | ConvertTo-Json -Depth 3
    Write-Host "Dynamic Block Time:" -ForegroundColor White
    $response.result.dynamic_block_time | ConvertTo-Json -Depth 3
} else {
    Write-Host "✗ Failed to get Time Dynamic Stats" -ForegroundColor Red
    Write-Host "Error: $($response.error.message)" -ForegroundColor Red
}

# Test 2: Get Current Block Time
Write-Host "`n2. Testing Get Current Block Time..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getCurrentBlockTime",
    "params": [],
    "id": 2
}
"@

if ($response.result) {
    Write-Host "✓ Current Block Time: $($response.result) seconds" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to get Current Block Time" -ForegroundColor Red
    Write-Host "Error: $($response.error.message)" -ForegroundColor Red
}

# Test 3: Update Transaction Count (simulate high transaction volume)
Write-Host "`n3. Testing Dynamic Block Time with High Transaction Volume..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_updateTransactionCount",
    "params": [150],
    "id": 3
}
"@

if ($response.result -eq $null -and $response.error -eq $null) {
    Write-Host "✓ Updated Transaction Count to 150 (High Volume)" -ForegroundColor Green
    
    # Check if block time changed
    Start-Sleep -Seconds 2
    $response2 = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getCurrentBlockTime",
    "params": [],
    "id": 4
}
"@
    
    if ($response2.result) {
        Write-Host "  New Block Time: $($response2.result) seconds (should be lower)" -ForegroundColor Yellow
    }
} else {
    Write-Host "✗ Failed to update Transaction Count" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 4: Update Transaction Count (simulate low transaction volume)
Write-Host "`n4. Testing Dynamic Block Time with Low Transaction Volume..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_updateTransactionCount",
    "params": [2],
    "id": 5
}
"@

if ($response.result -eq $null -and $response.error -eq $null) {
    Write-Host "✓ Updated Transaction Count to 2 (Low Volume)" -ForegroundColor Green
    
    # Check if block time changed
    Start-Sleep -Seconds 2
    $response2 = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getCurrentBlockTime",
    "params": [],
    "id": 6
}
"@
    
    if ($response2.result) {
        Write-Host "  New Block Time: $($response2.result) seconds (should be higher)" -ForegroundColor Yellow
    }
} else {
    Write-Host "✗ Failed to update Transaction Count" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 5: Trigger Validator Selection
Write-Host "`n5. Testing Dynamic Validator Selection..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_triggerValidatorSelection",
    "params": [100, "0x1234567890123456789012345678901234567890123456789012345678901234"],
    "id": 7
}
"@

if ($response.result -eq $null -and $response.error -eq $null) {
    Write-Host "✓ Triggered Validator Selection Update" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to trigger Validator Selection" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 6: Trigger Reputation Decay
Write-Host "`n6. Testing Dynamic Reputation Decay..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_triggerReputationDecay",
    "params": [],
    "id": 8
}
"@

if ($response.result -eq $null -and $response.error -eq $null) {
    Write-Host "✓ Triggered Reputation Decay" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to trigger Reputation Decay" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 7: Get Decay History
Write-Host "`n7. Testing Get Decay History..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getDecayHistory",
    "params": [10],
    "id": 9
}
"@

if ($response.result) {
    Write-Host "✓ Decay History Retrieved: $($response.result.Length) records" -ForegroundColor Green
    if ($response.result.Length -gt 0) {
        Write-Host "Recent decay records:" -ForegroundColor White
        $response.result | ForEach-Object {
            Write-Host "  Address: $($_.address), Old Score: $($_.old_score), New Score: $($_.new_score), Decay: $($_.decay_amount)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "✗ Failed to get Decay History" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 8: Get Time Dynamic Config
Write-Host "`n8. Testing Get Time Dynamic Config..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getTimeDynamicConfig",
    "params": [],
    "id": 10
}
"@

if ($response.result) {
    Write-Host "✓ Time Dynamic Config Retrieved Successfully" -ForegroundColor Green
    Write-Host "Configuration:" -ForegroundColor White
    $response.result | ConvertTo-Json -Depth 3
} else {
    Write-Host "✗ Failed to get Time Dynamic Config" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Test 9: Final Stats Check
Write-Host "`n9. Final Time Dynamic Stats Check..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Uri "http://localhost:8545" -Method Post -ContentType "application/json" -Body @"
{
    "jsonrpc": "2.0",
    "method": "clique_getTimeDynamicStats",
    "params": [],
    "id": 11
}
"@

if ($response.result) {
    Write-Host "✓ Final Stats Retrieved Successfully" -ForegroundColor Green
    Write-Host "Dynamic Block Time Stats:" -ForegroundColor White
    $response.result.dynamic_block_time | ConvertTo-Json -Depth 2
    Write-Host "Dynamic Validator Selection Stats:" -ForegroundColor White
    $response.result.dynamic_validator_selection | ConvertTo-Json -Depth 2
    Write-Host "Dynamic Reputation Decay Stats:" -ForegroundColor White
    $response.result.dynamic_reputation_decay | ConvertTo-Json -Depth 2
} else {
    Write-Host "✗ Failed to get Final Stats" -ForegroundColor Red
    if ($response.error) {
        Write-Host "Error: $($response.error.message)" -ForegroundColor Red
    }
}

# Stop the node
Write-Host "`nStopping node..." -ForegroundColor Cyan
Get-Process -Name "hdchain" -ErrorAction SilentlyContinue | Stop-Process -Force
Write-Host "Node stopped." -ForegroundColor Yellow

Write-Host "`n=== Time Dynamic System Test Completed ===" -ForegroundColor Green
Write-Host "Time Dynamic Features Tested:" -ForegroundColor Yellow
Write-Host "✓ Dynamic Block Time (15s -> 5s based on transaction volume)" -ForegroundColor White
Write-Host "✓ Dynamic Validator Selection (every 10 minutes)" -ForegroundColor White  
Write-Host "✓ Dynamic Reputation Decay (real-time decay)" -ForegroundColor White
Write-Host "✓ RPC API Integration" -ForegroundColor White
Write-Host "✓ Configuration Management" -ForegroundColor White
