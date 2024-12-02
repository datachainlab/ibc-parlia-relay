const ICS20TransferBank = artifacts.require("ICS20TransferBank");
const ICS20Bank = artifacts.require("ICS20Bank");

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const alice = accounts[0];
  const bob = accounts[1];

  const port = "transfer";
  const channel = "channel-0";
  const timeoutHeight = 10000000;
  const mintAmount = 100;
  const sendingAmount = 20
  console.log(alice, bob)

  try {
    // Mint
    const bank = await ICS20Bank.deployed()
    const mintResult = await bank.mint(alice, "simple_erc_20_token_for_test", mintAmount, {
      from: alice
    });
    console.log("mint success", mintResult.tx);

    // Send to counterparty chain
    const transfer = await ICS20TransferBank.deployed()
    const transferResult = await transfer.sendTransfer("simple_erc_20_token_for_test", sendingAmount, bob, port, channel, timeoutHeight, {
      from: alice,
    });
    console.log("send success", transferResult.tx);

    // Check reduced amount
    const aliceAmount = await bank.balanceOf(alice, "simple_erc_20_token_for_test")
    console.log("after = ", aliceAmount.toString())
    if (parseInt(aliceAmount.toString(), 10) !== mintAmount - sendingAmount) {
      return callback("alice amount error");
    }

    // Check escrow balance
    const escrowAmount = await bank.balanceOf(transfer.address, "simple_erc_20_token_for_test")
    console.log("escrow = ", escrowAmount.toString())
    if (parseInt(escrowAmount.toString(), 10) !== sendingAmount) {
      return callback("escrow amount error");
    }
    // Wait for chain B receive the packet
    callback()

  }catch (e) {
    console.log(e)
    callback(e);
  }
};
