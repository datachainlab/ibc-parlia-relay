require("@nomicfoundation/hardhat-toolbox");
require("@openzeppelin/hardhat-upgrades");

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
    solidity: {
        version: "0.8.20",
        settings: {
            optimizer: {
                enabled: true,
                runs: 9_999_999
            }
        },
    },
    networks: {
        bsc_local1: {
            url: 'http://localhost:8545',
            accounts: {
                mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
                path: "m/44'/60'/0'/0"
            },
        },
        bsc_local2: {
            url: 'http://localhost:8645',
            accounts: {
                mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
                path: "m/44'/60'/0'/0"
            },
        }
    }
}