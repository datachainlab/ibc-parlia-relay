package module

import (
	"context"
	"fmt"
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
// ex) previous validator : 4, current validator = 21
// latest | target
// 203    | 200
// 204    | 201
// 205    | 202
// 212    | 202
// 213    | 203 ( checkpoint by previous validator )
// 214    | 204
// 215    | 205
func (pr *Prover) GetLatestFinalizedHeader() (out core.Header, err error) {
	latestHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return nil, err
	}
	latestBlockNumber := latestHeight.GetRevisionHeight()
	epochCount := latestBlockNumber / epochBlockPeriod
	currentEpoch, err := pr.chain.Header(context.TODO(), epochCount*epochBlockPeriod)
	if err != nil {
		return nil, err
	}
	countToFinalizeCurrent := uint64(requiredCountToFinalize(currentEpoch))

	// genesis epoch
	if epochCount == 0 {
		countToFinalizeCurrentExceptTarget := countToFinalizeCurrent - 1
		if latestBlockNumber >= countToFinalizeCurrentExceptTarget {
			targetHeight := latestBlockNumber - countToFinalizeCurrentExceptTarget
			return pr.queryHeader(int64(targetHeight))
		}
		return nil, fmt.Errorf("no finalized header found : latest = %d", latestBlockNumber)
	}

	previousEpochHeight := (epochCount - 1) * epochBlockPeriod
	previousEpoch, err := pr.chain.Header(context.TODO(), previousEpochHeight)
	if err != nil {
		return nil, err
	}
	countToFinalizePrevious := uint64(requiredCountToFinalize(previousEpoch))

	// finalized by current validator set
	checkpoint := currentEpoch.Number.Uint64() + countToFinalizePrevious
	target := latestBlockNumber - (countToFinalizeCurrent - 1)
	if target >= checkpoint {
		return pr.queryHeader(int64(target))
	}

	// finalized by previous validator set
	return pr.queryHeader(math.MinInt64(int64(checkpoint-1), int64(latestBlockNumber-(countToFinalizePrevious-1))))
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
	epochCount := blockNumber / epochBlockPeriod
	previousEpochHeight := math.MaxInt64((epochCount-1)*epochBlockPeriod, 0)
	var previousEpochHeader core.Header
	previousEpochHeader, err = pr.queryHeaderWithoutAccountProof(previousEpochHeight)
	if err != nil {
		return nil, err
	}
	header = previousEpochHeader.(*Header)
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
		Timestamp:    target.Time,
		ValidatorSet: validatorSet,
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
		return nil, err
	}
	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(csRes.ClientState, &cs); err != nil {
		return nil, err
	}

	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if cs.GetLatestHeight().GetRevisionHeight() == header.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}

	// Append insufficient epoch blocks
	savedLatestHeight := cs.GetLatestHeight().GetRevisionHeight()
	firstUnsavedEpoch := (savedLatestHeight/epochBlockPeriod + 1) * epochBlockPeriod
	for epochHeight := firstUnsavedEpoch; epochHeight < header.GetHeight().GetRevisionHeight(); epochHeight += epochBlockPeriod {
		epoch, err := pr.queryHeader(int64(epochHeight))
		if err != nil {
			return nil, fmt.Errorf("SetupHeadersForUpdate failed to get past epochs : saved_latest = %d : %+v", savedLatestHeight, err)
		}
		targetHeaders = append(targetHeaders, epoch)
	}
	if len(targetHeaders) == 0 || targetHeaders[len(targetHeaders)-1].GetHeight() != header.GetHeight() {
		targetHeaders = append(targetHeaders, header)
	}

	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = pr.toHeight(cs.GetLatestHeight())
		} else {
			trustedHeight = pr.toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight
		if pr.config.Debug {
			log.Printf("SetupHeadersForUpdate: target height = %d, trustedHeight = %d\n", h.GetHeight().GetRevisionHeight(), trustedHeight.GetRevisionHeight())
		}
	}

	// debug log
	if pr.config.Debug {
		log.Printf("SetupHeadersForUpdate: last height = %d, \n%s\n", header.GetHeight().GetRevisionHeight(), header.ToPrettyString())
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
	res.ProofHeight = pr.toHeight(ctx.Height())
	res.Proof, err = pr.getStateCommitmentProof(host.ConnectionKey(
		pr.chain.Path().ConnectionID,
	), ctx.Height())
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
	res.Proof, err = pr.getStateCommitmentProof(host.ChannelKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
	), ctx.Height())
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
func (pr *Prover) queryHeader(height int64) (core.Header, error) {
	ethHeaders, err := pr.queryETHHeaders(uint64(height))
	if err != nil {
		return nil, fmt.Errorf("height = %d, %+v", height, err)
	}
	// get RLP-encoded account proof
	rlpAccountProof, err := pr.getAccountProof(height)
	if err != nil {
		return nil, fmt.Errorf("height = %d, %+v", height, err)
	}
	return &Header{
		AccountProof: rlpAccountProof,
		Headers:      ethHeaders,
	}, nil
}

// queryHeader returns the header corresponding to the height
func (pr *Prover) queryHeaderWithoutAccountProof(height int64) (core.Header, error) {
	ethHeaders, err := pr.queryETHHeaders(uint64(height))
	if err != nil {
		return nil, fmt.Errorf("height = %d, %+v", height, err)
	}
	return &Header{
		Headers:      ethHeaders,
		AccountProof: crypto.Keccak256(),
	}, nil
}

// queryETHHeaders returns the header corresponding to the height
func (pr *Prover) queryETHHeaders(height uint64) ([]*ETHHeader, error) {
	epochCount := height / epochBlockPeriod
	if epochCount > 0 {
		previousEpochHeight := (epochCount - 1) * epochBlockPeriod
		previousEpochBlock, err := pr.chain.Header(context.TODO(), previousEpochHeight)
		if err != nil {
			return nil, err
		}
		threshold := requiredCountToFinalize(previousEpochBlock)
		if height%epochBlockPeriod < uint64(threshold) {
			// before checkpoint
			return pr.getETHHeaders(height, threshold)
		}
	}
	// genesis count or after checkpoint
	lastEpochNumber := epochCount * epochBlockPeriod
	currentEpochBlock, err := pr.chain.Header(context.TODO(), lastEpochNumber)
	if err != nil {
		return nil, err
	}
	return pr.getETHHeaders(height, requiredCountToFinalize(currentEpochBlock))
}

func (pr *Prover) getETHHeaders(start uint64, requiredCountToFinalize int) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := 0; i < requiredCountToFinalize; i++ {
		block, err := pr.chain.Header(context.TODO(), uint64(i)+start)
		if err != nil {
			return nil, err
		}
		header, err := newETHHeader(block)
		if err != nil {
			return nil, err
		}
		ethHeaders = append(ethHeaders, header)
	}
	return ethHeaders, nil
}

func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}

func requiredCountToFinalize(header *types.Header) int {
	validators := len(header.Extra[extraVanity:len(header.Extra)-extraSeal]) / validatorBytesLength
	// The checkpoint is [(block - 1) % epochCount == len(previousValidatorCount / 2)]
	// for example when the validator count is 21 the checkpoint is 211, 411, 611 ...
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/parlia.go#L605
	// https://github.com/bnb-chain/bsc/blob/master/consensus/parlia/snapshot.go#L191
	return validators/2 + 1
}
