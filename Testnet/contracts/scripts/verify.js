const hre = require("hardhat");

async function main() {
  console.log("🔍 Verifying DataTraceability contract...");
  
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
  
  try {
    // Verify the contract
    await hre.run("verify:verify", {
      address: contractAddress,
      constructorArguments: [], // No constructor arguments for DataTraceability
    });
    
    console.log("✅ Contract verified successfully!");
    console.log(`🌐 View on explorer: http://localhost:8080/#/address/${contractAddress}`);
    
  } catch (error) {
    if (error.message.includes("Already Verified")) {
      console.log("✅ Contract is already verified!");
    } else {
      console.error("❌ Verification failed:", error.message);
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
