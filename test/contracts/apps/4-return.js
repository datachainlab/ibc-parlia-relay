const MiniToken = artifacts.require("MiniToken");

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const alice = accounts[1]
  const bob = accounts[2];

  const mintAmount = 100;
  const sendAmount = 50;
  const port = "transfer";
  const channel = "channel-0";
  const timeoutHeight = 0;

  const miniToken = await MiniToken.deployed();
  await miniToken.sendTransfer(sendAmount, alice, port, channel, timeoutHeight, {
    from: bob,
  });
  callback();
};
