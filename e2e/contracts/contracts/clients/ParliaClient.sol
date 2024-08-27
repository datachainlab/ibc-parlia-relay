// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.12;

import "@hyperledger-labs/yui-ibc-solidity/contracts/core/02-client/ILightClient.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/core/02-client/IBCHeight.sol";
import "@hyperledger-labs/yui-ibc-solidity/contracts/proto/Client.sol";
import {
IbcLightclientsParliaV1ClientState as ClientState,
IbcLightclientsParliaV1ConsensusState as ConsensusState,
IbcLightclientsParliaV1Header as Header
} from "../ibc/lightclients/parlia/v1/parlia.sol";
import {GoogleProtobufAny as Any} from "@hyperledger-labs/yui-ibc-solidity/contracts/proto/GoogleProtobufAny.sol";
import {RLPReader} from "@hyperledger-labs/yui-ibc-solidity/contracts/clients/qbft/RLPReader.sol";
import {MPTProof} from "@hyperledger-labs/yui-ibc-solidity/contracts/clients/qbft/MPTProof.sol";

contract ParliaClient is ILightClient {
    using MPTProof for bytes;
    using RLPReader for bytes;
    using RLPReader for RLPReader.RLPItem;
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

    bytes32 private constant COMMITMENT_SLOT = 0x1ee222554989dda120e26ecacf756fe1235cd8d726706b57517715dde4f0c900;
    uint8 private constant ACCOUNT_STORAGE_ROOT_INDEX = 2;

    address internal ibcHandler;
    mapping(string => ClientState.Data) internal clientStates;
    mapping(string => mapping(uint128 => ConsensusState.Data)) internal consensusStates;
    mapping(string => ClientStatus) internal statuses;

    constructor(address ibcHandler_) {
        ibcHandler = ibcHandler_;
    }

    function getLatestInfo(string calldata clientId)
    public
    view
    returns (Height.Data memory latestHeight, uint64 latestTimestamp, ClientStatus status)
    {
        latestHeight = getLatestHeight(clientId);
        latestTimestamp = consensusStates[clientId][latestHeight.toUint128()].timestamp;
        status = statuses[clientId];
    }

    function getStatus(string calldata clientId) external view virtual override returns (ClientStatus) {
        return statuses[clientId];
    }

    /**
     * @dev initializeClient creates a new client with the given state
     */
    function initializeClient(string calldata clientId, bytes calldata clientStateBytes, bytes calldata consensusStateBytes)
    external virtual override onlyIBC
    returns (Height.Data memory height) {
        ClientState.Data memory clientState;
        ConsensusState.Data memory consensusState;
        bool ok;

        (clientState, ok) = unmarshalClientState(clientStateBytes);
        if (!ok) {
            revert("invalid client state");
        }
        (consensusState, ok) = unmarshalConsensusState(consensusStateBytes);
        if (!ok) {
            revert("invalid cons state");
        }

        height.revision_height = clientState.latest_height.revision_height;
        height.revision_number = clientState.latest_height.revision_number;

        clientStates[clientId] = clientState;
        consensusStates[clientId][height.toUint128()] = consensusState;
        statuses[clientId] = ClientStatus.Active;
        return Height.Data({revision_number: 0, revision_height: clientState.latest_height.revision_height});
    }

    /**
     * @dev getTimestampAtHeight returns the timestamp of the consensus state at the given height.
     */
    function getTimestampAtHeight(string calldata clientId, Height.Data calldata height)
    external
    view
    override
    returns (uint64)
    {
        ConsensusState.Data storage consensusState = consensusStates[clientId][height.toUint128()];
        return consensusState.timestamp;
    }

    /**
     * @dev getLatestHeight returns the latest height of the client state corresponding to `clientId`.
     */
    function getLatestHeight(string calldata clientId) public view virtual override returns (Height.Data memory) {
        ClientState.Data storage clientState = clientStates[clientId];
        return Height.Data({revision_number: 0, revision_height: clientState.latest_height.revision_height});
    }

    /**
    * @dev routeUpdateClient returns the calldata to the receiving function of the client message.
     *      The light client encodes a client message as ethereum ABI.
     */
    function routeUpdateClient(string calldata clientId, bytes calldata protoClientMessage) external pure virtual override returns (bytes4, bytes memory)
    {
        Any.Data memory any = Any.decode(protoClientMessage);
        Header.Data memory header = Header.decode(any.value);
        return (this.updateClient.selector, abi.encode(clientId, header));
    }

    /**
     * @dev updateClient is intended to perform the followings:
     * 1. verify a given client message(e.g. header)
     * 2. check misbehaviour such like duplicate block height
     * 3. if misbehaviour is found, update state accordingly and return
     * 4. update state(s) with the client message
     * 5. persist the state(s) on the host
     */
    function updateClient(string calldata clientId, Header.Data calldata header)
    public
    returns (Height.Data[] memory heights)
    {
        bytes memory rlpEthHeader  = header.headers[0].header;
        RLPReader.RLPItem[] memory items = rlpEthHeader.toRlpItem().toList();
        Height.Data memory height = Height.Data({revision_number: 0, revision_height: uint64(items[8].toUint())});
        uint64 timestamp = uint64(items[11].toUint());
        bytes32 stateRoot = bytes32(items[3].toBytes());

        //TODO verify header

        clientStates[clientId].latest_height.revision_number = height.revision_number;
        clientStates[clientId].latest_height.revision_height = height.revision_height;
        consensusStates[clientId][height.toUint128()].timestamp = timestamp;
        consensusStates[clientId][height.toUint128()].state_root = abi.encodePacked(
            verifyStorageProof(address(bytes20(clientStates[clientId].ibc_store_address)), stateRoot, header.account_proof));

        heights = new Height.Data[](1);
        heights[0] = height;
        return heights;
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
            bytes32(consensusState.state_root),
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
            proof, bytes32(consensusState.state_root), keccak256(abi.encodePacked(keccak256(path), COMMITMENT_SLOT))
        );
    }

    // Same as IBFT2Client.sol
    function verifyMembership(bytes calldata proof, bytes32 root, bytes32 slot, bytes32 expectedValue)
    internal
    pure
    returns (bool)
    {
        bytes32 path = keccak256(abi.encodePacked(slot));
        bytes memory dataHash = proof.verifyRLPProof(root, path);
        return expectedValue == bytes32(dataHash.toRlpItem().toUint());
    }

    function verifyNonMembership(bytes calldata proof, bytes32 root, bytes32 slot) internal pure returns (bool) {
        // bytes32 path = keccak256(abi.encodePacked(slot));
        // bytes memory dataHash = proof.verifyRLPProof(root, path); // reverts if proof is invalid
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
        bytes memory accountRLP = accountStateProof.verifyRLPProof(stateRoot, proofPath); // reverts if proof is invalid
        return bytes32(accountRLP.toRlpItem().toList()[ACCOUNT_STORAGE_ROOT_INDEX].toUint());
    }

    modifier onlyIBC() {
        require(msg.sender == ibcHandler);
        _;
    }
}