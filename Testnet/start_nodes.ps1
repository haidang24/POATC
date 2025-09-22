# Simple script to start both nodes
Write-Host "Starting POA nodes with Random POA + Anomaly Detection + Whitelist/Blacklist..." -ForegroundColor Green

# Start Node1
Write-Host "Starting Node1..." -ForegroundColor Yellow
$node1Command = "..\hdchain.exe --datadir .\node1 --networkid 1337 --port 30306 --http --http.addr 0.0.0.0 --http.port 8547 --http.api eth,net,web3,txpool,debug,personal,admin,clique,miner --http.corsdomain * --http.vhosts * --ws --ws.addr 0.0.0.0 --ws.port 8548 --ws.api eth,net,web3,txpool,debug,personal,admin,clique,miner --ws.origins * --allow-insecure-unlock --syncmode full --mine --miner.etherbase 0x6519B747fC2c4DD4393843855Bef77f28875B07C --unlock 0x6519B747fC2c4DD4393843855Bef77f28875B07C --password .\node1\password.txt"

Start-Process powershell -ArgumentList "-NoExit -Command & { $node1Command }" | Out-Null
Write-Host "Node1 started on port 8547" -ForegroundColor Green

# Wait a bit
Start-Sleep -Seconds 3

# Start Node2
Write-Host "Starting Node2..." -ForegroundColor Yellow
$node2Command = "..\hdchain.exe --datadir .\node2 --networkid 1337 --port 30307 --http --http.addr 127.0.0.1 --http.port 8549 --http.api eth,net,web3,txpool,debug,personal,admin,clique,miner --http.corsdomain * --http.vhosts * --ws --ws.addr 127.0.0.1 --ws.port 8550 --ws.api eth,net,web3,txpool,debug,personal,admin,clique,miner --ws.origins * --allow-insecure-unlock --syncmode full --mine --miner.etherbase 0x89aEae88fE9298755eaa5B9094C5DA1e7536a505 --unlock 0x89aEae88fE9298755eaa5B9094C5DA1e7536a505 --password .\node2\password.txt"

Start-Process powershell -ArgumentList "-NoExit -Command & { $node2Command }" | Out-Null
Write-Host "Node2 started on port 8549" -ForegroundColor Green

Write-Host ""
Write-Host "Both nodes are starting..." -ForegroundColor Cyan
Write-Host "Node1: http://localhost:8547" -ForegroundColor White
Write-Host "Node2: http://localhost:8549" -ForegroundColor White
Write-Host ""
Write-Host "Wait 15 seconds for nodes to initialize, then run:" -ForegroundColor Yellow
Write-Host ".\quick_test.ps1" -ForegroundColor White
