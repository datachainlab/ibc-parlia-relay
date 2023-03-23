const ICS20TransferBank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20TransferBank");
const ICS20Bank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20Bank");

const bankAddr = "0x2F5703804E29F4252FA9405B8D357220d11b3bd9";
const transferAddr = "0xaE1C9125BbcF63bf51294C4D15CBD472782E330D";

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const alice = accounts[0];
  const bob = accounts[1];

  const mintAmount = 100;
  const sendAmount = 50;
  const port = "transfer";
  const channel = "channel-0";
  const timeoutHeight = 0;

  // Mint
  const bank = await ICS20Bank.at(bankAddr)
  const mintResult = await bank.mint(alice, "simple", mintAmount, {
    from: alice
  });
  console.log("mint success", mintResult.tx);

  // Send to counterparty chain
  const transfer = await ICS20TransferBank.at(transferAddr)
  const transferResult = await transfer.sendTransfer("simple", sendAmount, bob, port, channel, timeoutHeight, {
    from: alice,
  });
  console.log("send success", transferResult.tx);

  //const logs = await transfer.getPastEvents("SendTransfer", {
   // filter: { from: accounts[0], to: accounts[1] },
  //})
  //console.log(logs[0])
  callback();
};
