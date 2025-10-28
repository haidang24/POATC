// Upload sample data to DataTraceability contract
const { Web3 } = require('web3');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Configuration
const RPC_URL = process.env.RPC_URL || 'http://127.0.0.1:8545';
const USER_ADDRESS = process.env.USER_ADDRESS || '0x3003d6498603fAD5F232452B21c8B6EB798d20f1';

// Sample data templates
const sampleData = [
  {
    dataType: 'Product',
    description: 'Organic Coffee Beans - Ethiopia Grade A',
    metadata: JSON.stringify({
      origin: 'Ethiopia, Yirgacheffe',
      harvestDate: '2024-01-15',
      quantity: '1000kg',
      quality: 'Grade A',
      certifications: ['Organic', 'Fair Trade'],
      batchNumber: 'ETH-2024-001'
    })
  },
  {
    dataType: 'Document',
    description: 'Supply Chain Certificate - Batch #SC2024-456',
    metadata: JSON.stringify({
      documentType: 'Certificate',
      issuer: 'International Supply Chain Authority',
      issuedDate: '2024-02-01',
      validUntil: '2025-02-01',
      certificateNumber: 'SC2024-456'
    })
  },
  {
    dataType: 'Transaction',
    description: 'Shipment from Warehouse A to Distribution Center B',
    metadata: JSON.stringify({
      shipmentId: 'SHIP-2024-789',
      origin: 'Warehouse A, Location 1',
      destination: 'Distribution Center B, Location 2',
      departureDate: '2024-02-15',
      arrivalDate: '2024-02-18',
      carrier: 'Express Logistics Inc.',
      trackingNumber: 'EXP123456789'
    })
  },
  {
    dataType: 'Product',
    description: 'Medical Equipment - Surgical Masks FFP2',
    metadata: JSON.stringify({
      manufacturer: 'MedSupply Corp',
      manufactureDate: '2024-01-20',
      quantity: '10000 units',
      standard: 'EN 149:2001+A1:2009',
      batchNumber: 'FFP2-2024-123',
      sterile: true
    })
  },
  {
    dataType: 'IoT Sensor Data',
    description: 'Temperature monitoring during cold chain transport',
    metadata: JSON.stringify({
      sensorId: 'TEMP-SENSOR-001',
      recordingPeriod: '2024-02-15 to 2024-02-18',
      avgTemperature: '4.2°C',
      minTemperature: '2.8°C',
      maxTemperature: '5.9°C',
      alertsTriggered: 0,
      dataPoints: 864 // 3 days, every 5 minutes
    })
  }
];

async function uploadSampleData() {
  console.log('╔══════════════════════════════════════════════════════════╗');
  console.log('║      UPLOADING SAMPLE DATA TO CONTRACT                  ║');
  console.log('╚══════════════════════════════════════════════════════════╝\n');

  // Connect to blockchain
  const web3 = new Web3(RPC_URL);
  console.log(`📡 Connecting to: ${RPC_URL}`);

  try {
    const blockNumber = await web3.eth.getBlockNumber();
    console.log(`✓ Connected! Block: ${blockNumber}\n`);
  } catch (error) {
    console.error('❌ Connection failed:', error.message);
    process.exit(1);
  }

  // Load deployment info
  const deploymentPath = path.join(__dirname, 'build', 'deployment.json');
  
  if (!fs.existsSync(deploymentPath)) {
    console.error('❌ Contract not deployed! Run: node deploy.js');
    process.exit(1);
  }

  const deployment = JSON.parse(fs.readFileSync(deploymentPath, 'utf8'));
  const { address, abi } = deployment;

  console.log(`📍 Contract: ${address}`);
  console.log(`👤 User: ${USER_ADDRESS}\n`);

  // Create contract instance
  const contract = new web3.eth.Contract(abi, address);

  // Check contract info
  try {
    const info = await contract.methods.getContractInfo().call();
    console.log('📊 Contract Info:');
    console.log(`   Owner: ${info.contractOwner}`);
    console.log(`   Total Records: ${info.totalRecords}`);
    console.log(`   Block Number: ${info.blockNumber}\n`);
  } catch (error) {
    console.error('❌ Failed to read contract:', error.message);
    process.exit(1);
  }

  console.log(`🚀 Creating ${sampleData.length} sample records...\n`);

  const createdRecords = [];

  for (let i = 0; i < sampleData.length; i++) {
    const data = sampleData[i];
    
    // Generate unique hash for this data
    const dataContent = JSON.stringify(data);
    const dataHash = crypto.createHash('sha256').update(dataContent).digest('hex');

    console.log(`📝 Record ${i + 1}/${sampleData.length}: ${data.dataType}`);
    console.log(`   Description: ${data.description}`);
    console.log(`   Hash: 0x${dataHash.substring(0, 16)}...`);

    try {
      const receipt = await contract.methods.createRecord(
        '0x' + dataHash,
        data.dataType,
        data.description,
        data.metadata
      ).send({
        from: USER_ADDRESS,
        gas: 500000
      });

      const recordId = receipt.events.RecordCreated.returnValues.recordId;
      
      console.log(`   ✓ Created! Record ID: ${recordId}`);
      console.log(`   Tx: ${receipt.transactionHash}\n`);

      createdRecords.push({
        recordId: recordId.toString(),
        txHash: receipt.transactionHash,
        ...data
      });

      // Add trace steps for some records
      if (i === 0 || i === 2) {
        console.log(`   Adding trace step...`);
        await contract.methods.addTraceStep(
          recordId,
          'updated',
          `Quality check passed at ${new Date().toISOString()}`
        ).send({
          from: USER_ADDRESS,
          gas: 200000
        });
        console.log(`   ✓ Trace step added\n`);
      }

      // Small delay to avoid nonce issues
      await new Promise(resolve => setTimeout(resolve, 1000));

    } catch (error) {
      console.error(`   ❌ Failed:`, error.message);
    }
  }

  // Verify first record (if owner)
  console.log('\n🔍 Verifying first record...');
  try {
    const verifyReceipt = await contract.methods.verifyRecord(1).send({
      from: USER_ADDRESS,
      gas: 200000
    });
    console.log(`✓ Record #1 verified!`);
    console.log(`  Tx: ${verifyReceipt.transactionHash}\n`);
  } catch (error) {
    console.log(`⚠️  Verification skipped: ${error.message}\n`);
  }

  // Save created records info
  const recordsPath = path.join(__dirname, 'build', 'sample_records.json');
  fs.writeFileSync(recordsPath, JSON.stringify({
    contractAddress: address,
    userAddress: USER_ADDRESS,
    createdAt: new Date().toISOString(),
    records: createdRecords
  }, null, 2));

  console.log('╔══════════════════════════════════════════════════════════╗');
  console.log('║            SAMPLE DATA UPLOADED! ✅                      ║');
  console.log('╚══════════════════════════════════════════════════════════╝\n');

  console.log(`✅ Created ${createdRecords.length} records`);
  console.log(`💾 Record info saved to: ${recordsPath}`);
  console.log(`\n🌐 View in Explorer:`);
  console.log(`   http://localhost:8080/#/address/${address}`);
  console.log(`\n📋 Record IDs:`);
  createdRecords.forEach((record, idx) => {
    console.log(`   ${idx + 1}. ID ${record.recordId} - ${record.dataType}: ${record.description}`);
  });
}

// Run
if (require.main === module) {
  uploadSampleData().then(() => {
    console.log('\n✅ All done!');
    process.exit(0);
  }).catch(error => {
    console.error('\n❌ Error:', error);
    process.exit(1);
  });
}

module.exports = { uploadSampleData };

