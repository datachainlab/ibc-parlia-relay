package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

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
func (pr *Prover) CreateInitialLightClientState(height exported.Height) (exported.ClientState, exported.ConsensusState, error) {
	latestHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return nil, nil, err
	}
	var finalizedHeader []*ETHHeader
	if height == nil {
		_, finalizedHeader, err = queryLatestFinalizedHeader(pr.chain.Header, latestHeight.GetRevisionHeight())
	} else {
		finalizedHeader, err = queryFinalizedHeader(pr.chain.Header, height.GetRevisionHeight(), latestHeight.GetRevisionHeight())
	}
	if err != nil {
		return nil, nil, err
	}
	if len(finalizedHeader) == 0 {
		return nil, nil, fmt.Errorf("no finalized headers were found up to %d", latestHeight.GetRevisionHeight())
	}
	//Header should be Finalized, not necessarily Verifiable.
	return pr.buildInitialState(&Header{
		Headers: finalizedHeader,
	})
}

// GetLatestFinalizedHeader returns the latest finalized header from the chain
func (pr *Prover) GetLatestFinalizedHeader() (out core.Header, err error) {
	latestHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return nil, err
	}
	header, err := pr.GetLatestFinalizedHeaderByLatestHeight(latestHeight.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	log.GetLogger().Debug("GetLatestFinalizedHeader", "finalized", header.GetHeight(), "latest", latestHeight)
	return header, err
}

// GetLatestFinalizedHeaderByLatestHeight returns the latest finalized verifiable header from the chain
func (pr *Prover) GetLatestFinalizedHeaderByLatestHeight(latestBlockNumber uint64) (core.Header, error) {
	height, finalizedHeader, err := queryLatestFinalizedHeader(pr.chain.Header, latestBlockNumber)
	if err != nil {
		return nil, err
	}
	// Make headers verifiable
	return pr.withProofAndValidators(height, finalizedHeader)
}

// SetupHeadersForUpdate creates a new header based on a given header
func (pr *Prover) SetupHeadersForUpdate(counterparty core.FinalityAwareChain, latestFinalizedHeader core.Header) ([]core.Header, error) {
	header := latestFinalizedHeader.(*Header)
	// LCP doesn't need height / EVM needs latest height
	latestHeightOnDstChain, err := counterparty.LatestHeight()
	if err != nil {
		return nil, err
	}
	csRes, err := counterparty.QueryClientState(core.NewQueryContext(context.TODO(), latestHeightOnDstChain))
	if err != nil {
		return nil, fmt.Errorf("no client state found : SetupHeadersForUpdate: height = %d, %+v", latestHeightOnDstChain.GetRevisionHeight(), err)
	}
	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(csRes.ClientState, &cs); err != nil {
		return nil, err
	}
	return pr.SetupHeadersForUpdateByLatestHeight(cs.GetLatestHeight(), header)
}

func (pr *Prover) SetupHeadersForUpdateByLatestHeight(clientStateLatestHeight exported.Height, latestFinalizedHeader *Header) ([]core.Header, error) {
	queryVerifiableNeighboringEpochHeader := func(height uint64, limitHeight uint64) (core.Header, error) {
		ethHeaders, err := queryFinalizedHeader(pr.chain.Header, height, limitHeight)
		if err != nil {
			return nil, err
		}
		// No finalized header found
		if ethHeaders == nil {
			return nil, nil
		}
		return pr.withProofAndValidators(height, ethHeaders)
	}
	latestHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return nil, err
	}
	return setupHeadersForUpdate(
		queryVerifiableNeighboringEpochHeader,
		pr.chain.Header,
		clientStateLatestHeight,
		latestFinalizedHeader,
		latestHeight)
}

func (pr *Prover) ProveState(ctx core.QueryContext, path string, value []byte) ([]byte, clienttypes.Height, error) {
	proofHeight := toHeight(ctx.Height())
	proof, err := pr.getStateCommitmentProof([]byte(path), proofHeight)
	return proof, proofHeight, err
}

func (pr *Prover) CheckRefreshRequired(counterparty core.ChainInfoICS02Querier) (bool, error) {
	cpQueryHeight, err := counterparty.LatestHeight()
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the counterparty chain: %+v", err)
	}
	cpQueryCtx := core.NewQueryContext(context.TODO(), cpQueryHeight)

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

	selfQueryHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the self chain: %+v", err)
	}

	selfTimestamp, err := pr.chain.Timestamp(selfQueryHeight)
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
		log.GetLogger().Debug("needs refresh", "elapsedTime", elapsedTime, "threshold", threshold)
		return true, nil
	}

	// Check if the block difference exceeds the threshold
	blockDiffThreshold := pr.config.RefreshBlockDifferenceThreshold
	if blockDiffThreshold == 0 || selfQueryHeight.GetRevisionHeight() < cs.GetLatestHeight().GetRevisionHeight() {
		return false, nil
	}
	blockDiff := selfQueryHeight.GetRevisionHeight() - cs.GetLatestHeight().GetRevisionHeight()
	if blockDiff > blockDiffThreshold {
		log.GetLogger().Debug("needs refresh due to block diff",
			"chain", cpQueryHeight.GetRevisionHeight(),
			"cs", cs.GetLatestHeight().GetRevisionHeight(),
			"threshold", blockDiffThreshold)
		return true, nil
	}
	return false, nil

}

func (pr *Prover) withProofAndValidators(height uint64, ethHeaders []*ETHHeader) (core.Header, error) {
	return withProofAndValidators(pr.chain.Header, pr.getAccountProof, height, ethHeaders)
}

func (pr *Prover) buildInitialState(dstHeader core.Header) (exported.ClientState, exported.ConsensusState, error) {
	currentEpoch := getCurrentEpoch(dstHeader.GetHeight().GetRevisionHeight())
	currentValidators, currentTurnLength, err := queryValidatorSetAndTurnLength(pr.chain.Header, currentEpoch)
	if err != nil {
		return nil, nil, err
	}

	previousEpoch := getPreviousEpoch(dstHeader.GetHeight().GetRevisionHeight())
	previousValidators, previousTurnLength, err := queryValidatorSetAndTurnLength(pr.chain.Header, previousEpoch)
	if err != nil {
		return nil, nil, err
	}
	header, err := dstHeader.(*Header).Target()
	if err != nil {
		return nil, nil, err
	}

	stateRoot, err := pr.GetStorageRoot(header)
	if err != nil {
		return nil, nil, err
	}

	chainID, err := pr.chain.CanonicalChainID(context.TODO())
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
	}
	consensusState := ConsensusState{
		Timestamp:              header.Time,
		PreviousValidatorsHash: makeEpochHash(previousValidators, previousTurnLength),
		CurrentValidatorsHash:  makeEpochHash(currentValidators, currentTurnLength),
		StateRoot:              stateRoot.Bytes(),
	}
	return &clientState, &consensusState, nil
}

func makeEpochHash(validators Validators, turnLength uint8) []byte {
	validatorsHash := crypto.Keccak256(validators...)
	return crypto.Keccak256(append([]byte{turnLength}, validatorsHash...))
}
