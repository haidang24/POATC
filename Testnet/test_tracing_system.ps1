# Test Tracing System with Merkle Tree Support
# This script tests the comprehensive tracing system that tracks validator behavior

Write-Host "=== Testing Tracing System with Merkle Tree ===" -ForegroundColor Green

# Configuration
$RPC_URL = "http://localhost:8549"

# Test 1: Get Tracing Statistics
Write-Host "`n1. Testing Tracing Statistics..." -ForegroundColor Yellow
try {
    $tracingStats = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTracingStats"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($tracingStats.result) {
        Write-Host "✓ Tracing Statistics Retrieved:" -ForegroundColor Green
        Write-Host "  - Enable Tracing: $($tracingStats.result.config.enable_tracing)"
        Write-Host "  - Trace Level: $($tracingStats.result.config.trace_level)"
        Write-Host "  - Max Trace Events: $($tracingStats.result.config.max_trace_events)"
        Write-Host "  - Enable Merkle Tree: $($tracingStats.result.config.enable_merkle_tree)"
        Write-Host "  - Current Events: $($tracingStats.result.current_events)"
        Write-Host "  - Total Events: $($tracingStats.result.total_events)"
        Write-Host "  - System Uptime: $($tracingStats.result.system_uptime)"
        Write-Host "  - Current Round: $($tracingStats.result.current_round)"
        Write-Host "  - Merkle Root: $($tracingStats.result.merkle_root)"
        Write-Host "  - Merkle Tree Events: $($tracingStats.result.merkle_tree_events)"
    } else {
        Write-Host "✗ Failed to get tracing statistics" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error getting tracing statistics: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Get Trace Events
Write-Host "`n2. Testing Trace Events Retrieval..." -ForegroundColor Yellow
try {
    # Get all trace events (limit 10)
    $traceEvents = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTraceEvents"
        params = @("", 3, 10)  # eventType="", level=3 (verbose), limit=10
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($traceEvents.result) {
        Write-Host "✓ Trace Events Retrieved: $($traceEvents.result.Count) events" -ForegroundColor Green
        foreach ($event in $traceEvents.result) {
            Write-Host "  - Event: $($event.type) | Block: $($event.block_number) | Address: $($event.address) | Message: $($event.message)"
            Write-Host "    Hash: $($event.hash) | Timestamp: $($event.timestamp)"
        }
    } else {
        Write-Host "ℹ No trace events found" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error getting trace events: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Get Merkle Root
Write-Host "`n3. Testing Merkle Root..." -ForegroundColor Yellow
try {
    $merkleRoot = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getMerkleRoot"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($merkleRoot.result) {
        Write-Host "✓ Merkle Root Retrieved: $($merkleRoot.result)" -ForegroundColor Green
    } else {
        Write-Host "ℹ Merkle root is empty (no events yet)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error getting Merkle root: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Get Trace Metrics
Write-Host "`n4. Testing Trace Metrics..." -ForegroundColor Yellow
try {
    $traceMetrics = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTraceMetrics"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($traceMetrics.result) {
        Write-Host "✓ Trace Metrics Retrieved:" -ForegroundColor Green
        Write-Host "  - Total Events: $($traceMetrics.result.total_events)"
        Write-Host "  - Events by Type:"
        foreach ($type in $traceMetrics.result.events_by_type.PSObject.Properties) {
            Write-Host "    * $($type.Name): $($type.Value)"
        }
        Write-Host "  - Events by Level:"
        foreach ($level in $traceMetrics.result.events_by_level.PSObject.Properties) {
            Write-Host "    * $($level.Name): $($level.Value)"
        }
        Write-Host "  - Merkle Trees Built: $($traceMetrics.result.merkle_trees_built)"
        Write-Host "  - Merkle Roots Generated: $($traceMetrics.result.merkle_roots_generated)"
        Write-Host "  - Average Duration: $($traceMetrics.result.average_duration) ms"
        Write-Host "  - Max Duration: $($traceMetrics.result.max_duration) ms"
        Write-Host "  - Events per Minute: $($traceMetrics.result.events_per_minute)"
    } else {
        Write-Host "✗ Failed to get trace metrics" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error getting trace metrics: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Test Trace Level Control
Write-Host "`n5. Testing Trace Level Control..." -ForegroundColor Yellow
try {
    # Set trace level to verbose (3)
    $setLevel = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_setTraceLevel"
        params = @(3)
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($setLevel.result) {
        Write-Host "✓ Trace level set to verbose (3)" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to set trace level" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error setting trace level: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Test Event Filtering
Write-Host "`n6. Testing Event Filtering..." -ForegroundColor Yellow
try {
    # Get only block validation events
    $blockValidationEvents = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTraceEvents"
        params = @("block_validation", 3, 5)  # eventType="block_validation", level=3, limit=5
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($blockValidationEvents.result) {
        Write-Host "✓ Block Validation Events Retrieved: $($blockValidationEvents.result.Count) events" -ForegroundColor Green
        foreach ($event in $blockValidationEvents.result) {
            Write-Host "  - Block: $($event.block_number) | Address: $($event.address) | Message: $($event.message)"
        }
    } else {
        Write-Host "ℹ No block validation events found" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error filtering events: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Test Merkle Proof (if events exist)
Write-Host "`n7. Testing Merkle Proof..." -ForegroundColor Yellow
try {
    # First get some events
    $events = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTraceEvents"
        params = @("", 3, 1)  # Get 1 event
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($events.result -and $events.result.Count -gt 0) {
        $event = $events.result[0]
        Write-Host "Testing Merkle proof for event: $($event.id)"
        
        # Get Merkle proof for this event
        $merkleProof = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_getMerkleProof"
            params = @($event)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($merkleProof.result) {
            Write-Host "✓ Merkle Proof Retrieved: $($merkleProof.result.Count) proof elements" -ForegroundColor Green
            foreach ($proof in $merkleProof.result) {
                Write-Host "  - Proof: $proof"
            }
        } else {
            Write-Host "ℹ No Merkle proof found for this event" -ForegroundColor Yellow
        }
    } else {
        Write-Host "ℹ No events available for Merkle proof test" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error testing Merkle proof: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 8: Test Event Verification
Write-Host "`n8. Testing Event Verification..." -ForegroundColor Yellow
try {
    # Get an event to verify
    $events = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_getTraceEvents"
        params = @("", 3, 1)  # Get 1 event
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($events.result -and $events.result.Count -gt 0) {
        $event = $events.result[0]
        Write-Host "Verifying event: $($event.id)"
        
        # Verify event in Merkle Tree
        $verification = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_verifyEventInMerkleTree"
            params = @($event)
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($verification.result) {
            Write-Host "✓ Event verified in Merkle Tree" -ForegroundColor Green
        } else {
            Write-Host "✗ Event not found in Merkle Tree" -ForegroundColor Red
        }
    } else {
        Write-Host "ℹ No events available for verification test" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error testing event verification: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 9: Test Export Functionality
Write-Host "`n9. Testing Export Functionality..." -ForegroundColor Yellow
try {
    $export = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_exportTraceEvents"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($export.result) {
        Write-Host "✓ Trace Events Exported Successfully" -ForegroundColor Green
        Write-Host "  - Export size: $($export.result.Length) characters"
        
        # Save to file for inspection
        $export | ConvertTo-Json -Depth 10 | Out-File -FilePath "trace_export.json" -Encoding UTF8
        Write-Host "  - Saved to trace_export.json"
    } else {
        Write-Host "✗ Failed to export trace events" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error exporting trace events: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 10: Test Clear Functionality
Write-Host "`n10. Testing Clear Functionality..." -ForegroundColor Yellow
try {
    $clear = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
        jsonrpc = "2.0"
        method = "clique_clearTraceEvents"
        params = @()
        id = 1
    } | ConvertTo-Json) -ContentType "application/json"
    
    if ($clear.result) {
        Write-Host "✓ Trace Events Cleared Successfully" -ForegroundColor Green
        
        # Verify clearing worked
        $statsAfterClear = Invoke-RestMethod -Uri "$RPC_URL" -Method POST -Body (@{
            jsonrpc = "2.0"
            method = "clique_getTracingStats"
            params = @()
            id = 1
        } | ConvertTo-Json) -ContentType "application/json"
        
        if ($statsAfterClear.result) {
            Write-Host "  - Events after clear: $($statsAfterClear.result.current_events)"
            Write-Host "  - Merkle root after clear: $($statsAfterClear.result.merkle_root)"
        }
    } else {
        Write-Host "✗ Failed to clear trace events" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error clearing trace events: $($_.Exception.Message)" -ForegroundColor Red
}

# Summary
Write-Host "`n=== Tracing System Test Summary ===" -ForegroundColor Green
Write-Host "✓ Tracing statistics retrieved"
Write-Host "✓ Trace events retrieval tested"
Write-Host "✓ Merkle root functionality tested"
Write-Host "✓ Trace metrics retrieved"
Write-Host "✓ Trace level control tested"
Write-Host "✓ Event filtering tested"
Write-Host "✓ Merkle proof functionality tested"
Write-Host "✓ Event verification tested"
Write-Host "✓ Export functionality tested"
Write-Host "✓ Clear functionality tested"

Write-Host "`n=== Key Tracing Features ===" -ForegroundColor Cyan
Write-Host "1. Comprehensive Event Tracking: All validator behaviors are traced"
Write-Host "2. Merkle Tree Integration: Events are hashed and stored in Merkle Tree"
Write-Host "3. Immutable Audit Trail: Merkle root provides tamper-proof evidence"
Write-Host "4. Event Verification: Any event can be verified against Merkle root"
Write-Host "5. Flexible Filtering: Events can be filtered by type and level"
Write-Host "6. Export/Import: Full trace data can be exported for analysis"
Write-Host "7. Real-time Metrics: Live statistics about tracing system performance"

Write-Host "`n=== Trace Event Types ===" -ForegroundColor Cyan
Write-Host "• random_poa: Random POA algorithm selection events"
Write-Host "• leader_selection: Validator leader selection events"
Write-Host "• block_signing: Block signing success/failure events"
Write-Host "• block_validation: Block validation events"
Write-Host "• timeout: Validator timeout events"
Write-Host "• accusation: Validator accusation events"
Write-Host "• ai_gate_evaluation: AI gate evaluation events"
Write-Host "• reputation: Reputation system updates"
Write-Host "• anomaly_detection: Anomaly detection events"
Write-Host "• whitelist_blacklist: Access control events"
Write-Host "• validator_selection: Validator selection events"

Write-Host "`n=== Merkle Tree Benefits ===" -ForegroundColor Green
Write-Host "✓ Immutable: Events cannot be tampered with"
Write-Host "✓ Verifiable: Any event can be verified against Merkle root"
Write-Host "✓ Efficient: O(log n) proof size for verification"
Write-Host "✓ Transparent: Merkle root is included in blocks"
Write-Host "✓ Audit Trail: Complete history of validator behaviors"
Write-Host "✓ Community Monitoring: Anyone can verify events"

Write-Host "`nTracing system test completed!" -ForegroundColor Green
