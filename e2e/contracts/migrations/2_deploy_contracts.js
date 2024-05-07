const IBCClient = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCClient");
const IBCConnection = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCConnectionSelfStateNoValidation");
const IBCChannelHandshake = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelHandshake");
const IBCChannelPacketSendRecv = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelPacketSendRecv");
const IBCChannelPacketTimeout = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelPacketTimeout");
const IBCChannelUpgradeInitTryAck = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelUpgradeInitTryAck");
const IBCChannelUpgradeConfirmTimeoutCancel = artifacts.require("@hyperledger-labs/yui-ibc-solidity/IBCChannelUpgradeConfirmTimeoutCancel");
const IBCHandler = artifacts.require("@hyperledger-labs/yui-ibc-solidity/OwnableIBCHandler");
const ERC20Token = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ERC20Token");
const ICS20TransferBank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20TransferBank");
const ICS20Bank = artifacts.require("@hyperledger-labs/yui-ibc-solidity/ICS20Bank");
const ParliaClient = artifacts.require("ParliaClient");

module.exports = async function (deployer) {
  await deployer.deploy(IBCClient);
  await deployer.deploy(IBCConnection);
  await deployer.deploy(IBCChannelHandshake);
  await deployer.deploy(IBCChannelPacketSendRecv);
  await deployer.deploy(IBCChannelPacketTimeout);
  await deployer.deploy(IBCChannelUpgradeInitTryAck);
  await deployer.deploy(IBCChannelUpgradeConfirmTimeoutCancel);
  await deployer.deploy(IBCHandler,
      IBCClient.address,
      IBCConnection.address,
      IBCChannelHandshake.address,
      IBCChannelPacketSendRecv.address,
      IBCChannelPacketTimeout.address,
      IBCChannelUpgradeInitTryAck.address,
      IBCChannelUpgradeConfirmTimeoutCancel.address,
  );

  await deployer.deploy(ParliaClient, IBCHandler.address);
  await deployer.deploy(ERC20Token, "simple_erc_20_token_for_test", "simple_erc_20_token_for_test", 1000000);
  await deployer.deploy(ICS20Bank);
  await deployer.deploy(ICS20TransferBank, IBCHandler.address, ICS20Bank.address);
};
