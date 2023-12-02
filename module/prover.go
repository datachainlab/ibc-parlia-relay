package module

import (
	"context"
	"fmt"
	"github.com/hyperledger-labs/yui-relayer/log"
	"time"

	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
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
		_, finalizedHeader, err = getLatestFinalizedHeader(pr.chain.Header, latestHeight.GetRevisionHeight())
	} else {
		finalizedHeader, err = GetFinalizedHeader(pr.chain.Header, height.GetRevisionHeight(), latestHeight.GetRevisionHeight())
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
	height, finalizedHeader, err := getLatestFinalizedHeader(pr.chain.Header, latestBlockNumber)
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
	queryVerifyingHeader := func(height uint64, limitHeight uint64) (core.Header, error) {
		ethHeaders, err := GetFinalizedHeader(pr.chain.Header, height, limitHeight)
		if err != nil {
			return nil, err
		}
		// No finalized header found
		if ethHeaders == nil {
			return nil, nil
		}
		return pr.withProofAndValidators(height, ethHeaders)
	}
	return setupHeadersForUpdate(queryVerifyingHeader, clientStateLatestHeight, latestFinalizedHeader)
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

	// cons.GetTimestamp() returns not nsec but sec
	lcLastTimestamp := time.Unix(int64(cons.GetTimestamp()), 0)

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
	needsRefresh := elapsedTime > threshold
	if needsRefresh {
		log.GetLogger().Debug("needs refresh", "elapsedTime", elapsedTime, "threshold", threshold)
	}

	return needsRefresh, nil
}

func (pr *Prover) withProofAndValidators(height uint64, ethHeaders []*ETHHeader) (core.Header, error) {

	// get RLP-encoded account proof
	rlpAccountProof, _, err := pr.getAccountProof(int64(height))
	if err != nil {
		return nil, fmt.Errorf("failed to get account proof : height = %d, %+v", height, err)
	}

	header := &Header{
		AccountProof: rlpAccountProof,
		Headers:      ethHeaders,
	}

	// Get validator set for verify headers
	previousEpoch := GetPreviousEpoch(height)
	header.PreviousValidators, err = QueryValidatorSet(pr.chain.Header, previousEpoch)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in previous epoch : number = %d : %+v", previousEpoch, err)
	}
	// Epoch doesn't need to get validator set because it contains validator set.
	if !isEpoch(height) {
		currentEpoch := GetCurrentEpoch(height)
		header.CurrentValidators, err = QueryValidatorSet(pr.chain.Header, currentEpoch)
		if err != nil {
			return nil, fmt.Errorf("ValidatorSet was not found in current epoch : number= %d : %+v", currentEpoch, err)
		}
	}
	return header, nil
}

func getLatestFinalizedHeader(getHeader getHeaderFn, latestBlockNumber uint64) (uint64, []*ETHHeader, error) {
	logger := log.GetLogger()
	for i := latestBlockNumber; i > 0; i-- {
		header, err := getHeader(context.Background(), i)
		if err != nil {
			return 0, nil, err
		}
		vote, err := getVoteAttestationFromHeader(header)
		if err != nil {
			return 0, nil, err
		}
		if vote == nil {
			continue
		}
		probablyFinalized := vote.Data.SourceNumber

		logger.Debug("Try to seek verifying headers to finalize", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)

		headers, err := GetFinalizedHeader(getHeader, probablyFinalized, latestBlockNumber)
		if err != nil {
			return 0, nil, err
		}
		if headers != nil {
			return probablyFinalized, headers, nil
		}
		logger.Debug("Failed to seek verifying headers to finalize. So seek previous finalized header.", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)
	}
	return 0, nil, fmt.Errorf("no finalized header found: %d", latestBlockNumber)
}

func setupHeadersForUpdate(queryVerifyingHeader func(uint64, uint64) (core.Header, error), clientStateLatestHeight exported.Height, latestFinalizedHeader *Header) ([]core.Header, error) {
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}
	// Append insufficient epoch blocks
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()
	firstUnsavedEpoch := (savedLatestHeight/constant.BlocksPerEpoch + 1) * constant.BlocksPerEpoch
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()
	if latestFinalizedHeight > firstUnsavedEpoch {
		for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
			epoch, err := queryVerifyingHeader(epochHeight, epochHeight+constant.BlocksPerEpoch)
			if err != nil {
				return nil, fmt.Errorf("setupHeadersForUpdate failed to get past epochs : height=%d : %+v", epochHeight, err)
			}
			if epoch == nil {
				return nil, fmt.Errorf("setupHeadersForUpdate no finalized header found after epoch: height=%d", epochHeight)
			}
			targetHeaders = append(targetHeaders, epoch)
		}
	}
	targetHeaders = append(targetHeaders, latestFinalizedHeader)

	logger := log.GetLogger()
	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = toHeight(clientStateLatestHeight)
		} else {
			trustedHeight = toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight

		logger.Debug("setupHeadersForUpdate", "target", h.GetHeight(), "trusted", trustedHeight, "headerSize", len(h.(*Header).Headers))
	}
	return targetHeaders, nil
}

func (pr *Prover) buildInitialState(dstHeader core.Header) (exported.ClientState, exported.ConsensusState, error) {
	currentEpoch := GetCurrentEpoch(dstHeader.GetHeight().GetRevisionHeight())
	currentValidators, err := QueryValidatorSet(pr.chain.Header, currentEpoch)
	if err != nil {
		return nil, nil, err
	}

	previousEpoch := GetPreviousEpoch(dstHeader.GetHeight().GetRevisionHeight())
	previousValidators, err := QueryValidatorSet(pr.chain.Header, previousEpoch)
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

	var commitmentsSlot [32]byte
	latestHeight := toHeight(dstHeader.GetHeight())
	clientState := ClientState{
		TrustingPeriod:     pr.config.TrustingPeriod,
		MaxClockDrift:      pr.config.MaxClockDrift,
		ChainId:            chainID,
		LatestHeight:       &latestHeight,
		Frozen:             false,
		IbcStoreAddress:    pr.chain.IBCAddress().Bytes(),
		IbcCommitmentsSlot: commitmentsSlot[:],
	}
	consensusState := ConsensusState{
		Timestamp:              header.Time,
		PreviousValidatorsHash: crypto.Keccak256(previousValidators...),
		CurrentValidatorsHash:  crypto.Keccak256(currentValidators...),
		StateRoot:              stateRoot.Bytes(),
	}
	return &clientState, &consensusState, nil
}

func GetPreviousEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	if epochCount == 0 {
		return 0
	}
	return (epochCount - 1) * constant.BlocksPerEpoch
}

func isEpoch(v uint64) bool {
	return v%constant.BlocksPerEpoch == 0
}

func GetCurrentEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	return epochCount * constant.BlocksPerEpoch
}

func toHeight(height exported.Height) clienttypes.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
}
