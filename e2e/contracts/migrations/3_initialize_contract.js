const IBCHandler = artifacts.require("OwnableIBCHandler");
const ICS20TransferBank = artifacts.require("ICS20TransferBank");
const ICS20Bank = artifacts.require("ICS20Bank");
const ParliaClient = artifacts.require("ParliaClient");

const PortTransfer = "transfer"
const ParliaClientType = "xx-parlia"

module.exports = async function (deployer, network, accounts) {
  const ibcHandler = await IBCHandler.deployed();
  const ics20Bank = await ICS20Bank.deployed();

  for(const f of [
    () => ibcHandler.bindPort(PortTransfer, ICS20TransferBank.address),
    () => ibcHandler.registerClient(ParliaClientType, ParliaClient.address),
    () => ics20Bank.setOperator(ICS20TransferBank.address),
    () => ics20Bank.setOperator(accounts[0]), // make deployer mintable
  ]) {
    const result = await f();
    if(!result.receipt.status) {
      console.log(result);
      throw new Error(`transaction failed to execute. ${result.tx}`);
    }
  }
};

