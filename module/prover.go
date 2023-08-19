package module

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cometbft/cometbft/libs/math"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
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
func (pr *Prover) getLatestFinalizedHeader(latestBlockNumber uint64) (out core.Header, err error) {
	currentEpoch := getCurrentEpoch(latestBlockNumber)
	currentEpochValidators, err := pr.queryValidatorSet(currentEpoch)
	if err != nil {
		return nil, err
	}
	countToFinalizeCurrent := pr.requiredHeaderCountToFinalize(len(currentEpochValidators))

	// genesis epoch
	if currentEpoch == 0 {
		countToFinalizeCurrentExceptTarget := countToFinalizeCurrent - 1
		if latestBlockNumber >= countToFinalizeCurrentExceptTarget {
			target := latestBlockNumber - countToFinalizeCurrentExceptTarget
			return pr.queryVerifyingHeader(target, countToFinalizeCurrent)
		}
		return nil, fmt.Errorf("no finalized header found : latest = %d", latestBlockNumber)
	}

	previousEpoch := getPreviousEpoch(latestBlockNumber)
	previousEpochValidators, err := pr.queryValidatorSet(previousEpoch)
	if err != nil {
		return nil, err
	}
	// ex) checkpoint is 211 (if validator count is 21)
	checkpoint := currentEpoch + checkpointHeight(len(previousEpochValidators))
	target := latestBlockNumber - (countToFinalizeCurrent - 1)
	if target >= checkpoint {
		// finalized by current validator set. ex) target = 211, 212, 213 ...
		return pr.queryVerifyingHeader(target, countToFinalizeCurrent)
	}

	countToFinalizePrevious := pr.requiredHeaderCountToFinalize(len(previousEpochValidators))
	// ex) previous = 3, current = 11
	if countToFinalizePrevious < countToFinalizeCurrent {
		// ex) latest = 212, checkpoint = 203, previous = 3 -> target is min(212 - 2 = 210 , 203 - 1 = 202)
		// ex) latest = 204, checkpoint = 203, previous = 3 -> target is min(204 - 2 = 202 , 203 - 1 = 202)
		// ex) latest = 200, checkpoint = 203, previous = 3 -> target is min(200 - 2 = 198 , 203 - 1 = 202)
		target = uint64(math.MinInt64(int64(checkpoint-1), int64(latestBlockNumber-(countToFinalizePrevious-1))))
		return pr.queryVerifyingHeader(target, countToFinalizePrevious)
	}
	return pr.queryVerifyingHeaderReverse(countToFinalizePrevious, latestBlockNumber)
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(_ string, dstHeader core.Header, _ sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	// Initial client_state must be previous epoch header because lcp-parlia requires validator set when update_client
	previousEpoch := getPreviousEpoch(dstHeader.GetHeight().GetRevisionHeight())
	previousEpochHeader, err := pr.chain.Header(context.TODO(), previousEpoch)
	if err != nil {
		return nil, err
	}
	previousValidators, err := extractValidatorSet(previousEpochHeader)
	if err != nil {
		return nil, err
	}

	// get chain id
	chainID, err := pr.chain.CanonicalChainID(context.TODO())
	if err != nil {
		return nil, err
	}

	var commitmentsSlot [32]byte
	// create initial client state
	latestHeight := clienttypes.NewHeight(dstHeader.GetHeight().GetRevisionNumber(), previousEpoch)
	clientState := ClientState{
		TrustLevel: &Fraction{
			Numerator:   pr.config.TrustLevelNumerator,
			Denominator: pr.config.TrustLevelDenominator,
		},
		TrustingPeriod:     pr.config.TrustingPeriod,
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
		Timestamp:      previousEpochHeader.Time,
		ValidatorsHash: crypto.Keccak256(previousValidators...),
		// Since ibc handler may not be deployed at the target epoch when create_client is used, state_root is not obtained.
		StateRoot: pr.getStateRootOrEmpty(previousEpochHeader).Bytes(),
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
	if latestFinalizedHeight > firstUnsavedEpoch {
		previousValidatorSet, err := pr.queryValidatorSet(firstUnsavedEpoch)
		if err != nil {
			return nil, fmt.Errorf("SetupHeadersForUpdate failed to get previous validator set : firstUnsavedEpoch = %d : %+v", firstUnsavedEpoch, err)
		}
		for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
			epoch, err := pr.queryVerifyingHeader(epochHeight, pr.requiredHeaderCountToFinalize(len(previousValidatorSet)))
			if err != nil {
				return nil, fmt.Errorf("SetupHeadersForUpdate failed to get past epochs : height=%d : %+v", epochHeight, err)
			}
			unwrap, err := epoch.(*Header).Target()
			if err != nil {
				return nil, fmt.Errorf("SetupHeadersForUpdate failed to unwrap header : height=%d : %+v", epoch.GetHeight(), err)
			}
			previousValidatorSet, err = extractValidatorSet(unwrap)
			if err != nil {
				return nil, fmt.Errorf("SetupHeadersForUpdate failed to extract validator : height=%d : %+v", epoch.GetHeight(), err)
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
			log.Printf("SetupHeadersForUpdate: targetHeight=%v, trustedHeight=%v, headerLength=%d, \n", h.GetHeight(), trustedHeight, len(h.(*Header).Headers))
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
func (pr *Prover) queryVerifyingHeader(height uint64, count uint64) (core.Header, error) {
	ethHeaders, err := pr.queryETHHeaders(height, count)
	if err != nil {
		return nil, fmt.Errorf("failed to get query : height = %d, %+v", height, err)
	}
	return pr.newVerifyingHeader(height, ethHeaders)
}

// queryETHHeaders returns the ETHHeaders
func (pr *Prover) queryETHHeaders(start uint64, count uint64) ([]*ETHHeader, error) {
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

// queryValidatorSet returns the validator set
func (pr *Prover) queryValidatorSet(epochBlockNumber uint64) ([][]byte, error) {
	header, err := pr.chain.Header(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, err
	}
	return extractValidatorSet(header)
}

// queryVerifyingHeaderReverse returns the block count to finalize across checkpoints
func (pr *Prover) queryVerifyingHeaderReverse(mustUniqueCount uint64, start uint64) (core.Header, error) {
	var ethHeaders []*ETHHeader
	coinbase := map[common.Address]struct{}{}
	count := uint64(0)
	height := start
	for {
		block, err := pr.chain.Header(context.TODO(), height)
		if err != nil {
			return nil, fmt.Errorf("failed to get ETHHeaders : count = %d, height = %d, %+v", count, height, err)
		}
		header, err := newETHHeader(block)
		if err != nil {
			return nil, fmt.Errorf("failed to encode rlp height=%d, %+v", block.Number.Uint64(), err)
		}
		ethHeaders = append(ethHeaders, header)
		if _, ok := coinbase[block.Coinbase]; !ok {
			coinbase[block.Coinbase] = struct{}{}
			count++
		}
		if count >= mustUniqueCount {
			break
		}
		height--
	}
	reversed := make([]*ETHHeader, len(ethHeaders))
	for i := 0; i < len(ethHeaders); i++ {
		reversed[i] = ethHeaders[len(ethHeaders)-1-i]
	}
	return pr.newVerifyingHeader(height, reversed)
}

// newVerifyingHeader returns headers to finalize
func (pr *Prover) newVerifyingHeader(height uint64, ethHeaders []*ETHHeader) (core.Header, error) {
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
	previousEpoch := getPreviousEpoch(height)
	header.PreviousValidators, err = pr.queryValidatorSet(previousEpoch)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in previous epoch : number = %d : %+v", previousEpoch, err)
	}
	// Epoch doesn't need to get validator set because it contains validator set.
	if !isEpoch(height) {
		currentEpoch := getCurrentEpoch(height)
		header.CurrentValidators, err = pr.queryValidatorSet(currentEpoch)
		if err != nil {
			return nil, fmt.Errorf("ValidatorSet was not found in current epoch : number= %d : %+v", currentEpoch, err)
		}
	}
	return header, nil
}

func (pr *Prover) requiredHeaderCountToFinalize(validatorCount int) uint64 {
	return ceilDiv(uint64(validatorCount)*pr.config.TrustLevelNumerator, pr.config.TrustLevelDenominator)
}

// newETHHeader returns the new ETHHeader
func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}

// checkpoint return the checkpoint height after epoch
func checkpointHeight(validatorCount int) uint64 {
	// The checkpoint is [(block - 1) % epochCount == len(previousValidatorCount / 2)]
	// for example when the validator count is 21 the checkpoint is 211, 411, 611 ...
	// https://github.com/bnb-chain/bsc/blob/48aaee69e9cb50fc2cedf1398ae4b98b099697db/consensus/parlia/parlia.go#L607
	// https://github.com/bnb-chain/bsc/blob/48aaee69e9cb50fc2cedf1398ae4b98b099697db/consensus/parlia/snapshot.go#L191
	return uint64(validatorCount/2 + 1)
}

func ceilDiv(x uint64, y uint64) uint64 {
	if y == 0 {
		return 0
	}
	return (x + y - 1) / y
}

func getPreviousEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	return uint64(math.MaxInt64(0, int64(epochCount)-1)) * constant.BlocksPerEpoch
}

func isEpoch(v uint64) bool {
	return v%constant.BlocksPerEpoch == 0
}

func getCurrentEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	return epochCount * constant.BlocksPerEpoch
}

func toHeight(height exported.Height) clienttypes.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
}
