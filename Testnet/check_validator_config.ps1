# Script to check current validator selection configuration
Write-Host "=== Validator Selection Configuration Check ===" -ForegroundColor Magenta
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

# Test URLs
$node1Url = "http://localhost:8547"

# Check if node is running
Write-Host "1. Checking Node Connectivity..." -ForegroundColor Yellow
$node1Block = Invoke-RPC -Url $node1Url -Method "eth_blockNumber"

if ($node1Block -and $node1Block.result) {
    $block1 = [Convert]::ToInt32($node1Block.result, 16)
    Write-Host "✅ Node1 Block: $block1" -ForegroundColor Green
} else {
    Write-Host "❌ Node not responding. Please start node first." -ForegroundColor Red
    Write-Host "Run: .\start_nodes.ps1" -ForegroundColor Yellow
    exit 1
}

# Get Validator Selection Stats
Write-Host "`n2. Current Validator Selection Configuration:" -ForegroundColor Yellow

$validatorStats = Invoke-RPC -Url $node1Url -Method "clique_getValidatorSelectionStats"
if ($validatorStats -and -not $validatorStats.error) {
    $stats = $validatorStats.result
    $config = $stats.config
    
    Write-Host "✅ Configuration Details:" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 VALIDATOR SELECTION SETTINGS:" -ForegroundColor Cyan
    Write-Host "  • Enable Validator Selection: $($config.enable_validator_selection)" -ForegroundColor White
    Write-Host "  • Small Validator Set Size: $($config.small_validator_set_size) validators" -ForegroundColor White
    Write-Host "  • Selection Method: $($config.selection_method)" -ForegroundColor White
    Write-Host "  • Selection Window: $($config.selection_window)" -ForegroundColor White
    
    Write-Host ""
    Write-Host "📊 CURRENT STATUS:" -ForegroundColor Cyan
    Write-Host "  • Total Validators: $($stats.validators.total)" -ForegroundColor White
    Write-Host "  • Active Validators: $($stats.validators.active)" -ForegroundColor White
    Write-Host "  • Current Small Set Size: $($stats.validators.small_set_size)" -ForegroundColor White
    
    Write-Host ""
    Write-Host "🎯 SELECTION WEIGHTS (Hybrid Method):" -ForegroundColor Cyan
    if ($config.selection_method -eq "hybrid") {
        Write-Host "  • Stake Weight: $($config.stake_weight * 100)%" -ForegroundColor White
        Write-Host "  • Reputation Weight: $($config.reputation_weight * 100)%" -ForegroundColor White
        Write-Host "  • Random Weight: $($config.random_weight * 100)%" -ForegroundColor White
    }
    
    Write-Host ""
    Write-Host "📈 TOTALS:" -ForegroundColor Cyan
    Write-Host "  • Total Stake: $($stats.totals.stake)" -ForegroundColor White
    Write-Host "  • Total Reputation: $($stats.totals.reputation)" -ForegroundColor White
    Write-Host "  • Total Blocks Mined: $($stats.totals.blocks)" -ForegroundColor White
    
    Write-Host ""
    Write-Host "🔄 SELECTION HISTORY:" -ForegroundColor Cyan
    Write-Host "  • History Records: $($stats.selection.history_count)" -ForegroundColor White
    Write-Host "  • Last Selection: $($stats.selection.last_selection)" -ForegroundColor White
    
    # Get current small validator set
    Write-Host ""
    Write-Host "👥 CURRENT SMALL VALIDATOR SET:" -ForegroundColor Cyan
    $currentSet = $stats.selection.current_set
    if ($currentSet -and $currentSet.Count -gt 0) {
        for ($i = 0; $i -lt $currentSet.Count; $i++) {
            Write-Host "  $($i+1). $($currentSet[$i])" -ForegroundColor White
        }
    } else {
        Write-Host "  No small validator set available" -ForegroundColor Yellow
    }
    
} else {
    Write-Host "❌ Failed to get validator selection stats" -ForegroundColor Red
    if ($validatorStats.error) {
        Write-Host "  Error: $($validatorStats.error.message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== SUMMARY ===" -ForegroundColor Magenta
Write-Host "🎯 Tầng 1: Chọn 3 validators từ tất cả validators" -ForegroundColor Green
Write-Host "🎲 Tầng 2: Random chọn 1 validator từ 3 validators đó" -ForegroundColor Green
Write-Host "⏰ Thời gian giữ nguyên tập validator: 1 giờ" -ForegroundColor Green
Write-Host "🔄 Phương pháp chọn: hybrid (stake + reputation + random)" -ForegroundColor Green

Write-Host ""
Write-Host "📝 Trả lời câu hỏi:" -ForegroundColor Cyan
Write-Host "  • Số validator chọn ở tầng 1: 3 validators" -ForegroundColor White
Write-Host "  • 1 vòng kéo dài: 1 giờ (không phải theo block)" -ForegroundColor White
Write-Host "  • Mỗi block sẽ random chọn 1 validator từ 3 validators" -ForegroundColor White