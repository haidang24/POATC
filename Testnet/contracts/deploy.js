// Deploy DataTraceability contract to POATC blockchain
const { Web3 } = require('web3');
const fs = require('fs');
const path = require('path');

// Configuration
const RPC_URL = process.env.RPC_URL || 'http://127.0.0.1:8545';
const DEPLOYER_ADDRESS = process.env.DEPLOYER_ADDRESS || '0x3003d6498603fAD5F232452B21c8B6EB798d20f1';

async function deploy() {
  console.log('╔══════════════════════════════════════════════════════════╗');
  console.log('║   DEPLOYING DataTraceability CONTRACT TO POATC          ║');
  console.log('╚══════════════════════════════════════════════════════════╝\n');

  // Connect to blockchain
  const web3 = new Web3(RPC_URL);
  
  console.log(`📡 Connecting to: ${RPC_URL}`);
  
  try {
    const blockNumber = await web3.eth.getBlockNumber();
    console.log(`✓ Connected! Current block: ${blockNumber}\n`);
  } catch (error) {
    console.error('❌ Failed to connect to blockchain:', error.message);
    process.exit(1);
  }

  // Load compiled contract
  const contractPath = path.join(__dirname, 'build', 'DataTraceability.json');
  
  if (!fs.existsSync(contractPath)) {
    console.error('❌ Contract not compiled! Run: node compile.js');
    process.exit(1);
  }

  const contractData = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
  const { abi, bytecode } = contractData;

  console.log('📄 Contract loaded:');
  console.log(`   - Name: DataTraceability`);
  console.log(`   - Bytecode size: ${bytecode.length / 2} bytes`);
  console.log(`   - Functions: ${abi.filter(i => i.type === 'function').length}`);
  console.log(`   - Events: ${abi.filter(i => i.type === 'event').length}\n`);

  // Create contract instance
  const contract = new web3.eth.Contract(abi);

  console.log(`👤 Deployer address: ${DEPLOYER_ADDRESS}\n`);
  console.log('🚀 Deploying contract...\n');

  try {
    // Deploy
    const deployTx = contract.deploy({
      data: bytecode,
      arguments: [] // Constructor has no arguments
    });

    const gas = await deployTx.estimateGas({ from: DEPLOYER_ADDRESS });
    console.log(`⛽ Estimated gas: ${gas.toLocaleString()}\n`);

    const deployedContract = await deployTx.send({
      from: DEPLOYER_ADDRESS,
      gas: Math.floor(gas * 1.2), // Add 20% buffer
      gasPrice: await web3.eth.getGasPrice()
    });

    const contractAddress = deployedContract.options.address;

    console.log('╔══════════════════════════════════════════════════════════╗');
    console.log('║              DEPLOYMENT SUCCESSFUL! ✅                   ║');
    console.log('╚══════════════════════════════════════════════════════════╝\n');
    console.log(`📍 Contract Address: ${contractAddress}\n`);

    // Save deployment info
    const deploymentInfo = {
      contractName: 'DataTraceability',
      address: contractAddress,
      deployer: DEPLOYER_ADDRESS,
      deployedAt: new Date().toISOString(),
      blockNumber: await web3.eth.getBlockNumber(),
      network: 'POATC Testnet',
      rpcUrl: RPC_URL,
      abi: abi
    };

    const deploymentPath = path.join(__dirname, 'build', 'deployment.json');
    fs.writeFileSync(deploymentPath, JSON.stringify(deploymentInfo, null, 2));

    console.log('💾 Deployment info saved to:', deploymentPath);
    console.log('\n📋 Quick copy:');
    console.log(`   Contract Address: ${contractAddress}`);
    console.log(`   Explorer: http://localhost:8080/#/address/${contractAddress}`);

    return contractAddress;

  } catch (error) {
    console.error('\n❌ Deployment failed:', error.message);
    if (error.message.includes('unlock')) {
      console.error('\n💡 Hint: Make sure the account is unlocked!');
      console.error('   The node should be started with --unlock flag');
    }
    process.exit(1);
  }
}

// Run deployment
if (require.main === module) {
  deploy().then(address => {
    console.log('\n✅ Deployment complete!');
    process.exit(0);
  }).catch(error => {
    console.error('\n❌ Error:', error);
    process.exit(1);
  });
}

module.exports = { deploy };

