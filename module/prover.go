package module

import (
	"context"
	"fmt"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tendermint/libs/math"
	"log"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var _ core.Prover = (*Prover)(nil)

type DebuggableChain struct {
	Chain
}

type Prover struct {
	chain  Chain
	config *ProverConfig
}

func NewProver(chain Chain, config *ProverConfig) core.Prover {
	return &Prover{
		chain:  chain,
		config: config,
	}
}

// Init initializes the chain
func (pr *Prover) Init(homePath string, timeout time.Duration, codec codec.ProtoCodecMarshaler, debug bool) error {
	return nil
}

// SetRelayInfo sets source's path and counterparty's info to the chain
func (pr *Prover) SetRelayInfo(path *core.PathEnd, counterparty *core.ProvableChain, counterpartyPath *core.PathEnd) error {
	return nil
}

// SetupForRelay performs chain-specific setup before starting the relay
func (pr *Prover) SetupForRelay(ctx context.Context) error {
	return nil
}

// GetLatestFinalizedHeader returns the latest finalized header from the chain
func (pr *Prover) GetLatestFinalizedHeader() (out core.Header, err error) {
	latestHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return nil, err
	}
	header, err := pr.getLatestFinalizedHeader(latestHeight.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	if pr.config.Debug {
		log.Printf("GetLatestFinalizedHeader: finalized = %d, latest = %d\n", header.GetHeight(), latestHeight)
	}
	return header, err
}

// getLatestFinalizedHeader returns the latest finalized header from the chain
// ex) previous validator : 4, current validator = 21
// latest | target
// 203    | 200
// 204    | 201
// 205    | 202
// 212    | 202
// 213    | 203 ( checkpoint by previous validator )
// 214    | 204
// 215    | 205
func (pr *Prover) getLatestFinalizedHeader(latestBlockNumber uint64) (out core.Header, err error) {
	epochCount := latestBlockNumber / constant.BlocksPerEpoch
	currentEpoch := epochCount * constant.BlocksPerEpoch
	currentEpochValidators, err := pr.getValidatorSet(currentEpoch)
	if err != nil {
		return nil, err
	}
	countToFinalizeCurrent := requiredHeaderCountToFinalize(len(currentEpochValidators))

	// genesis epoch
	if epochCount == 0 {
		countToFinalizeCurrentExceptTarget := countToFinalizeCurrent - 1
		if latestBlockNumber >= countToFinalizeCurrentExceptTarget {
			target := latestBlockNumber - countToFinalizeCurrentExceptTarget
			return pr.queryVerifyingHeader(int64(target), countToFinalizeCurrent)
		}
		return nil, fmt.Errorf("no finalized header found : latest = %d", latestBlockNumber)
	}

	previousEpoch := (epochCount - 1) * constant.BlocksPerEpoch
	previousEpochValidators, err := pr.getValidatorSet(previousEpoch)
	if err != nil {
		return nil, err
	}
	countToFinalizePrevious := requiredHeaderCountToFinalize(len(previousEpochValidators))

	checkpoint := currentEpoch + countToFinalizePrevious
	target := latestBlockNumber - (countToFinalizeCurrent - 1)
	if target >= checkpoint {
		// finalized by current validator set
		return pr.queryVerifyingHeader(int64(target), countToFinalizeCurrent)
	}
	target = uint64(math.MinInt64(int64(checkpoint-1), int64(latestBlockNumber-(countToFinalizePrevious-1))))
	if target > currentEpoch {
		// across checkpoint.
		heightFromEpoch := target - currentEpoch
		requiredHeaderCount, err := pr.requiredHeaderCountToVerifyBetweenCheckpoint(heightFromEpoch, countToFinalizePrevious, previousEpochValidators, currentEpochValidators)
		if err != nil {
			return nil, err
		}
		if target+(requiredHeaderCount-1) > latestBlockNumber {
			return pr.queryVerifyingHeader(int64(currentEpoch), countToFinalizePrevious)
		}
		return pr.queryVerifyingHeader(int64(target), requiredHeaderCount)
	}
	// finalized by previous validator set
	return pr.queryVerifyingHeader(int64(target), countToFinalizePrevious)
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(_ string, dstHeader core.Header, _ sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	header := dstHeader.(*Header)
	target, err := header.Target()
	if err != nil {
		return nil, err
	}

	// Initial client_state must be previous epoch header because lcp-parlia requires validator set when update_client
	blockNumber := target.Number.Uint64()
	epochCount := blockNumber / constant.BlocksPerEpoch
	previousEpoch := uint64(math.MaxInt64(int64((epochCount-1)*constant.BlocksPerEpoch), 0))
	previousValidators, err := pr.getValidatorSet(previousEpoch)
	if err != nil {
		return nil, err
	}

	// get chain id
	chainID, err := pr.chain.CanonicalChainID(context.TODO())
	if err != nil {
		return nil, err
	}

	// create initial client state
	latestHeight := clienttypes.NewHeight(dstHeader.GetHeight().GetRevisionNumber(), previousEpoch)
	clientState := ClientState{
		TrustLevel: &Fraction{
			Numerator:   pr.config.TrustLevelNumerator,
			Denominator: pr.config.TrustLevelDenominator,
		},
		TrustingPeriod:  pr.config.TrustingPeriod,
		ChainId:         chainID,
		LatestHeight:    &latestHeight,
		Frozen:          false,
		IbcStoreAddress: pr.chain.IBCAddress().Bytes(),
	}
	anyClientState, err := codectypes.NewAnyWithValue(&clientState)
	if err != nil {
		return nil, err
	}
	consensusState := ConsensusState{
		Timestamp:      target.Time,
		ValidatorsHash: crypto.Keccak256(previousValidators...),
		// Since ibc handler may not be deployed at the target epoch when create_client is used, state_root is not obtained.
		StateRoot: crypto.Keccak256(),
	}
	anyConsensusState, err := codectypes.NewAnyWithValue(&consensusState)
	if err != nil {
		return nil, err
	}

	return &clienttypes.MsgCreateClient{
		ClientState:    anyClientState,
		ConsensusState: anyConsensusState,
		Signer:         "",
	}, nil
}

// SetupHeadersForUpdate creates a new header based on a given header
func (pr *Prover) SetupHeadersForUpdate(dstChain core.ChainInfoICS02Querier, latestFinalizedHeader core.Header) ([]core.Header, error) {
	header := latestFinalizedHeader.(*Header)
	// LCP doesn't need height / EVM needs latest height
	latestHeightOnDstChain, err := dstChain.LatestHeight()
	if err != nil {
		return nil, err
	}
	csRes, err := dstChain.QueryClientState(core.NewQueryContext(context.TODO(), latestHeightOnDstChain))
	if err != nil {
		return nil, fmt.Errorf("no client state found : SetupHeadersForUpdate: height = %d, %+v", latestHeightOnDstChain.GetRevisionHeight(), err)
	}
	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(csRes.ClientState, &cs); err != nil {
		return nil, err
	}
	return pr.setupHeadersForUpdate(cs.GetLatestHeight(), header)
}

func (pr *Prover) setupHeadersForUpdate(clientStateLatestHeight exported.Height, latestFinalizedHeader *Header) ([]core.Header, error) {
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}

	// Append insufficient epoch blocks
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()
	firstUnsavedEpoch := (savedLatestHeight/constant.BlocksPerEpoch + 1) * constant.BlocksPerEpoch
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()
	previousValidatorSet, err := pr.getValidatorSet(firstUnsavedEpoch)
	if err != nil {
		return nil, fmt.Errorf("SetupHeadersForUpdate failed to get previous validator set : firstUnsavedEpoch = %d : %+v", firstUnsavedEpoch, err)
	}
	for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
		epoch, err := pr.queryVerifyingHeader(int64(epochHeight), requiredHeaderCountToFinalize(len(previousValidatorSet)))
		if err != nil {
			return nil, fmt.Errorf("SetupHeadersForUpdate failed to get past epochs : saved_latest = %d : %+v", savedLatestHeight, err)
		}
		previousValidatorSet = epoch.(*Header).CurrentValidators
		targetHeaders = append(targetHeaders, epoch)
	}
	if len(targetHeaders) == 0 || targetHeaders[len(targetHeaders)-1].GetHeight() != latestFinalizedHeader.GetHeight() {
		targetHeaders = append(targetHeaders, latestFinalizedHeader)
	}

	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = pr.toHeight(clientStateLatestHeight)
		} else {
			trustedHeight = pr.toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight

		if pr.config.Debug {
			log.Printf("SetupHeadersForUpdate: target height = %v, trustedHeight = %v\n", h.GetHeight(), trustedHeight)
		}
	}
	return targetHeaders, nil
}

// QueryClientConsensusStateWithProof returns the ClientConsensusState and its proof
func (pr *Prover) QueryClientConsensusStateWithProof(ctx core.QueryContext, dstClientConsHeight exported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	res, err := pr.chain.QueryClientConsensusState(ctx, dstClientConsHeight)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(host.FullConsensusStateKey(
		pr.chain.Path().ClientID,
		dstClientConsHeight,
	), ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryClientStateWithProof returns the ClientState and its proof
func (pr *Prover) QueryClientStateWithProof(ctx core.QueryContext) (*clienttypes.QueryClientStateResponse, error) {
	res, err := pr.chain.QueryClientState(ctx)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(host.FullClientStateKey(
		pr.chain.Path().ClientID,
	), ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryConnectionWithProof returns the Connection and its proof
func (pr *Prover) QueryConnectionWithProof(ctx core.QueryContext) (*conntypes.QueryConnectionResponse, error) {
	res, err := pr.chain.QueryConnection(ctx)
	if err != nil {
		return nil, err
	}
	if res.Connection.State == conntypes.UNINITIALIZED {
		// connection not found
		return res, nil
	}
	key := host.ConnectionKey(
		pr.chain.Path().ConnectionID,
	)
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(key, ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryChannelWithProof returns the Channel and its proof
func (pr *Prover) QueryChannelWithProof(ctx core.QueryContext) (chanRes *chantypes.QueryChannelResponse, err error) {
	res, err := pr.chain.QueryChannel(ctx)
	if err != nil {
		return nil, err
	}
	if res.Channel.State == chantypes.UNINITIALIZED {
		// channel not found
		return res, nil
	}
	res.ProofHeight = pr.toHeight(ctx.Height())
	key := host.ChannelKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
	)
	res.Proof, err = pr.getStateCommitmentProof(key, ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryPacketCommitmentWithProof returns the packet commitment and its proof
func (pr *Prover) QueryPacketCommitmentWithProof(ctx core.QueryContext, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	res, err := pr.chain.QueryPacketCommitment(ctx, seq)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(host.PacketCommitmentKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
		seq,
	), ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryPacketAcknowledgementCommitmentWithProof returns the packet acknowledgement commitment and its proof
func (pr *Prover) QueryPacketAcknowledgementCommitmentWithProof(ctx core.QueryContext, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	res, err := pr.chain.QueryPacketAcknowledgementCommitment(ctx, seq)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(host.PacketAcknowledgementKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
		seq,
	), ctx.Height())
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pr *Prover) toHeight(height exported.Height) clienttypes.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
}

// queryHeader returns the header corresponding to the height
func (pr *Prover) queryVerifyingHeader(height int64, count uint64) (core.Header, error) {
	uheight := uint64(height)
	ethHeaders, err := pr.getETHHeaders(uheight, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get query : height = %d, %+v", height, err)
	}
	// get RLP-encoded account proof
	rlpAccountProof, _, err := pr.getAccountProof(height)
	if err != nil {
		return nil, fmt.Errorf("failed to get account proof : height = %d, %+v", height, err)
	}

	header := &Header{
		AccountProof: rlpAccountProof,
		Headers:      ethHeaders,
	}

	// Get validator set for verify headers
	epochCount := uheight / constant.BlocksPerEpoch
	var previousEpoch uint64
	if epochCount == 0 {
		previousEpoch = epochCount * constant.BlocksPerEpoch
	} else {
		previousEpoch = (epochCount - 1) * constant.BlocksPerEpoch
	}
	header.PreviousValidators, err = pr.getValidatorSet(previousEpoch)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in previous epoch : number = %d : %+v", previousEpoch, err)
	}
	// Epoch doesn't need to get validator set because it contains validator set.
	if uheight%constant.BlocksPerEpoch != 0 {
		currentEpoch := epochCount * constant.BlocksPerEpoch
		header.CurrentValidators, err = pr.getValidatorSet(currentEpoch)
		if err != nil {
			return nil, fmt.Errorf("ValidatorSet was not found in current epoch : number= %d : %+v", currentEpoch, err)
		}
	}
	return header, nil
}

func (pr *Prover) getETHHeaders(start uint64, count uint64) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := 0; i < int(count); i++ {
		height := uint64(i) + start
		block, err := pr.chain.Header(context.TODO(), height)
		if err != nil {
			return nil, fmt.Errorf("failed to get ETHHeaders : count = %d, height = %d, %+v", count, height, err)
		}
		header, err := newETHHeader(block)
		if err != nil {
			return nil, fmt.Errorf("failed to encode rlp height=%d, %+v", block.Number.Uint64(), err)
		}
		ethHeaders = append(ethHeaders, header)
	}
	return ethHeaders, nil
}

func (pr *Prover) getValidatorSet(epochBlockNumber uint64) ([][]byte, error) {
	header, err := pr.chain.Header(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, err
	}
	return extractValidatorSet(header)
}

func (pr *Prover) requiredHeaderCountToVerifyBetweenCheckpoint(heightFromEpoch uint64, threshold uint64, previousEpochValidators [][]byte, currentEpochValidators [][]byte) (uint64, error) {
	beforeCheckpointCount := threshold - heightFromEpoch
	afterCheckpointCount := heightFromEpoch

	if len(previousEpochValidators) < int(beforeCheckpointCount) {
		return 0, fmt.Errorf("insufficient validator count actual=%d, expected=%d", len(previousEpochValidators), beforeCheckpointCount)
	}
	if len(currentEpochValidators) < int(afterCheckpointCount) {
		return 0, fmt.Errorf("insufficient validator count actual=%d, expected=%d", len(currentEpochValidators), afterCheckpointCount)
	}

	// Get duplicated validators between current epoch and previous epoch.
	validatorsToVerifyBeforeCheckpoint := previousEpochValidators[0:beforeCheckpointCount]
	duplicatedValidatorsCount := 0
	validatorsToVerifyAfterCheckpoint := currentEpochValidators[0:afterCheckpointCount]
	for _, aValidator := range validatorsToVerifyAfterCheckpoint {
		for _, bValidator := range validatorsToVerifyBeforeCheckpoint {
			// same validator is used
			if common.Bytes2Hex(aValidator) == common.Bytes2Hex(bValidator) {
				duplicatedValidatorsCount++
			}
		}
	}
	// Increase the number of header to verify by the amount of duplicates
	increasing := uint64(0)
	restValidatorsAfterCheckpoint := currentEpochValidators[afterCheckpointCount:]
	for _, rValidator := range restValidatorsAfterCheckpoint {
		if duplicatedValidatorsCount == 0 {
			break
		}
		increasing++
		if !contains(rValidator, validatorsToVerifyBeforeCheckpoint) {
			duplicatedValidatorsCount--
		}
	}
	/*
		if pr.config.Debug {
			for i, e := range validatorsToVerifyBeforeCheckpoint {
				log.Printf(" before Val %d: %s\n", i, common.Bytes2Hex(e))
			}
			for i, e := range validatorsToVerifyAfterCheckpoint {
				log.Printf(" after Val %d: %s\n", i, common.Bytes2Hex(e))
			}
			for i, e := range restValidatorsAfterCheckpoint {
				log.Printf(" rest Val %d: %s\n", i, common.Bytes2Hex(e))
			}
		}
	*/
	log.Printf("getHeaderCountToVerifyBetweenCheckpoint heightFromEpoch=%d, duplciated=%d, threshold=%d, increasing=%d", heightFromEpoch, duplicatedValidatorsCount, threshold, increasing)
	return threshold + increasing, nil
}

func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}

func requiredHeaderCountToFinalize(validatorCount int) uint64 {
	// The checkpoint is [(block - 1) % epochCount == len(previousValidatorCount / 2)]
	// for example when the validator count is 21 the checkpoint is 211, 411, 611 ...
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/parlia.go#L605
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/snapshot.go#L191
	return uint64(validatorCount/2 + 1)
}

func contains(target []byte, list [][]byte) bool {
	for _, e := range list {
		if common.Bytes2Hex(target) == common.Bytes2Hex(e) {
			return true
		}
	}
	return false
}
