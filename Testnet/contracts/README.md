# 🔐 POATC Smart Contracts - Data Traceability System

## 📋 Overview

Complete smart contract system for data traceability and provenance tracking on POATC blockchain.

### ✨ Features

- ✅ **Data Record Creation** - Create immutable records with hash verification
- ✅ **Trace History** - Complete audit trail for every record
- ✅ **Verification System** - Contract owner can verify records
- ✅ **Metadata Storage** - Flexible JSON metadata for any use case
- ✅ **Event Logging** - All actions emit events for transparency
- ✅ **User Management** - Track records by creator address

## 📁 Files

```
contracts/
├── DataTraceability.sol      # Main smart contract (Solidity)
├── compile.js                 # Compile contract to bytecode & ABI
├── deploy.js                  # Deploy contract to blockchain
├── upload_sample_data.js      # Upload test data
├── test_contract.js           # Test contract functions
├── deploy_and_test.ps1        # Complete deployment script
├── package.json               # NPM dependencies
└── build/                     # Compiled outputs (generated)
    ├── DataTraceability.json  # ABI + Bytecode
    ├── deployment.json        # Deployment info
    └── sample_records.json    # Created records
```

## 🚀 Quick Start

### Prerequisites

1. **Blockchain nodes running**
   ```powershell
   cd testnet
   .\restart_nodes.ps1
   ```

2. **Node.js installed** (v16+)

### Option 1: One-Click Deploy (Recommended)

```powershell
cd testnet/contracts
.\deploy_and_test.ps1
```

This will:
1. ✅ Install dependencies
2. ✅ Compile contract
3. ✅ Deploy to blockchain
4. ✅ Upload 5 sample records
5. ✅ Run tests

### Option 2: Step-by-Step

```powershell
cd testnet/contracts

# 1. Install dependencies
npm install

# 2. Compile contract
npm run compile

# 3. Deploy contract
npm run deploy

# 4. Upload sample data
npm run upload-data

# 5. Test contract
npm test
```

## 📊 Contract Functions

### Read Functions (View)

| Function | Description |
|----------|-------------|
| `getRecord(uint256 id)` | Get complete record details |
| `getTraceHistory(uint256 id)` | Get all trace steps for a record |
| `getUserRecords(address user)` | Get all records created by a user |
| `getTraceStepCount(uint256 id)` | Get number of trace steps |
| `getContractInfo()` | Get contract metadata |

### Write Functions (Transactions)

| Function | Description | Gas Estimate |
|----------|-------------|--------------|
| `createRecord(...)` | Create new data record | ~200,000 |
| `verifyRecord(uint256 id)` | Verify a record (owner only) | ~50,000 |
| `addTraceStep(...)` | Add trace step to record | ~80,000 |
| `transferOwnership(address)` | Transfer contract ownership | ~30,000 |

## 📝 Sample Data

The system includes 5 sample records:

1. **Product** - Organic Coffee Beans from Ethiopia
2. **Document** - Supply Chain Certificate
3. **Transaction** - Shipment tracking
4. **Product** - Medical Equipment (Surgical Masks)
5. **IoT Sensor Data** - Temperature monitoring

Each record includes:
- Unique SHA256 hash
- Data type classification
- Human-readable description
- JSON metadata with detailed info
- Complete trace history

## 🔍 Testing

### Run Tests

```powershell
npm test
```

### Expected Output

```
✅ Test 1: Get Contract Info
✅ Test 2: Get Record Details  
✅ Test 3: Get Trace History
✅ Test 4: Get User Records
✅ Test 5: Read Recent Events
```

### Manual Testing with Web3

```javascript
const { Web3 } = require('web3');
const web3 = new Web3('http://127.0.0.1:8545');

// Load contract
const deployment = require('./build/deployment.json');
const contract = new web3.eth.Contract(deployment.abi, deployment.address);

// Get record
const record = await contract.methods.getRecord(1).call();
console.log(record);

// Get trace history
const history = await contract.methods.getTraceHistory(1).call();
console.log(history);
```

## 🌐 Explorer Integration

After deployment, view contracts in the POATC Explorer:

```
http://localhost:8080/#/address/[CONTRACT_ADDRESS]
```

The explorer will show:
- Contract balance
- Transaction history
- Internal transactions
- Event logs
- Contract interactions

## 📋 Sample Record Structure

```json
{
  "recordId": "1",
  "dataHash": "0x742d35cc6634c0532925a3b844bc9c7eb6d45c...",
  "dataType": "Product",
  "description": "Organic Coffee Beans - Ethiopia Grade A",
  "creator": "0x3003d6498603fAD5F232452B21c8B6EB798d20f1",
  "timestamp": 1707926400,
  "metadata": {
    "origin": "Ethiopia, Yirgacheffe",
    "harvestDate": "2024-01-15",
    "quantity": "1000kg",
    "quality": "Grade A",
    "certifications": ["Organic", "Fair Trade"],
    "batchNumber": "ETH-2024-001"
  },
  "verified": true,
  "verifiedBy": "0x3003d6498603fAD5F232452B21c8B6EB798d20f1",
  "traceHistory": [
    {
      "action": "created",
      "actor": "0x3003d6498603fAD5F232452B21c8B6EB798d20f1",
      "timestamp": 1707926400,
      "details": "Record created on blockchain"
    },
    {
      "action": "updated",
      "actor": "0x3003d6498603fAD5F232452B21c8B6EB798d20f1",
      "timestamp": 1707926460,
      "details": "Quality check passed at 2024-02-14T12:34:56.000Z"
    },
    {
      "action": "verified",
      "actor": "0x3003d6498603fAD5F232452B21c8B6EB798d20f1",
      "timestamp": 1707926520,
      "details": "Record verified by contract owner"
    }
  ]
}
```

## 🔐 Security Features

- ✅ **Unique Hash Validation** - Prevents duplicate records
- ✅ **Owner-only Verification** - Only contract owner can verify
- ✅ **Creator Permissions** - Only record creator can add trace steps
- ✅ **Immutable Records** - Data cannot be modified after creation
- ✅ **Event Transparency** - All actions emit events
- ✅ **Access Control** - Modifiers protect critical functions

## 🛠️ Development

### Modify Contract

1. Edit `DataTraceability.sol`
2. Recompile: `npm run compile`
3. Redeploy: `npm run deploy`

### Add New Functions

```solidity
function myNewFunction(uint256 _param) public view returns (string memory) {
    // Your code here
    return "result";
}
```

Then recompile and redeploy.

## 📊 Gas Costs

Average gas costs (may vary):

| Operation | Gas Used |
|-----------|----------|
| Contract Deployment | ~2,500,000 |
| Create Record | ~200,000 |
| Verify Record | ~50,000 |
| Add Trace Step | ~80,000 |
| Read Record | Free (view) |
| Read History | Free (view) |

## 🐛 Troubleshooting

### "Account not unlocked"

Make sure blockchain node is started with `--unlock` flag:
```powershell
--unlock 0x3003d6498603fAD5F232452B21c8B6EB798d20f1 --password ./password.txt
```

### "Contract not deployed"

Run deployment first:
```powershell
npm run deploy
```

### "Connection refused"

Start blockchain nodes:
```powershell
cd testnet
.\restart_nodes.ps1
```

## 📈 Next Steps

1. ✅ Deploy contract
2. ✅ Upload sample data
3. ✅ Test all functions
4. 🔄 Integrate with Explorer UI
5. 🔄 Add contract verification
6. 🔄 Create frontend interaction UI

## 📚 Resources

- **Solidity Docs**: https://docs.soliditylang.org/
- **Web3.js Docs**: https://web3js.readthedocs.io/
- **POATC Blockchain**: Custom PoA consensus with anomaly detection

---

**Built for POATC Blockchain** - Proof-of-Authority with Anomaly Detection & Tracing

