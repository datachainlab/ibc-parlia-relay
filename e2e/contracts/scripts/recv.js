async function readContract(contractName) {
  const fs = require("fs");
  const path = require("path");

  const filepath = path.join("addresses", contractName);
  const address = fs.readFileSync(filepath, "utf-8");
  return await hre.ethers.getContractAt(contractName, address);
}

async function main() {
  const accounts = await hre.ethers.getSigners();
  const bob = accounts[1]

  try {

    const bank = await readContract("ICS20Bank");
    const bobAmount = await bank.balanceOf(bob, `transfer/channel-0/simple_erc_20_token_for_test`)
    console.log("received = ", bobAmount.toString())
    if (parseInt(bobAmount.toString(), 10) !== 20) {
      throw new Error("bob amount error");
    }

  }catch (e) {
    console.log(e)
  }
}

if (require.main === module) {
  main()
      .then(() => process.exit(0))
      .catch((error) => {
        console.error(error);
        process.exit(1);
      });
}