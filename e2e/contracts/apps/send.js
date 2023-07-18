const ICS20TransferBank = artifacts.require("ICS20TransferBank");
const ICS20Bank = artifacts.require("ICS20Bank");

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms))

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const alice = accounts[0];
  const bob = accounts[1];

  const port = "transfer";
  const channel = "channel-0";
  const timeoutHeight = 0;

  try {
    // Mint
    const bank = await ICS20Bank.deployed()
    const initialAliceAmount = await bank.balanceOf(alice, "simple")
    console.log("before = ", initialAliceAmount.toString())

    const mintResult = await bank.mint(alice, "simple", 100, {
      from: alice
    });
    console.log("mint success", mintResult.tx);

    // Send to counterparty chain
    const transfer = await ICS20TransferBank.deployed()
    const transferResult = await transfer.sendTransfer("simple", 20, bob, port, channel, timeoutHeight, {
      from: alice,
    });
    console.log("send success", transferResult.tx);

    // Wait for chain B receive the packet
    await sleep(10000)
    callback()

  }catch (e) {
    callback(e);
  }
};
