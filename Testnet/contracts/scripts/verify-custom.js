const hre = require("hardhat");

async function main() {
  console.log("🔍 Verifying DataTraceability contract with custom source...");
  
  // Get the deployed contract address from deployment.json
  const fs = require('fs');
  const path = require('path');
  const deploymentPath = path.join(__dirname, '../build/deployment.json');
  
  if (!fs.existsSync(deploymentPath)) {
    console.error("❌ Deployment file not found. Please deploy the contract first.");
    process.exit(1);
  }
  
  const deployment = JSON.parse(fs.readFileSync(deploymentPath, 'utf8'));
  const contractAddress = deployment.address;
  
  console.log(`📍 Contract Address: ${contractAddress}`);
  
  // Read the source code
  const sourcePath = path.join(__dirname, '../contracts/DataTraceability.sol');
  const sourceCode = fs.readFileSync(sourcePath, 'utf8');
  
  try {
    // Verify the contract with custom source
    await hre.run("verify:verify", {
      address: contractAddress,
      constructorArguments: [], // No constructor arguments for DataTraceability
      contract: "contracts/DataTraceability.sol:DataTraceability",
      sourceCode: sourceCode,
    });
    
    console.log("✅ Contract verified successfully with custom source!");
    console.log(`🌐 View on explorer: http://localhost:8080/#/address/${contractAddress}`);
    
  } catch (error) {
    if (error.message.includes("Already Verified")) {
      console.log("✅ Contract is already verified!");
    } else {
      console.error("❌ Verification failed:", error.message);
      console.log("💡 Try running: npx hardhat verify --network poatc", contractAddress);
      process.exit(1);
    }
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
