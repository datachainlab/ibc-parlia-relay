package module

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cometbft/cometbft/libs/math"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
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
	header, err := pr.GetLatestFinalizedHeaderByLatestHeight(latestHeight.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	if pr.config.Debug {
		log.Printf("GetLatestFinalizedHeader: finalized = %d, latest = %d\n", header.GetHeight(), latestHeight)
	}
	return header, err
}

// GetLatestFinalizedHeaderByLatestHeight returns the latest finalized header from the chain
func (pr *Prover) GetLatestFinalizedHeaderByLatestHeight(latestBlockNumber uint64) (core.Header, error) {
	for i := latestBlockNumber; i > 0; i-- {
		header, err := pr.chain.Header(context.Background(), i)
		if err != nil {
			return nil, err
		}
		vote, err := getVoteAttestationFromHeader(header)
		if err != nil {
			return nil, err
		}
		if vote == nil {
			continue
		}
		probablyFinalized := vote.Data.SourceNumber
		if pr.config.Debug {
			log.Printf("Try to seek verifying headers to finalize %d, latest=%d\n", probablyFinalized, latestBlockNumber)
		}
		headers, err := pr.QueryVerifyingEthHeaders(probablyFinalized, latestBlockNumber)
		if err != nil {
			return nil, err
		}
		if headers != nil {
			return pr.withProofAndValidators(probablyFinalized, headers)
		}
		if pr.config.Debug {
			log.Printf("Failed to seek verifying headers to finalize %d, latest=%d. So seek previous finalized header.\n", probablyFinalized, latestBlockNumber)
		}
	}
	return nil, fmt.Errorf("no finalized header found: %d", latestBlockNumber)
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(_ string, dstHeader core.Header, _ sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	currentEpoch := GetCurrentEpoch(dstHeader.GetHeight().GetRevisionHeight())
	currentValidators, err := pr.QueryValidatorSet(currentEpoch)
	if err != nil {
		return nil, err
	}

	previousEpoch := GetPreviousEpoch(dstHeader.GetHeight().GetRevisionHeight())
	previousValidators, err := pr.QueryValidatorSet(previousEpoch)
	if err != nil {
		return nil, err
	}
	header, err := dstHeader.(*Header).Target()
	if err != nil {
		return nil, err
	}

	stateRoot, err := pr.GetStorageRoot(header)
	if err != nil {
		return nil, err
	}

	chainID, err := pr.chain.CanonicalChainID(context.TODO())
	if err != nil {
		return nil, err
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
	anyClientState, err := codectypes.NewAnyWithValue(&clientState)
	if err != nil {
		return nil, err
	}
	consensusState := ConsensusState{
		Timestamp:              header.Time,
		PreviousValidatorsHash: crypto.Keccak256(previousValidators...),
		CurrentValidatorsHash:  crypto.Keccak256(currentValidators...),
		StateRoot:              stateRoot.Bytes(),
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
	return pr.SetupHeadersForUpdateByLatestHeight(cs.GetLatestHeight(), header)
}

func (pr *Prover) SetupHeadersForUpdateByLatestHeight(clientStateLatestHeight exported.Height, latestFinalizedHeader *Header) ([]core.Header, error) {
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
			epoch, err := pr.queryVerifyingHeader(epochHeight, epochHeight+constant.BlocksPerEpoch)
			if err != nil {
				return nil, fmt.Errorf("SetupHeadersForUpdateByLatestHeight failed to get past epochs : height=%d : %+v", epochHeight, err)
			}
			if epoch == nil {
				return nil, fmt.Errorf("SetupHeadersForUpdateByLatestHeight no finalized header found after epoch: height=%d", epochHeight)
			}
			targetHeaders = append(targetHeaders, epoch)
		}
	}
	targetHeaders = append(targetHeaders, latestFinalizedHeader)

	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = toHeight(clientStateLatestHeight)
		} else {
			trustedHeight = toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight

		if pr.config.Debug {
			log.Printf("SetupHeadersForUpdateByLatestHeight: targetHeight=%v, trustedHeight=%v, headerLength=%d, \n", h.GetHeight(), trustedHeight, len(h.(*Header).Headers))
		}
	}
	return targetHeaders, nil
}

func (pr *Prover) ProveState(ctx core.QueryContext, path string, value []byte) ([]byte, clienttypes.Height, error) {
	proofHeight := toHeight(ctx.Height())
	proof, err := pr.getStateCommitmentProof([]byte(path), proofHeight)
	return proof, proofHeight, err
}

// queryVerifyingHeader returns headers to finalize
func (pr *Prover) queryVerifyingHeader(height uint64, limit uint64) (core.Header, error) {
	ethHeaders, err := pr.QueryVerifyingEthHeaders(height, limit)
	if err != nil {
		return nil, err
	}
	return pr.withProofAndValidators(height, ethHeaders)
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
	header.PreviousValidators, err = pr.QueryValidatorSet(previousEpoch)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in previous epoch : number = %d : %+v", previousEpoch, err)
	}
	// Epoch doesn't need to get validator set because it contains validator set.
	if !isEpoch(height) {
		currentEpoch := GetCurrentEpoch(height)
		header.CurrentValidators, err = pr.QueryValidatorSet(currentEpoch)
		if err != nil {
			return nil, fmt.Errorf("ValidatorSet was not found in current epoch : number= %d : %+v", currentEpoch, err)
		}
	}
	return header, nil
}

func (pr *Prover) QueryVerifyingEthHeaders(height uint64, limit uint64) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := height; i+2 <= limit; i++ {
		targetBlock, targetETHHeader, _, err := pr.queryETHHeader(i)
		if err != nil {
			return nil, err
		}
		childBlock, childETHHeader, childVote, err := pr.queryETHHeader(i + 1)
		if err != nil {
			return nil, err
		}
		_, grandChildETHHeader, grandChildVote, err := pr.queryETHHeader(i + 2)
		if err != nil {
			return nil, err
		}

		if childVote == nil || grandChildVote == nil ||
			grandChildVote.Data.SourceNumber != targetBlock.Number.Uint64() ||
			grandChildVote.Data.SourceNumber != childVote.Data.TargetNumber ||
			grandChildVote.Data.TargetNumber != childBlock.Number.Uint64() {
			// Append to verify header sequence
			ethHeaders = append(ethHeaders, targetETHHeader)
			continue
		}
		return append(ethHeaders, targetETHHeader, childETHHeader, grandChildETHHeader), nil
	}
	if pr.config.Debug {
		log.Printf("Insufficient verifying headers to finalize %d. limit=%d", height, limit)
	}
	return nil, nil
}

func (pr *Prover) queryETHHeader(height uint64) (*types.Header, *ETHHeader, *VoteAttestation, error) {
	block, err := pr.chain.Header(context.TODO(), height)
	if err != nil {
		return nil, nil, nil, err
	}
	ethHeader, err := newETHHeader(block)
	if err != nil {
		return nil, nil, nil, err
	}
	vote, err := getVoteAttestationFromHeader(block)
	if err != nil {
		return nil, nil, nil, err
	}
	return block, ethHeader, vote, err
}

// QueryValidatorSet returns the validator set
func (pr *Prover) QueryValidatorSet(epochBlockNumber uint64) ([][]byte, error) {
	header, err := pr.chain.Header(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, err
	}
	return ExtractValidatorSet(header)
}

// newETHHeader returns the new ETHHeader
func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}

func GetPreviousEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	return uint64(math.MaxInt64(0, int64(epochCount)-1)) * constant.BlocksPerEpoch
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
