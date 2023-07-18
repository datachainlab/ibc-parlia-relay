const ICS20Bank = artifacts.require("ICS20Bank");

module.exports = async (callback) => {
  const accounts = await web3.eth.getAccounts();
  const alice = accounts[0];

  try {

    const bank = await ICS20Bank.deployed()
    const aliceAmount = await bank.balanceOf(alice, "simple")
    console.log("after = ", aliceAmount.toString())
    if (parseInt(aliceAmount.toString(), 10) !== 80) {
      callback("amount error");
    } else {
      callback()
    }

  }catch (e) {
    callback(e);
  }
};
