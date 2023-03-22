const IBFT2Client = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBFT2Client");
const MockClient = artifacts.require("@hyperledger-labs/yui-ibc-solidity/MockClient");
const IBCClient = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCClient");
const IBCConnection = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCConnection");
const IBCChannel = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannel");
const IBCHandler = artifacts.require("@hyperledger-labs/yui-ibc-solidity/OwnableIBCHandler");
const IBCMsgs = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCMsgs");
const IBCCommitment = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCCommitment");
const SimpleToken = artifacts.require("@hyperledger-labs/yui-ibc-solidity/SimpleToken");
const ICS20TransferBank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20TransferBank");
const ICS20Bank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20Bank");
const ParliaClient = artifacts.require("ParliaClient");

module.exports = async function (deployer) {
  await deployer.deploy(IBCCommitment);
  await deployer.link(IBCCommitment, [IBCHandler, IBCClient, IBCConnection, IBCChannel]);

  await deployer.deploy(IBCMsgs);
  await deployer.link(IBCMsgs, [IBCClient, IBCConnection, IBCChannel, IBCHandler]);

  await deployer.deploy(IBCClient);
  await deployer.deploy(IBCConnection);
  await deployer.deploy(IBCChannel);
  await deployer.deploy(IBCHandler, IBCClient.address, IBCConnection.address, IBCChannel.address, IBCChannel.address);

  await deployer.deploy(MockClient, IBCHandler.address);
  await deployer.deploy(ParliaClient, IBCHandler.address);
  await deployer.deploy(IBFT2Client, IBCHandler.address);
  await deployer.deploy(SimpleToken, "simple", "simple", 1000000);
  await deployer.deploy(ICS20Bank);
  await deployer.deploy(ICS20TransferBank, IBCHandler.address, ICS20Bank.address);
};
