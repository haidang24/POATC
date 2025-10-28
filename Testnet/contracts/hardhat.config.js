require('@nomicfoundation/hardhat-toolbox');
require('@nomicfoundation/hardhat-verify');
// require('dotenv').config();

const RPC_URL = 'http://127.0.0.1:8545';
const DEPLOYER ='0x05c72e4c1ef832ca53ec61b4398a049368913cb8a7c6e7e421a55aa1d16ec8f2';

module.exports = {
  solidity: {
    version: '0.8.20',
    settings: {
      optimizer: { enabled: true, runs: 200 },
    },
  },
  defaultNetwork: 'poatc',
  networks: {
    poatc: {
      url: RPC_URL,
      accounts: [DEPLOYER], 
    },
  },
};
