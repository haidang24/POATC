const hre = require('hardhat');

async function main() {
  console.log('\nDeploying DataTraceability via Hardhat...');

  const DataTraceability = await hre.ethers.getContractFactory('DataTraceability');
  const contract = await DataTraceability.deploy();
  await contract.waitForDeployment();

  const address = await contract.getAddress();
  console.log('✓ Deployed at:', address);

  // Save deployment file compatible with existing scripts
  const fs = require('fs');
  const path = require('path');
  const outDir = path.join(__dirname, '..', 'build');
  if (!fs.existsSync(outDir)) fs.mkdirSync(outDir, { recursive: true });

  const artifact = await hre.artifacts.readArtifact('DataTraceability');

  const deployment = {
    contractName: 'DataTraceability',
    address,
    deployer: process.env.DEPLOYER_ADDRESS || '0x3003d6498603fAD5F232452B21c8B6EB798d20f1',
    deployedAt: new Date().toISOString(),
    network: 'POATC Testnet',
    rpcUrl: process.env.RPC_URL || 'http://127.0.0.1:8545',
    abi: artifact.abi,
  };

  fs.writeFileSync(path.join(outDir, 'deployment.json'), JSON.stringify(deployment, null, 2));
  console.log('Saved build/deployment.json');
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
