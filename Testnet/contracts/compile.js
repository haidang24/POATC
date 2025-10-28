// Compile Solidity contract using solc
const solc = require('solc');
const fs = require('fs');
const path = require('path');

console.log('📝 Compiling DataTraceability.sol...\n');

// Read the contract source
const contractPath = path.join(__dirname, 'DataTraceability.sol');
const source = fs.readFileSync(contractPath, 'utf8');

// Prepare input for compiler
const input = {
  language: 'Solidity',
  sources: {
    'DataTraceability.sol': {
      content: source
    }
  },
  settings: {
    outputSelection: {
      '*': {
        '*': ['abi', 'evm.bytecode']
      }
    },
    optimizer: {
      enabled: true,
      runs: 200
    }
  }
};

// Compile
const output = JSON.parse(solc.compile(JSON.stringify(input)));

// Check for errors
if (output.errors) {
  output.errors.forEach(error => {
    if (error.severity === 'error') {
      console.error('❌ Compilation Error:', error.formattedMessage);
      process.exit(1);
    } else {
      console.warn('⚠️  Warning:', error.formattedMessage);
    }
  });
}

// Extract compiled contract
const contract = output.contracts['DataTraceability.sol']['DataTraceability'];

if (!contract) {
  console.error('❌ Contract not found in compilation output');
  process.exit(1);
}

// Save ABI and Bytecode
const outputDir = path.join(__dirname, 'build');
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir);
}

const abi = contract.abi;
const bytecode = contract.evm.bytecode.object;

fs.writeFileSync(
  path.join(outputDir, 'DataTraceability.abi.json'),
  JSON.stringify(abi, null, 2)
);

fs.writeFileSync(
  path.join(outputDir, 'DataTraceability.bytecode.txt'),
  bytecode
);

// Save combined
fs.writeFileSync(
  path.join(outputDir, 'DataTraceability.json'),
  JSON.stringify({
    contractName: 'DataTraceability',
    abi: abi,
    bytecode: '0x' + bytecode,
    compiler: {
      name: 'solc',
      version: solc.version()
    }
  }, null, 2)
);

console.log('✅ Compilation successful!\n');
console.log('📁 Output files:');
console.log(`   - ${path.join(outputDir, 'DataTraceability.abi.json')}`);
console.log(`   - ${path.join(outputDir, 'DataTraceability.bytecode.txt')}`);
console.log(`   - ${path.join(outputDir, 'DataTraceability.json')}`);
console.log('\n📊 Contract info:');
console.log(`   - Bytecode size: ${bytecode.length / 2} bytes`);
console.log(`   - ABI methods: ${abi.filter(item => item.type === 'function').length}`);
console.log(`   - Events: ${abi.filter(item => item.type === 'event').length}`);
console.log(`   - Compiler: ${solc.version()}`);

