const util = require("./util");

async function main() {
  const accounts = await hre.ethers.getSigners();
  const alice = accounts[0].address;
  const bob = accounts[1].address;

  const channel = "channel-0";
  const timeoutHeight = [
    [BigInt(1), BigInt(10000000)],
    BigInt(0)
  ];
  const sendingAmount = 20

  try {
    const token = await util.readContract("ERC20Token");
    const before= await token.balanceOf(alice)
    console.log("before = ", before.toString())
    const transfer = await util.readContract("ICS20Transfer");

    // Deposit and SendTransfer
    await token.approve(transfer.target, sendingAmount, { from: alice }).then(tx => tx.wait());
    const depositResult = await transfer.depositSendTransfer(channel, token.target, sendingAmount, bob, timeoutHeight, {
      from: alice
    })
    const depositReceipt = await depositResult.wait()
    console.log("deposit success", depositReceipt.hash);

    // Check reduced amount
    const after= await token.balanceOf(alice)
    console.log("after = ", after.toString())
    if (parseInt(after.toString(), 10) !== Number(before) - sendingAmount) {
      throw new Error("alice amount error");
    }

    // Check escrow balance
    const escrowAmount = await token.balanceOf(transfer.target)
    console.log("escrow = ", escrowAmount.toString())
    if (parseInt(escrowAmount.toString(), 10) !== sendingAmount) {
      throw new Error("escrow amount error");
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