const util = require("./util");

async function main() {
  const accounts = await hre.ethers.getSigners();
  const bob = accounts[1]
  try {

    const token= await util.readContract("ERC20Token");
    const baseDenom = token.target.toLowerCase();
    const transfer = await util.readContract("ICS20Transfer");
    const bobAmount = await transfer.balanceOf(bob, `transfer/${util.config.channel}/${baseDenom}`)
    console.log("received = ", bobAmount.toString())
    if (parseInt(bobAmount.toString(), 10) !== util.config.amount) {
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