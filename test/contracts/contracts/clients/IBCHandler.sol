// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

import "@hyperledger-labs/yui-ibc-solidity/contracts/core/24-host/IBCHost.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IBCClientHandler.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IBCConnectionHandler.sol";
import "./IBCChannelHandler.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IBCPacketHandler.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IBCQuerier.sol";

/**
 * @dev IBCHandler is a contract that implements [ICS-25](https://github.com/cosmos/ibc/tree/main/spec/core/ics-025-handler-interface).
 */
abstract contract IBCHandler is
    IBCHost,
    IBCClientHandler,
    IBCConnectionHandler,
    IBCChannelHandler,
    IBCPacketHandler,
    IBCQuerier
{
    /**
     * @dev The arguments of constructor must satisfy the followings:
     * @param ibcClient is the address of a contract that implements `IIBCClient`.
     * @param ibcConnection is the address of a contract that implements `IIBCConnectionHandshake`.
     * @param ibcChannel is the address of a contract that implements `IIBCChannelHandshake`.
     * @param ibcPacket is the address of a contract that implements `IIBCPacket`.
     */
    constructor(address ibcClient, address ibcConnection, address ibcChannel, address ibcPacket)
        IBCClientHandler(ibcClient)
        IBCConnectionHandler(ibcConnection)
        IBCChannelHandler(ibcChannel)
        IBCPacketHandler(ibcPacket)
    {}
}
