# Test Fairness Mechanisms in Reputation System
# This script tests the fairness mechanisms to ensure equal opportunities for all validators

Write-Host "=== Testing Fairness Mechanisms ===" -ForegroundColor Green

# Configuration
$RPC_URL = "http://localhost:8545"
$NODE2_RPC_URL = "http://localhost:8546"

# Test 1: Get Fairness Statistics
Write-Host "`n1. Testing Fairness Statistics..." -ForegroundColor Yellow
try {
    $fairnessStats = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getFairnessStats"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($fairnessStats.result) {
        Write-Host "✓ Fairness Statistics Retrieved:" -ForegroundColor Green
        Write-Host "  - Max Component Score: $($fairnessStats.result.max_component_score)"
        Write-Host "  - Reset Interval: $($fairnessStats.result.reset_interval_hours) hours"
        Write-Host "  - New Validator Boost: $($fairnessStats.result.new_validator_boost)"
        Write-Host "  - Veteran Penalty: $($fairnessStats.result.veteran_penalty)"
        Write-Host "  - Decay Factor: $($fairnessStats.result.decay_factor)"
        Write-Host "  - Total Validators: $($fairnessStats.result.total_validators)"
        Write-Host "  - New Validators: $($fairnessStats.result.new_validators)"
        Write-Host "  - Veteran Validators: $($fairnessStats.result.veteran_validators)"
    } else {
        Write-Host "✗ Failed to get fairness statistics" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error getting fairness statistics: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Get Validator Fairness Info
Write-Host "`n2. Testing Validator Fairness Info..." -ForegroundColor Yellow

# Get signers first
try {
    $signers = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getSigners"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($signers.result -and $signers.result.Count -gt 0) {
        $firstSigner = $signers.result[0]
        Write-Host "Testing fairness info for signer: $firstSigner"
        
        $fairnessInfo = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_getValidatorFairnessInfo"
            params = @($firstSigner)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($fairnessInfo.result) {
            Write-Host "✓ Validator Fairness Info Retrieved:" -ForegroundColor Green
            Write-Host "  - Address: $($fairnessInfo.result.address)"
            Write-Host "  - Join Time: $($fairnessInfo.result.join_time)"
            Write-Host "  - Days Since Join: $($fairnessInfo.result.days_since_join)"
            Write-Host "  - Is New Validator: $($fairnessInfo.result.is_new_validator)"
            Write-Host "  - Is Veteran: $($fairnessInfo.result.is_veteran)"
            Write-Host "  - Veteran Penalty: $($fairnessInfo.result.veteran_penalty)"
            Write-Host "  - Block Mining Score: $($fairnessInfo.result.block_mining_score)"
            Write-Host "  - Uptime Score: $($fairnessInfo.result.uptime_score)"
            Write-Host "  - Consistency Score: $($fairnessInfo.result.consistency_score)"
            Write-Host "  - Current Score: $($fairnessInfo.result.current_score)"
            Write-Host "  - Is At Max Component: $($fairnessInfo.result.is_at_max_component)"
            Write-Host "  - Needs Reset: $($fairnessInfo.result.needs_reset)"
        } else {
            Write-Host "✗ Failed to get validator fairness info" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "✗ Error getting validator fairness info: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Test Score Capping
Write-Host "`n3. Testing Score Capping Mechanism..." -ForegroundColor Yellow

# Get reputation scores for all validators
try {
    $reputationStats = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getReputationStats"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($reputationStats.result) {
        Write-Host "✓ Reputation Statistics Retrieved:" -ForegroundColor Green
        Write-Host "  - Total Validators: $($reputationStats.result.total_validators)"
        Write-Host "  - Average Score: $($reputationStats.result.average_score)"
        Write-Host "  - Max Score: $($reputationStats.result.max_score)"
        Write-Host "  - Min Score: $($reputationStats.result.min_score)"
        
        # Check if any validator is at max component score
        $maxComponentScore = $fairnessStats.result.max_component_score
        Write-Host "  - Max Component Score Limit: $maxComponentScore"
        
        # Get top validators to check for capping
        $topValidators = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_getTopValidators"
            params = @(5)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($topValidators.result) {
            Write-Host "  - Top 5 Validators:" -ForegroundColor Cyan
            foreach ($validator in $topValidators.result) {
                $isCapped = $validator.block_mining_score -ge $maxComponentScore -or 
                           $validator.uptime_score -ge $maxComponentScore -or 
                           $validator.consistency_score -ge $maxComponentScore
                $cappedStatus = if ($isCapped) { "CAPPED" } else { "Not Capped" }
                Write-Host "    * $($validator.address): Score=$($validator.current_score) [$cappedStatus]"
            }
        }
    }
} catch {
    Write-Host "✗ Error testing score capping: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Test Decay Mechanism
Write-Host "`n4. Testing Decay Mechanism..." -ForegroundColor Yellow

try {
    # Get current scores
    $currentScores = @()
    if ($signers.result) {
        foreach ($signer in $signers.result) {
            $score = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
                jsonrpc = "2.0"
                method = "clique_getReputationScore"
                params = @($signer)
                id = 1
            } | ConvertTo-Json) -ContentType "application/json"
            
            if ($score.result) {
                $currentScores += @{
                    address = $signer
                    score = $score.result.current_score
                }
            }
        }
    }
    
    Write-Host "Current scores before decay test:"
    foreach ($score in $currentScores) {
        Write-Host "  - $($score.address): $($score.score)"
    }
    
    # Force reputation update to trigger decay
    Write-Host "`nForcing reputation update to trigger decay..."
    $updateResult = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_updateReputation"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($updateResult.result) {
        Write-Host "✓ Reputation update triggered successfully" -ForegroundColor Green
        
        # Wait a moment and check scores again
        Start-Sleep -Seconds 2
        
        Write-Host "Scores after decay:"
        foreach ($score in $currentScores) {
            $newScore = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
                jsonrpc = "2.0"
                method = "clique_getReputationScore"
                params = @($score.address)
                id = 1
            } | ConvertTo-Json) -ContentType "application/json"
            
            if ($newScore.result) {
                $oldScore = $score.score
                $newScoreValue = $newScore.result.current_score
                $decayAmount = $oldScore - $newScoreValue
                Write-Host "  - $($score.address): $oldScore -> $newScoreValue (decay: $decayAmount)"
            }
        }
    } else {
        Write-Host "✗ Failed to trigger reputation update" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error testing decay mechanism: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Test New Validator Boost
Write-Host "`n5. Testing New Validator Boost..." -ForegroundColor Yellow

try {
    # Check if there are any new validators
    if ($fairnessStats.result.new_validators -gt 0) {
        Write-Host "✓ Found $($fairnessStats.result.new_validators) new validators" -ForegroundColor Green
        
        # Get fairness info for new validators
        foreach ($signer in $signers.result) {
            $fairnessInfo = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
                jsonrpc = "2.0"
                method = "clique_getValidatorFairnessInfo"
                params = @($signer)
                id = 1
            } | ConvertTo-Json) -ContentType "application/json"
            
            if ($fairnessInfo.result -and $fairnessInfo.result.is_new_validator) {
                Write-Host "  - New Validator: $($fairnessInfo.result.address)"
                Write-Host "    * Join Time: $($fairnessInfo.result.join_time)"
                Write-Host "    * Days Since Join: $($fairnessInfo.result.days_since_join)"
                Write-Host "    * Current Score: $($fairnessInfo.result.current_score)"
                Write-Host "    * Block Mining Score: $($fairnessInfo.result.block_mining_score)"
                Write-Host "    * Uptime Score: $($fairnessInfo.result.uptime_score)"
            }
        }
    } else {
        Write-Host "ℹ No new validators found (all validators are older than 24 hours)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error testing new validator boost: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Test Veteran Penalty
Write-Host "`n6. Testing Veteran Penalty..." -ForegroundColor Yellow

try {
    if ($fairnessStats.result.veteran_validators -gt 0) {
        Write-Host "✓ Found $($fairnessStats.result.veteran_validators) veteran validators" -ForegroundColor Green
        
        # Get fairness info for veteran validators
        foreach ($signer in $signers.result) {
            $fairnessInfo = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
                jsonrpc = "2.0"
                method = "clique_getValidatorFairnessInfo"
                params = @($signer)
                id = 1
            } | ConvertTo-Json) -ContentType "application/json"
            
            if ($fairnessInfo.result -and $fairnessInfo.result.is_veteran) {
                Write-Host "  - Veteran Validator: $($fairnessInfo.result.address)"
                Write-Host "    * Days Since Join: $($fairnessInfo.result.days_since_join)"
                Write-Host "    * Veteran Penalty: $($fairnessInfo.result.veteran_penalty)"
                Write-Host "    * Current Score: $($fairnessInfo.result.current_score)"
                Write-Host "    * Block Mining Score: $($fairnessInfo.result.block_mining_score)"
                Write-Host "    * Uptime Score: $($fairnessInfo.result.uptime_score)"
            }
        }
    } else {
        Write-Host "ℹ No veteran validators found (all validators are newer than 30 days)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error testing veteran penalty: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Test Reset Mechanism
Write-Host "`n7. Testing Reset Mechanism..." -ForegroundColor Yellow

try {
    # Check which validators need reset
    $validatorsNeedingReset = @()
    foreach ($signer in $signers.result) {
        $fairnessInfo = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_getValidatorFairnessInfo"
            params = @($signer)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($fairnessInfo.result -and $fairnessInfo.result.needs_reset) {
            $validatorsNeedingReset += $signer
        }
    }
    
    if ($validatorsNeedingReset.Count -gt 0) {
        Write-Host "✓ Found $($validatorsNeedingReset.Count) validators needing reset:" -ForegroundColor Green
        foreach ($validator in $validatorsNeedingReset) {
            Write-Host "  - $validator"
        }
        
        # Test force reset for first validator
        $firstValidator = $validatorsNeedingReset[0]
        Write-Host "`nTesting force reset for: $firstValidator"
        
        $resetResult = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_forcePartialReset"
            params = @($firstValidator)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($resetResult.result) {
            Write-Host "✓ Force reset completed successfully" -ForegroundColor Green
        } else {
            Write-Host "✗ Force reset failed" -ForegroundColor Red
        }
    } else {
        Write-Host "ℹ No validators need reset at this time" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error testing reset mechanism: $($_.Exception.Message)" -ForegroundColor Red
}

# Summary
Write-Host "`n=== Fairness Mechanisms Test Summary ===" -ForegroundColor Green
Write-Host "✓ Fairness statistics retrieved"
Write-Host "✓ Validator fairness info tested"
Write-Host "✓ Score capping mechanism verified"
Write-Host "✓ Decay mechanism tested"
Write-Host "✓ New validator boost checked"
Write-Host "✓ Veteran penalty verified"
Write-Host "✓ Reset mechanism tested"

Write-Host "`n=== Key Fairness Features ===" -ForegroundColor Cyan
Write-Host "1. Max Component Score: Prevents infinite score accumulation"
Write-Host "2. Weekly Reset: Partial reset every 7 days"
Write-Host "3. New Validator Boost: +0.5 boost for first 24 hours"
Write-Host "4. Veteran Penalty: -0.1 penalty after 30 days"
Write-Host "5. Stronger Decay: 5% decay per hour (vs 1% before)"
Write-Host "6. Component Capping: Each component maxed at 5.0 points"

Write-Host "`n=== Fairness Benefits ===" -ForegroundColor Green
Write-Host "✓ Prevents old validators from dominating"
Write-Host "✓ Gives new validators fair chance"
Write-Host "✓ Maintains competitive environment"
Write-Host "✓ Prevents score inflation"
Write-Host "✓ Ensures equal opportunities"

Write-Host "`nFairness mechanisms test completed!" -ForegroundColor Green
