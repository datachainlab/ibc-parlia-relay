const IBCClient = artifacts.require("IBCClient");
const IBCConnection = artifacts.require("IBCConnectionSelfStateNoValidation");
const IBCChannelHandshake = artifacts.require("IBCChannelHandshake");
const IBCChannelPacketSendRecv = artifacts.require("IBCChannelPacketSendRecv");
const IBCChannelPacketTimeout = artifacts.require("IBCChannelPacketTimeout");
const IBCChannelUpgradeInitTryAck = artifacts.require("IBCChannelUpgradeInitTryAck");
const IBCChannelUpgradeConfirmTimeoutCancel = artifacts.require("IBCChannelUpgradeConfirmTimeoutCancel");
const IBCHandler = artifacts.require("OwnableIBCHandler");
const ERC20Token = artifacts.require("ERC20Token");
const ICS20TransferBank = artifacts.require("ICS20TransferBank");
const ICS20Bank = artifacts.require("ICS20Bank");
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
