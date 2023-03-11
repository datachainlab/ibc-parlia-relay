const HDWalletProvider = require('@truffle/hdwallet-provider');
const mnemonic =
  'math razor capable expose worth grape metal sunset metal sudden usage scheme';

module.exports = {
  networks: {
    bsc_local: {
      network_id: '*', // Any network (default: none)
      provider: () =>
        new HDWalletProvider({
          mnemonic: {
            phrase: mnemonic,
          },
          providerOrUrl: 'http://127.0.0.1:8545',
          addressIndex: 0,
          numberOfAddresses: 5,
        }),
    },
    eth_local: {
      host: '127.0.0.1',
      port: 8645,
      network_id: '2018',
    },
  },

  mocha: {
    // timeout: 100000
  },

  compilers: {
    solc: {
      version: '0.8.9', // Fetch exact version from solc-bin (default: truffle's version)
      // docker: true,        // Use "0.5.1" you've installed locally with docker (default: false)
      settings: {
        // See the solidity docs for advice about optimization and evmVersion
        optimizer: {
          enabled: true,
          runs: 1000,
        },
        //  evmVersion: "byzantium"
      },
    },
  },
};
