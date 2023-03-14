// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

import "@hyperledger-labs/yui-ibc-solidity/contracts/core/02-client/ILightClient.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/02-client/IBCHeight.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/proto/Client.sol";
import {
    IbcLightclientsParliaV1ClientState as ClientState,
    IbcLightclientsParliaV1ConsensusState as ConsensusState,
    IbcLightclientsParliaV1Header as Header
} from "../ibc/lightclients/parlia/v1/parlia.sol";
import {GoogleProtobufAny as Any} from "@hyperledger-labs/yui-ibc-solidity/contracts/proto/GoogleProtobufAny.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/lib/Bytes.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/lib/RLP.sol";

contract MockParliaClient is ILightClient {
    using RLP for bytes;
    using RLP for RLP.RLPItem;
    using Bytes for bytes;
    using IBCHeight for Height.Data;

    string private constant HEADER_TYPE_URL = "/ibc.lightclients.parlia.v1.Header";
    string private constant CLIENT_STATE_TYPE_URL = "/ibc.lightclients.parlia.v1.ClientState";
    string private constant CONSENSUS_STATE_TYPE_URL = "/ibc.lightclients.parlia.v1.ConsensusState";

    bytes32 private constant HEADER_TYPE_URL_HASH = keccak256(abi.encodePacked(HEADER_TYPE_URL));
    bytes32 private constant CLIENT_STATE_TYPE_URL_HASH =
        keccak256(abi.encodePacked(CLIENT_STATE_TYPE_URL));
    bytes32 private constant CONSENSUS_STATE_TYPE_URL_HASH =
        keccak256(abi.encodePacked(CONSENSUS_STATE_TYPE_URL));

    address internal ibcHandler;
    mapping(string => ClientState.Data) internal clientStates;
    mapping(string => mapping(uint128 => ConsensusState.Data)) internal consensusStates;

    constructor(address ibcHandler_) {
        ibcHandler = ibcHandler_;
    }

    /**
     * @dev createClient creates a new client with the given state
     */
    function createClient(
        string calldata clientId,
        Height.Data calldata height,
        bytes calldata clientStateBytes,
        bytes calldata consensusStateBytes
    ) external onlyIBC override returns (bytes32 clientStateCommitment, ConsensusStateUpdate memory update, bool ok) {
        ClientState.Data memory clientState;
        ConsensusState.Data memory consensusState;

        (clientState, ok) = unmarshalClientState(clientStateBytes);
        if (!ok) {
            return (clientStateCommitment, update, false);
        }
        (consensusState, ok) = unmarshalConsensusState(consensusStateBytes);
        if (!ok) {
            return (clientStateCommitment, update, false);
        }
        if (
            clientState.latest_height.revision_height == 0 || consensusState.timestamp == 0
        ) {
            return (clientStateCommitment, update, false);
        }
        clientStates[clientId] = clientState;
        consensusStates[clientId][height.toUint128()] = consensusState;
        return (keccak256(clientStateBytes), ConsensusStateUpdate({consensusStateCommitment: keccak256(consensusStateBytes), height: height}), true);
    }

    /**
     * @dev getTimestampAtHeight returns the timestamp of the consensus state at the given height.
     */
    function getTimestampAtHeight(string calldata clientId, Height.Data calldata height)
        external
        view
        override
        returns (uint64, bool)
    {
        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        return (consensusState.timestamp, consensusState.timestamp != 0);
    }

    /**
     * @dev getLatestHeight returns the latest height of the client state corresponding to `clientId`.
     */
    function getLatestHeight(string calldata clientId) external view override returns (Height.Data memory, bool) {
        ClientState.Data storage clientState = clientStates[clientId];
        return (Height.Data({revision_number: 0, revision_height: clientState.latest_height.revision_height}), clientState.latest_height.revision_height != 0);
    }

    /**
     * @dev updateClient is intended to perform the followings:
     * 1. verify a given client message(e.g. header)
     * 2. check misbehaviour such like duplicate block height
     * 3. if misbehaviour is found, update state accordingly and return
     * 4. update state(s) with the client message
     * 5. persist the state(s) on the host
     */
    function updateClient(string calldata clientId, bytes calldata clientMessageBytes)
        external
        onlyIBC
        override
        returns (bytes32 clientStateCommitment, ConsensusStateUpdate[] memory updates, bool ok)
    {
        Height.Data memory height;
        uint64 timestamp;
        Any.Data memory anyClientState;
        Any.Data memory anyConsensusState;

        (height, timestamp) = parseHeader(clientMessageBytes);
        anyClientState.type_url = CLIENT_STATE_TYPE_URL;
        anyClientState.value = ClientState.encode(clientStates[clientId]);

        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        consensusState.timestamp = timestamp;

        anyConsensusState.type_url = CONSENSUS_STATE_TYPE_URL;
        anyConsensusState.value = ConsensusState.encode(consensusState);

        updates = new ConsensusStateUpdate[](1);
        updates[0] =
            ConsensusStateUpdate({consensusStateCommitment: keccak256(Any.encode(anyConsensusState)), height: height});
        return (keccak256(Any.encode(anyClientState)), updates, true);
    }

    /**
     * @dev verifyMembership is a generic proof verification method which verifies a proof of the existence of a value at a given CommitmentPath at the specified height.
     * The caller is expected to construct the full CommitmentPath from a CommitmentPrefix and a standardized path (as defined in ICS 24).
     */
    function verifyMembership(
        string calldata clientId,
        Height.Data calldata height,
        uint64,
        uint64,
        bytes calldata proof,
        bytes memory,
        bytes memory,
        bytes calldata value
    ) external view override returns (bool) {
        require(consensusStates[clientId][height.toUint128()].timestamp != 0, "consensus state not found");
        return sha256(value) == proof.toBytes32();
    }

    /* State accessors */

    /**
     * @dev getClientState returns the clientState corresponding to `clientId`.
     *      If it's not found, the function returns false.
     */
    function getClientState(
        string calldata clientId
    ) external view returns (bytes memory clientStateBytes, bool) {
        ClientState.Data storage clientState = clientStates[clientId];
        return (Any.encode(Any.Data({
            type_url: CLIENT_STATE_TYPE_URL,
            value: ClientState.encode(clientState)
        })), true);
    }

    /**
     * @dev getConsensusState returns the consensusState corresponding to `clientId` and `height`.
     *      If it's not found, the function returns false.
     */
    function getConsensusState(
        string calldata clientId,
        Height.Data calldata height
    ) external view returns (bytes memory consensusStateBytes, bool) {
        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        return (Any.encode(Any.Data({
            type_url: CONSENSUS_STATE_TYPE_URL,
            value: ConsensusState.encode(consensusState)
        })), true);
    }

    /* Internal functions */

    function parseHeader(bytes memory bz) internal pure returns (Height.Data memory, uint64) {
        Any.Data memory any = Any.decode(bz);
        require(keccak256(abi.encodePacked(any.type_url)) == HEADER_TYPE_URL_HASH, "invalid header type");
        Header.Data memory header = Header.decode(any.value);
        bytes memory rlpEthHeader  = header.headers[0].header;

        RLP.RLPItem[] memory items = rlpEthHeader.toRLPItem().toList();
        return (Height.Data({revision_number: 0, revision_height: uint64(items[8].toUint())}), uint64(items[11].toUint()));
    }

    function unmarshalClientState(bytes calldata bz)
        internal
        pure
        returns (ClientState.Data memory clientState, bool ok)
    {
        Any.Data memory anyClientState = Any.decode(bz);
        if (keccak256(abi.encodePacked(anyClientState.type_url)) != CLIENT_STATE_TYPE_URL_HASH) {
            return (clientState, false);
        }
        return (ClientState.decode(anyClientState.value), true);
    }

    function unmarshalConsensusState(bytes calldata bz)
        internal
        pure
        returns (ConsensusState.Data memory consensusState, bool ok)
    {
        Any.Data memory anyConsensusState = Any.decode(bz);
        if (keccak256(abi.encodePacked(anyConsensusState.type_url)) != CONSENSUS_STATE_TYPE_URL_HASH) {
            return (consensusState, false);
        }
        return (ConsensusState.decode(anyConsensusState.value), true);
    }

    modifier onlyIBC() {
        require(msg.sender == ibcHandler);
        _;
    }
}
