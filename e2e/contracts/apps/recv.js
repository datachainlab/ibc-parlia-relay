const ICS20Bank = artifacts.require("ICS20Bank");

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const bob = accounts[1]

  try {

    const bank = await ICS20Bank.deployed()
    const bobAmount = await bank.balanceOf(bob, `transfer/channel-0/simple`)
    console.log("received = ", bobAmount.toString())
    if (parseInt(bobAmount.toString(), 10) !== 20) {
      return callback("bob amount error");
    }
    callback()

  }catch (e) {
    console.log(e)
    callback(e);
  }
};
