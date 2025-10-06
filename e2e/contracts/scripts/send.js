const util = require("./util");

async function main() {
  const accounts = await hre.ethers.getSigners();
  const alice = accounts[0].address;
  const bob = accounts[1].address;
  const timeoutHeight = [[1, 10000000],0];
  try {
    const token = await util.readContract("ERC20Token");
    const transfer = await util.readContract("ICS20Transfer");

    const before= await token.balanceOf(alice)
    console.log("before = ", before.toString())

    const escrowBefore= await token.balanceOf(transfer.target)
    console.log("escrow before = ", escrowBefore.toString())

    // Deposit and SendTransfer
    await token.approve(transfer.target, util.config.amount, { from: alice }).then(tx => tx.wait());
    const depositResult = await transfer.depositSendTransfer(util.config.channel, token.target, util.config.amount, bob, timeoutHeight, {
      from: alice
    })
    const depositReceipt = await depositResult.wait()
    console.log("depositSendTransfer success", depositReceipt.hash);

    // Check reduced amount
    const after= await token.balanceOf(alice)
    console.log("after = ", after.toString())
    if (parseInt(after.toString(), 10) !== Number(before) - util.config.amount) {
      throw new Error("alice amount error");
    }

    // Check escrow balance
    const escrowAfter= await token.balanceOf(transfer.target)
    console.log("escrow after = ", escrowAfter.toString())
    if (parseInt(escrowAfter.toString(), 10) !== Number(escrowBefore) + util.config.amount) {
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