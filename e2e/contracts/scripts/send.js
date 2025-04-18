const util = require("./util");

async function main() {
  const accounts = await hre.ethers.getSigners();
  const alice = accounts[0].address;
  const bob = accounts[1].address;

  const port = "transfer";
  const channel = "channel-0";
  const timeoutHeight = 10000000;
  const mintAmount = 100;
  const sendingAmount = 20

  try {
    // Mint
    const bank = await util.readContract("ICS20Bank");
    const mintResult = await bank.mint(alice, "simple_erc_20_token_for_test", mintAmount, {
      from: alice
    });
    const mintReceipt = await mintResult.wait()
    console.log("mint success", mintReceipt.hash);

    // Send to counterparty chain
    const transfer = await util.readContract("ICS20TransferBank");
    const transferResult = await transfer.sendTransfer("simple_erc_20_token_for_test", sendingAmount, bob, port, channel, timeoutHeight, {
      from: alice,
    });
    const transferReceipt = await transferResult.wait()
    console.log("send success", transferReceipt.hash);

    // Check reduced amount
    const aliceAmount = await bank.balanceOf(alice, "simple_erc_20_token_for_test")
    console.log("after = ", aliceAmount.toString())
    if (parseInt(aliceAmount.toString(), 10) !== mintAmount - sendingAmount) {
      throw new Error("alice amount error");
    }

    // Check escrow balance
    const escrowAmount = await bank.balanceOf(transfer.target, "simple_erc_20_token_for_test")
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