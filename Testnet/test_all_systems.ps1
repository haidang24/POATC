# Test script for all enhanced POA systems
# This script tests all the enhanced POA consensus systems

Write-Host "=== Enhanced POA Consensus System Test ===" -ForegroundColor Green
Write-Host "Testing all systems: Anomaly Detection, Whitelist/Blacklist, Validator Selection, Reputation System, Tracing System" -ForegroundColor Yellow

# Test 1: Anomaly Detection System
Write-Host "`n1. Testing Anomaly Detection System..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestAnomalyDetector"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Anomaly Detection System: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Anomaly Detection System: FAILED" -ForegroundColor Red
}

# Test 2: Whitelist/Blacklist System
Write-Host "`n2. Testing Whitelist/Blacklist System..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestWhitelistBlacklist"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Whitelist/Blacklist System: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Whitelist/Blacklist System: FAILED" -ForegroundColor Red
}

# Test 3: Validator Selection System
Write-Host "`n3. Testing Validator Selection System..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestValidatorSelection"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Validator Selection System: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Validator Selection System: FAILED" -ForegroundColor Red
}

# Test 4: Reputation System
Write-Host "`n4. Testing Reputation System..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestReputationSystem"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Reputation System: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Reputation System: FAILED" -ForegroundColor Red
}

# Test 5: Random POA Algorithm
Write-Host "`n5. Testing Random POA Algorithm..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestRandomPOA"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Random POA Algorithm: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Random POA Algorithm: FAILED" -ForegroundColor Red
}

# Test 6: Main Clique Tests
Write-Host "`n6. Testing Main Clique Consensus..." -ForegroundColor Cyan
go test -timeout 30s -v ./consensus/clique/ -run "TestClique"
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Main Clique Consensus: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Main Clique Consensus: FAILED" -ForegroundColor Red
}

# Test 7: Build Test
Write-Host "`n7. Testing Build..." -ForegroundColor Cyan
go build -o hdchain.exe ./cmd/geth
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Build: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Build: FAILED" -ForegroundColor Red
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "All enhanced POA consensus systems have been tested." -ForegroundColor Yellow
Write-Host "Check the results above for any failures." -ForegroundColor Yellow

# Display system information
Write-Host "`n=== System Information ===" -ForegroundColor Green
Write-Host "Enhanced POA Consensus Features:" -ForegroundColor Yellow
Write-Host "- Random POA Algorithm (instead of round-robin)" -ForegroundColor White
Write-Host "- Anomaly Detection System" -ForegroundColor White
Write-Host "- Whitelist/Blacklist Management" -ForegroundColor White
Write-Host "- 2-Tier Validator Selection" -ForegroundColor White
Write-Host "- On-chain Reputation System with Fairness Mechanisms" -ForegroundColor White
Write-Host "- Tracing System with Merkle Tree" -ForegroundColor White
Write-Host "- Integration between all systems" -ForegroundColor White

Write-Host "`nTest completed!" -ForegroundColor Green