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
import "solidity-bytes-utils/contracts/BytesLib.sol";
import "solidity-rlp/contracts/Helper.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/lib/TrieProofs.sol";

contract ParliaClient is ILightClient {
    using TrieProofs for bytes;
    using RLPReader for bytes;
    using RLPReader for RLPReader.RLPItem;
    using BytesLib for bytes;
    using IBCHeight for Height.Data;

    ///ibc.lightclients.${module}.v1.${resource}
    string private constant HEADER_TYPE_URL = "/ibc.lightclients.parlia.v1.Header";
    string private constant CLIENT_STATE_TYPE_URL = "/ibc.lightclients.parlia.v1.ClientState";
    string private constant CONSENSUS_STATE_TYPE_URL = "/ibc.lightclients.parlia.v1.ConsensusState";

    bytes32 private constant HEADER_TYPE_URL_HASH = keccak256(abi.encodePacked(HEADER_TYPE_URL));
    bytes32 private constant CLIENT_STATE_TYPE_URL_HASH =
        keccak256(abi.encodePacked(CLIENT_STATE_TYPE_URL));
    bytes32 private constant CONSENSUS_STATE_TYPE_URL_HASH =
        keccak256(abi.encodePacked(CONSENSUS_STATE_TYPE_URL));

    uint256 private constant COMMITMENT_SLOT = 0;
    uint8 private constant ACCOUNT_STORAGE_ROOT_INDEX = 2;

    address internal ibcHandler;
    mapping(string => ClientState.Data) internal clientStates;
    mapping(string => mapping(uint128 => ConsensusState.Data)) internal consensusStates;

    constructor(address ibcHandler_) {
        ibcHandler = ibcHandler_;
    }

    /**
     * @dev createClient creates a new client with the given state
     */
    function createClient(string calldata clientId, bytes calldata clientStateBytes, bytes calldata consensusStateBytes)
    external
    override
    onlyIBC
    returns (bytes32 clientStateCommitment, ConsensusStateUpdate memory update, bool ok) {
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

        Height.Data memory height;
        height.revision_height = clientState.latest_height.revision_height;
        height.revision_number = clientState.latest_height.revision_number;

        clientStates[clientId] = clientState;
        consensusStates[clientId][height.toUint128()] = consensusState;
        return (
            keccak256(clientStateBytes),
            ConsensusStateUpdate({consensusStateCommitment: keccak256(consensusStateBytes), height: height}),
            true
        );
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
        bytes32 stateRoot;
        bytes memory accountProof;
        Any.Data memory anyClientState;
        Any.Data memory anyConsensusState;

        (height, stateRoot, timestamp, accountProof) = parseHeader(clientMessageBytes);

        ClientState.Data storage clientState = clientStates[clientId];
        clientState.latest_height.revision_number = height.revision_number;
        clientState.latest_height.revision_height = height.revision_height;
        anyClientState.type_url = CLIENT_STATE_TYPE_URL;
        anyClientState.value = ClientState.encode(clientState);

        //TODO verify header

        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        consensusState.timestamp = timestamp;
        consensusState.state_root = abi.encodePacked(
            verifyStorageProof(BytesLib.toAddress(clientState.ibc_store_address, 0), stateRoot, accountProof));

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
        bytes calldata path,
        bytes calldata value
    ) external view override returns (bool) {
        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        require(consensusState.timestamp != 0,  "consensus state not found");
        return verifyMembership(
            proof,
            consensusState.state_root.toBytes32(0),
            keccak256(abi.encodePacked(keccak256(path), COMMITMENT_SLOT)),
            keccak256(value)
        );
    }

    /**
    * @dev verifyNonMembership is a generic proof verification method which verifies the absence of a given CommitmentPath at a specified height.
     * The caller is expected to construct the full CommitmentPath from a CommitmentPrefix and a standardized path (as defined in ICS 24).
     */
    function verifyNonMembership(
        string calldata clientId,
        Height.Data calldata height,
        uint64,
        uint64,
        bytes calldata proof,
        bytes memory,
        bytes calldata path
    ) external view override returns (bool) {
        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        require(consensusState.timestamp != 0,  "consensus state not found");
        return verifyNonMembership(
            proof, consensusState.state_root.toBytes32(0), keccak256(abi.encodePacked(keccak256(path), COMMITMENT_SLOT))
        );
    }

    // Same as IBFT2Client.sol
    function verifyMembership(bytes calldata proof, bytes32 root, bytes32 slot, bytes32 expectedValue)
    internal
    pure
    returns (bool)
    {
        bytes32 path = keccak256(abi.encodePacked(slot));
        bytes memory dataHash = proof.verify(root, path);
        return expectedValue == bytes32(dataHash.toRlpItem().toUint());
    }

    function verifyNonMembership(bytes calldata proof, bytes32 root, bytes32 slot) internal pure returns (bool) {
        // bytes32 path = keccak256(abi.encodePacked(slot));
        // bytes memory dataHash = proof.verify(root, path); // reverts if proof is invalid
        // return dataHash.toRlpItem().toBytes().length == 0;
        revert("not implemented");
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

    function parseHeader(bytes memory bz) internal pure returns (Height.Data memory, bytes32, uint64, bytes memory) {
        Any.Data memory any = Any.decode(bz);
        require(keccak256(abi.encodePacked(any.type_url)) == HEADER_TYPE_URL_HASH, "invalid header type");
        Header.Data memory header = Header.decode(any.value);
        bytes memory rlpEthHeader  = header.headers[0].header;

        RLPReader.RLPItem[] memory items = rlpEthHeader.toRlpItem().toList();
        Height.Data memory height = Height.Data({revision_number: 0, revision_height: uint64(items[8].toUint())});
        uint64 timestamp = uint64(items[11].toUint());
        bytes32 stateRoot = items[3].toBytes().toBytes32(0);
        return (height,stateRoot, timestamp, header.account_proof);
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

    function verifyStorageProof(address account, bytes32 stateRoot, bytes memory accountStateProof)
    internal
    pure
    returns (bytes32)
    {
        bytes32 proofPath = keccak256(abi.encodePacked(account));
        bytes memory accountRLP = accountStateProof.verify(stateRoot, proofPath); // reverts if proof is invalid
        return bytes32(accountRLP.toRlpItem().toList()[ACCOUNT_STORAGE_ROOT_INDEX].toUint());
    }

    modifier onlyIBC() {
        require(msg.sender == ibcHandler);
        _;
    }
}
