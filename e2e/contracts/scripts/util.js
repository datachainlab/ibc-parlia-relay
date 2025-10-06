async function readContract(contractName) {
    const fs = require("fs");
    const path = require("path");

    const filepath = path.join("addresses", contractName);
    const address = fs.readFileSync(filepath, "utf-8");
    return await hre.ethers.getContractAt(contractName, address);
}

const config = {
    channel:"channel-0",
    amount: 20
}

module.exports = {
    readContract,
    config
};