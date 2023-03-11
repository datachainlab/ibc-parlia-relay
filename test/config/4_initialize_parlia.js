const IBCHandler = artifacts.require("OwnableIBCHandler");
const MockClient = artifacts.require("MockClient");

const ParliaClientType = "parlia-client"

module.exports = async function (deployer) {
  const ibcHandler = await IBCHandler.deployed();

  for(const f of [
    () => ibcHandler.registerClient(ParliaClientType, MockClient.address),
  ]) {
    const result = await f();
    if(!result.receipt.status) {
      console.log(result);
      throw new Error(`transaction failed to execute. ${result.tx}`);
    }
  }
};
