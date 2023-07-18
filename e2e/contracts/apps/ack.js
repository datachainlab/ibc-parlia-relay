const ICS20Bank = artifacts.require("ICS20Bank");
const ICS20TransferBank = artifacts.require("ICS20TransferBank");

module.exports = async (callback) => {
  try {
    const escrow = await ICS20TransferBank.deployed()
    const bank = await ICS20Bank.deployed()
    const escrowAmount = await bank.balanceOf(escrow.address, "simple")
    console.log("escrow = ", escrowAmount.toString())
    if (parseInt(escrowAmount.toString(), 10) !== 0) {
      return callback("escrow amount error");
    }
    callback()

  }catch (e) {
    callback(e);
  }
};
