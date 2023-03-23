const ICS20TransferBank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20TransferBank");
const ICS20Bank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20Bank");

const bankAddr = "0x78247fcA07EE10B76FCe148A69A00A726A4Ca003";
const transferAddr = "0xa0ba0aC3E2b19F99fA7Dc603c7A3ec711c30Bfd1";

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
