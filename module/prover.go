package module

import (
	"context"
	"fmt"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
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
		log.Printf("GetLatestFinalizedHeader: finalized = %d, latest = %d\n", header.GetHeight().GetRevisionHeight(), latestHeight.GetRevisionHeight())
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
	currentEpoch, err := pr.chain.Header(context.TODO(), epochCount*constant.BlocksPerEpoch)
	if err != nil {
		return nil, err
	}
	currentEpochValidators, err := extractValidatorSet(currentEpoch)
	if err != nil {
		return nil, err
	}
	countToFinalizeCurrent := uint64(requiredCountToFinalize(len(currentEpochValidators)))

	// genesis epoch
	if epochCount == 0 {
		countToFinalizeCurrentExceptTarget := countToFinalizeCurrent - 1
		if latestBlockNumber >= countToFinalizeCurrentExceptTarget {
			targetHeight := latestBlockNumber - countToFinalizeCurrentExceptTarget
			return pr.queryVerifyingHeader(int64(targetHeight))
		}
		return nil, fmt.Errorf("no finalized header found : latest = %d", latestBlockNumber)
	}

	previousEpochHeight := (epochCount - 1) * constant.BlocksPerEpoch
	previousEpoch, err := pr.chain.Header(context.TODO(), previousEpochHeight)
	if err != nil {
		return nil, err
	}
	previousEpochValidators, err := extractValidatorSet(previousEpoch)
	if err != nil {
		return nil, err
	}
	countToFinalizePrevious := uint64(requiredCountToFinalize(len(previousEpochValidators)))

	// finalized by current validator set
	checkpoint := currentEpoch.Number.Uint64() + countToFinalizePrevious
	target := latestBlockNumber - (countToFinalizeCurrent - 1)
	if target >= checkpoint {
		return pr.queryVerifyingHeader(int64(target))
	}

	// finalized by previous validator set
	return pr.queryVerifyingHeader(math.MinInt64(int64(checkpoint-1), int64(latestBlockNumber-(countToFinalizePrevious-1))))
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(_ string, dstHeader core.Header, _ sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	header := dstHeader.(*Header)
	target, err := header.Target()
	if err != nil {
		return nil, err
	}

	// Initial client_state must be previous epoch header because lcp-parlia requires validator set when update_client
	blockNumber := target.Number.Int64()
	epochCount := blockNumber / int64(constant.BlocksPerEpoch)
	previousEpochHeight := math.MaxInt64((epochCount-1)*int64(constant.BlocksPerEpoch), 0)
	ethHeaders, err := pr.queryETHHeaders(uint64(previousEpochHeight))
	if err != nil {
		return nil, err
	}
	header = &Header{Headers: ethHeaders}
	target, err = header.Target()
	if err != nil {
		return nil, err
	}

	// get chain id
	chainID, err := pr.chain.CanonicalChainID(context.TODO())
	if err != nil {
		return nil, err
	}

	// create initial client state
	height := header.GetHeight()
	latestHeight := clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
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

	// create initial consensus state
	validatorSet, err := extractValidatorSet(target)
	if err != nil {
		return nil, err
	}
	consensusState := ConsensusState{
		Timestamp:      target.Time,
		ValidatorsHash: crypto.Keccak256(validatorSet...),
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
	for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
		epoch, err := pr.queryVerifyingHeader(int64(epochHeight))
		if err != nil {
			return nil, fmt.Errorf("SetupHeadersForUpdate failed to get past epochs : saved_latest = %d : %+v", savedLatestHeight, err)
		}
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
func (pr *Prover) queryVerifyingHeader(height int64) (core.Header, error) {
	uheight := uint64(height)
	ethHeaders, err := pr.queryETHHeaders(uheight)
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

// queryETHHeaders returns the header corresponding to the height
func (pr *Prover) queryETHHeaders(height uint64) ([]*ETHHeader, error) {
	epochCount := height / constant.BlocksPerEpoch
	if epochCount > 0 {
		previousEpochHeight := (epochCount - 1) * constant.BlocksPerEpoch
		previousEpochBlock, err := pr.chain.Header(context.TODO(), previousEpochHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to get header : previousEpochHeight = %d %+v", previousEpochHeight, err)
		}
		previousEpochValidators, err := extractValidatorSet(previousEpochBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to get validator from header : previousEpochHeight = %d %+v", previousEpochHeight, err)
		}
		threshold := requiredCountToFinalize(len(previousEpochValidators))
		if height%constant.BlocksPerEpoch < uint64(threshold) {
			// before checkpoint
			return pr.getETHHeaders(height, threshold)
		}
	}
	// genesis count or after checkpoint
	currentEpochHeight := epochCount * constant.BlocksPerEpoch
	currentEpochBlock, err := pr.chain.Header(context.TODO(), currentEpochHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get header : currentEpochBlock = %d %+v", currentEpochHeight, err)
	}
	currentEpochValidators, err := extractValidatorSet(currentEpochBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator from header : currentEpochHeight = %d %+v", currentEpochHeight, err)
	}
	return pr.getETHHeaders(height, requiredCountToFinalize(len(currentEpochValidators)))
}

func (pr *Prover) getETHHeaders(start uint64, requiredCountToFinalize int) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := 0; i < requiredCountToFinalize; i++ {
		height := uint64(i) + start
		block, err := pr.chain.Header(context.TODO(), height)
		if err != nil {
			return nil, fmt.Errorf("failed to get ETHHeaders : requiredCountToFinalize = %d, height = %d, %+v", requiredCountToFinalize, height, err)
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

func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}

func requiredCountToFinalize(validatorCount int) int {
	// The checkpoint is [(block - 1) % epochCount == len(previousValidatorCount / 2)]
	// for example when the validator count is 21 the checkpoint is 211, 411, 611 ...
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/parlia.go#L605
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/snapshot.go#L191
	return validatorCount/2 + 1
}
