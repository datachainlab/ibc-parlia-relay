package module

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/log"

	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var _ core.Prover = (*Prover)(nil)

// keccak256(abi.encode(uint256(keccak256("ibc.commitment")) - 1)) & ~bytes32(uint256(0xff))
var IBCCommitmentsSlot = common.HexToHash("1ee222554989dda120e26ecacf756fe1235cd8d726706b57517715dde4f0c900")

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

// CreateInitialLightClientState returns a pair of ClientState and ConsensusState based on the state of the self chain at `height`.
// These states will be submitted to the counterparty chain as MsgCreateClient.
// If `height` is nil, the latest finalized height is selected automatically.
func (pr *Prover) CreateInitialLightClientState(ctx context.Context, height exported.Height) (exported.ClientState, exported.ConsensusState, error) {
	latestHeight, err := pr.chain.LatestHeight(ctx)
	if err != nil {
		return nil, nil, err
	}
	var finalizedHeader []*ETHHeader
	if height == nil {
		_, finalizedHeader, err = queryLatestFinalizedHeader(ctx, pr.chain.Header, latestHeight.GetRevisionHeight())
	} else {
		finalizedHeader, err = queryFinalizedHeader(ctx, pr.chain.Header, height.GetRevisionHeight(), latestHeight.GetRevisionHeight())
	}
	if err != nil {
		return nil, nil, err
	}
	if len(finalizedHeader) == 0 {
		return nil, nil, fmt.Errorf("no finalized headers were found up to %d", latestHeight.GetRevisionHeight())
	}
	//Header should be Finalized, not necessarily Verifiable.
	return pr.buildInitialState(ctx, &Header{
		Headers: finalizedHeader,
	})
}

// GetLatestFinalizedHeader returns the latest finalized header from the chain
func (pr *Prover) GetLatestFinalizedHeader(ctx context.Context) (out core.Header, err error) {
	latestHeight, err := pr.chain.LatestHeight(ctx)
	if err != nil {
		return nil, err
	}
	header, err := pr.GetLatestFinalizedHeaderByLatestHeight(ctx, latestHeight.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	log.GetLogger().DebugContext(ctx, "GetLatestFinalizedHeader", "finalized", header.GetHeight(), "latest", latestHeight)
	return header, err
}

// GetLatestFinalizedHeaderByLatestHeight returns the latest finalized verifiable header from the chain
func (pr *Prover) GetLatestFinalizedHeaderByLatestHeight(ctx context.Context, latestBlockNumber uint64) (core.Header, error) {
	height, finalizedHeader, err := queryLatestFinalizedHeader(ctx, pr.chain.Header, latestBlockNumber)
	if err != nil {
		return nil, err
	}
	// Make headers verifiable
	return pr.withValidators(ctx, height, finalizedHeader)
}

// SetupHeadersForUpdate creates a new header based on a given header
func (pr *Prover) SetupHeadersForUpdate(ctx context.Context, counterparty core.FinalityAwareChain, latestFinalizedHeader core.Header) (<-chan *core.HeaderOrError, error) {
	header := latestFinalizedHeader.(*Header)
	// LCP doesn't need height / EVM needs latest height
	latestHeightOnDstChain, err := counterparty.LatestHeight(ctx)
	if err != nil {
		return nil, err
	}
	csRes, err := counterparty.QueryClientState(core.NewQueryContext(ctx, latestHeightOnDstChain))
	if err != nil {
		return nil, fmt.Errorf("no client state found : SetupHeadersForUpdate: height = %d, %+v", latestHeightOnDstChain.GetRevisionHeight(), err)
	}
	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(csRes.ClientState, &cs); err != nil {
		return nil, err
	}
	if headers, err := pr.SetupHeadersForUpdateByLatestHeight(ctx, cs.GetLatestHeight(), header); err != nil {
		return nil, err
	} else {
		return core.MakeHeaderStream(headers...), nil
	}
}

func (pr *Prover) SetupHeadersForUpdateByLatestHeight(ctx context.Context, clientStateLatestHeight exported.Height, latestFinalizedHeader *Header) ([]core.Header, error) {
	queryVerifiableNeighboringEpochHeader := func(ctx context.Context, height uint64, limitHeight uint64) (core.Header, error) {
		ethHeaders, err := queryFinalizedHeader(ctx, pr.chain.Header, height, limitHeight)
		if err != nil {
			return nil, err
		}
		// No finalized header found
		if ethHeaders == nil {
			return nil, nil
		}
		return pr.withValidators(ctx, height, ethHeaders)
	}
	latestHeight, err := pr.chain.LatestHeight(ctx)
	if err != nil {
		return nil, err
	}
	return setupHeadersForUpdate(
		ctx,
		queryVerifiableNeighboringEpochHeader,
		pr.chain.Header,
		clientStateLatestHeight,
		latestFinalizedHeader,
		latestHeight,
		GetForkParameters(Network(pr.config.Network)),
	)
}

func (pr *Prover) ProveState(ctx core.QueryContext, path string, value []byte) ([]byte, clienttypes.Height, error) {
	proofHeight := toHeight(ctx.Height())
	accountProof, commitmentProof, err := pr.getStateCommitmentProof(ctx.Context(), []byte(path), proofHeight)
	if err != nil {
		return nil, proofHeight, err
	}
	ret := ProveState{
		AccountProof:    accountProof,
		CommitmentProof: commitmentProof,
	}
	proof, err := ret.Marshal()
	if err != nil {
		return nil, proofHeight, err
	}

	return proof, proofHeight, err
}

func (pr *Prover) CheckRefreshRequired(ctx context.Context, counterparty core.ChainInfoICS02Querier) (bool, error) {
	cpQueryHeight, err := counterparty.LatestHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the counterparty chain: %+v", err)
	}
	cpQueryCtx := core.NewQueryContext(ctx, cpQueryHeight)

	resCs, err := counterparty.QueryClientState(cpQueryCtx)
	if err != nil {
		return false, fmt.Errorf("failed to query the client state on the counterparty chain: %+v", err)
	}

	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(resCs.ClientState, &cs); err != nil {
		return false, fmt.Errorf("failed to unpack Any into tendermint client state: %+v", err)
	}

	resCons, err := counterparty.QueryClientConsensusState(cpQueryCtx, cs.GetLatestHeight())
	if err != nil {
		return false, fmt.Errorf("failed to query the consensus state on the counterparty chain: %+v", err)
	}

	var cons exported.ConsensusState
	if err = pr.chain.Codec().UnpackAny(resCons.ConsensusState, &cons); err != nil {
		return false, fmt.Errorf("failed to unpack Any into tendermint consensus state: %+v", err)
	}
	lcLastTimestamp := time.Unix(0, int64(cons.GetTimestamp()))

	selfQueryHeight, err := pr.chain.LatestHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the self chain: %+v", err)
	}

	selfTimestamp, err := pr.chain.Timestamp(ctx, selfQueryHeight)
	if err != nil {
		return false, fmt.Errorf("failed to get timestamp of the self chain: %+v", err)
	}

	elapsedTime := selfTimestamp.Sub(lcLastTimestamp)

	durationMulByFraction := func(d time.Duration, f *Fraction) time.Duration {
		nsec := d.Nanoseconds() * int64(f.Numerator) / int64(f.Denominator)
		return time.Duration(nsec) * time.Nanosecond
	}
	threshold := durationMulByFraction(pr.config.GetTrustingPeriod(), pr.config.GetRefreshThresholdRate())
	if elapsedTime > threshold {
		log.GetLogger().DebugContext(ctx, "needs refresh", "elapsedTime", elapsedTime, "threshold", threshold)
		return true, nil
	}

	// Check if the block difference exceeds the threshold
	blockDiffThreshold := pr.config.RefreshBlockDifferenceThreshold
	if blockDiffThreshold == 0 || selfQueryHeight.GetRevisionHeight() < cs.GetLatestHeight().GetRevisionHeight() {
		return false, nil
	}
	blockDiff := selfQueryHeight.GetRevisionHeight() - cs.GetLatestHeight().GetRevisionHeight()
	if blockDiff > blockDiffThreshold {
		log.GetLogger().DebugContext(ctx, "needs refresh due to block diff",
			"chain", cpQueryHeight.GetRevisionHeight(),
			"cs", cs.GetLatestHeight().GetRevisionHeight(),
			"threshold", blockDiffThreshold)
		return true, nil
	}
	return false, nil

}

func (pr *Prover) withValidators(ctx context.Context, height uint64, ethHeaders []*ETHHeader) (core.Header, error) {
	return withValidators(ctx, pr.chain.Header, height, ethHeaders, pr.getForkParameters())
}

func (pr *Prover) getForkParameters() []*ForkSpec {
	return GetForkParameters(Network(pr.config.Network))
}

func (pr *Prover) buildInitialState(ctx context.Context, dstHeader core.Header) (exported.ClientState, exported.ConsensusState, error) {

	// Last ForkSpec must have height or CreateClient is less than fork spec timestamp
	forkSpecs := pr.getForkParameters()
	lastForkSpec := forkSpecs[len(forkSpecs)-1]
	lastForkSpecTime, ok := lastForkSpec.GetHeightOrTimestamp().(*ForkSpec_Timestamp)
	if ok && lastForkSpecTime != nil {
		target, err := dstHeader.(*Header).Target()
		if err != nil {
			return nil, nil, err
		}
		if MilliTimestamp(target) >= lastForkSpecTime.Timestamp {
			return nil, nil, fmt.Errorf("target timestamp must be less than the last fork spec timestamp to submit height to ELC. %d, %d ", lastForkSpecTime.Timestamp, MilliTimestamp(target))
		}
	}

	dstHeader, err := pr.withValidators(ctx, dstHeader.GetHeight().GetRevisionHeight(), dstHeader.(*Header).Headers)
	if err != nil {
		return nil, nil, err
	}

	downcast := dstHeader.(*Header)
	header, err := downcast.Target()

	if err != nil {
		return nil, nil, err
	}

	chainID, err := pr.chain.CanonicalChainID(ctx)
	if err != nil {
		return nil, nil, err
	}

	latestHeight := toHeight(dstHeader.GetHeight())
	clientState := ClientState{
		TrustingPeriod:     pr.config.TrustingPeriod,
		MaxClockDrift:      pr.config.MaxClockDrift,
		ChainId:            chainID,
		LatestHeight:       &latestHeight,
		Frozen:             false,
		IbcStoreAddress:    pr.chain.IBCAddress().Bytes(),
		IbcCommitmentsSlot: IBCCommitmentsSlot[:],
		ForkSpecs:          GetForkParameters(Network(pr.config.Network)),
	}
	consensusState := ConsensusState{
		Timestamp:              MilliTimestamp(header),
		PreviousValidatorsHash: makeEpochHash(downcast.PreviousValidators, uint8(downcast.PreviousTurnLength)),
		CurrentValidatorsHash:  makeEpochHash(downcast.CurrentValidators, uint8(downcast.CurrentTurnLength)),
		StateRoot:              header.Root.Bytes(),
	}
	return &clientState, &consensusState, nil
}

func makeEpochHash(validators Validators, turnLength uint8) []byte {
	validatorsHash := crypto.Keccak256(validators...)
	return crypto.Keccak256(append([]byte{turnLength}, validatorsHash...))
}
