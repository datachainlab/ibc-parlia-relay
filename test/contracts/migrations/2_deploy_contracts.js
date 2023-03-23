const IBCClient = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCClient");
const IBCConnection = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCConnection");
const IBCChannelHandshake = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelHandshake");
const IBCPacket = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCPacket");
const IBCCommitmentTestHelper = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCCommitmentTestHelper");
const IBCHandler = artifacts.require("@hyperledger-labs/yui-ibc-solidity/OwnableIBCHandler");
const SimpleToken = artifacts.require("@hyperledger-labs/yui-ibc-solidity/SimpleToken");
const ICS20TransferBank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20TransferBank");
const ICS20Bank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20Bank");
const MockClient = artifacts.require("MockClient");
const ParliaClient = artifacts.require("ParliaClient");

module.exports = async function (deployer) {
  await deployer.deploy(IBCClient);
  await deployer.deploy(IBCConnection);
  await deployer.deploy(IBCChannelHandshake);
  await deployer.deploy(IBCPacket);
  await deployer.deploy(IBCHandler, IBCClient.address, IBCConnection.address, IBCChannelHandshake.address, IBCPacket.address);

  await deployer.deploy(MockClient, IBCHandler.address);
  await deployer.deploy(ParliaClient, IBCHandler.address);
  await deployer.deploy(SimpleToken, "simple", "simple", 1000000);
  await deployer.deploy(ICS20Bank);
  await deployer.deploy(ICS20TransferBank, IBCHandler.address, ICS20Bank.address);

  await deployer.deploy(IBCCommitmentTestHelper);
};
