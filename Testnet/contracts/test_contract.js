// Test DataTraceability contract
const { Web3 } = require('web3');
const fs = require('fs');
const path = require('path');

const RPC_URL = process.env.RPC_URL || 'http://127.0.0.1:8545';

async function testContract() {
  console.log('╔══════════════════════════════════════════════════════════╗');
  console.log('║         TESTING DataTraceability CONTRACT               ║');
  console.log('╚══════════════════════════════════════════════════════════╝\n');

  const web3 = new Web3(RPC_URL);

  // Load deployment
  const deploymentPath = path.join(__dirname, 'build', 'deployment.json');
  if (!fs.existsSync(deploymentPath)) {
    console.error('❌ Contract not deployed!');
    process.exit(1);
  }

  const deployment = JSON.parse(fs.readFileSync(deploymentPath, 'utf8'));
  const contract = new web3.eth.Contract(deployment.abi, deployment.address);

  console.log(`📍 Testing contract at: ${deployment.address}\n`);

  try {
    // Test 1: Get contract info
    console.log('Test 1: Get Contract Info');
    const info = await contract.methods.getContractInfo().call();
    console.log(`   ✓ Owner: ${info.contractOwner}`);
    console.log(`   ✓ Total Records: ${info.totalRecords}`);
    console.log(`   ✓ Block Number: ${info.blockNumber}`);
    console.log(`   ✓ Timestamp: ${new Date(Number(info.blockTimestamp) * 1000).toISOString()}\n`);

    if (Number(info.totalRecords) > 0) {
      // Test 2: Get first record
      console.log('Test 2: Get Record Details');
      const record = await contract.methods.getRecord(1).call();
      console.log(`   ✓ Record ID: ${record.id}`);
      console.log(`   ✓ Type: ${record.dataType}`);
      console.log(`   ✓ Description: ${record.description}`);
      console.log(`   ✓ Creator: ${record.creator}`);
      console.log(`   ✓ Verified: ${record.verified}`);
      console.log(`   ✓ Hash: ${record.dataHash.substring(0, 20)}...`);
      
      if (record.metadata) {
        try {
          const metadata = JSON.parse(record.metadata);
          console.log(`   ✓ Metadata keys: ${Object.keys(metadata).join(', ')}`);
        } catch(e) {
          console.log(`   ✓ Metadata: ${record.metadata.substring(0, 50)}...`);
        }
      }
      console.log('');

      // Test 3: Get trace history
      console.log('Test 3: Get Trace History');
      const history = await contract.methods.getTraceHistory(1).call();
      console.log(`   ✓ Total steps: ${history.length}`);
      history.forEach((step, idx) => {
        console.log(`   ${idx + 1}. Action: "${step.action}" by ${step.actor.substring(0, 10)}...`);
        console.log(`      Details: ${step.details}`);
        console.log(`      Time: ${new Date(Number(step.timestamp) * 1000).toISOString()}`);
      });
      console.log('');

      // Test 4: Get user records
      console.log('Test 4: Get User Records');
      const userRecords = await contract.methods.getUserRecords(record.creator).call();
      console.log(`   ✓ User ${record.creator.substring(0, 10)}... has ${userRecords.length} record(s)`);
      console.log(`   ✓ Record IDs: ${userRecords.join(', ')}\n`);
    }

    // Test 5: Read events
    console.log('Test 5: Read Recent Events');
    const latestBlock = await web3.eth.getBlockNumber();
    const events = await contract.getPastEvents('allEvents', {
      fromBlock: Math.max(0, Number(latestBlock) - 100),
      toBlock: 'latest'
    });
    console.log(`   ✓ Found ${events.length} events in last 100 blocks`);
    
    const eventCounts = {};
    events.forEach(event => {
      eventCounts[event.event] = (eventCounts[event.event] || 0) + 1;
    });
    Object.entries(eventCounts).forEach(([name, count]) => {
      console.log(`   - ${name}: ${count}`);
    });
    console.log('');

    console.log('╔══════════════════════════════════════════════════════════╗');
    console.log('║              ALL TESTS PASSED! ✅                        ║');
    console.log('╚══════════════════════════════════════════════════════════╝\n');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  testContract().then(() => {
    process.exit(0);
  }).catch(error => {
    console.error('Error:', error);
    process.exit(1);
  });
}

module.exports = { testContract };

