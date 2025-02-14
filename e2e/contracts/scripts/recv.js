const util = require("./util");

async function main() {
  const accounts = await hre.ethers.getSigners();
  const bob = accounts[1]

  try {

    const bank = await util.readContract("ICS20Bank");
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