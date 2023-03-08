package module

import (
	"context"
	"fmt"
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

var _ core.ProverI = (*Prover)(nil)

type Prover struct {
	chain          ChainI
	config         *ProverConfig
	headerReader   HeaderReader
	revisionNumber uint64
}

func NewProver(chain ChainI, config *ProverConfig) core.ProverI {
	return &Prover{
		chain:          chain,
		config:         config,
		revisionNumber: 1, //TODO upgrade
		headerReader:   NewHeaderReader(chain.Header),
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

// GetChainID returns the chain ID
func (pr *Prover) GetChainID() string {
	return pr.chain.ChainID()
}

// QueryHeader returns the header corresponding to the height
func (pr *Prover) QueryHeader(height int64) (core.HeaderI, error) {

	ethHeaders, err := pr.headerReader.QueryETHHeaders(uint64(height))
	if err != nil {
		return nil, err
	}
	// get RLP-encoded account proof
	rlpAccountProof, err := pr.getAccountProof(height)
	if err != nil {
		return nil, err
	}
	return NewHeader(pr.revisionNumber, &Header{
		AccountProof: rlpAccountProof,
		Headers:      ethHeaders,
	})
}

// QueryLatestHeader returns the latest header from the chain
func (pr *Prover) QueryLatestHeader() (out core.HeaderI, err error) {
	latest, err := pr.chain.LatestHeight(context.TODO())
	if err != nil {
		return nil, err
	}
	epochCount := latest / epochBlockPeriod
	currentEpoch, err := pr.chain.Header(context.TODO(), epochCount*epochBlockPeriod)
	if err != nil {
		return nil, err
	}
	countToFinalizeCurrent := uint64(requiredCountToFinalize(currentEpoch))
	currentCheckpoint := currentEpoch.Number.Uint64() + countToFinalizeCurrent
	if epochCount == 0 {
		if latest >= countToFinalizeCurrent {
			targetHeight := latest - countToFinalizeCurrent
			return pr.QueryHeader(int64(targetHeight))
		}
		return nil, fmt.Errorf("no finalized header found : latest = %d", latest)
	}

	previousEpochHeight := (epochCount - 1) * epochBlockPeriod
	previousEpoch, err := pr.chain.Header(context.TODO(), previousEpochHeight)
	if err != nil {
		return nil, err
	}
	countToFinalizePrevious := uint64(requiredCountToFinalize(previousEpoch))
	heightAfterEpoch := latest % epochBlockPeriod

	// ex
	//  - previous validator count = 21
	//  - current validator count = 41
	//  - current checkpoint = 211 ( 200 + 21/2 + 1 )

	// latest >= 232 ( 11 + 21 ), Finalized by current validator set
	if heightAfterEpoch >= countToFinalizePrevious+countToFinalizeCurrent {
		return pr.QueryHeader(int64(latest - countToFinalizeCurrent))
	}

	// 211 >= latest < 232, Maybe finalized by current validator set.
	if heightAfterEpoch >= countToFinalizePrevious && heightAfterEpoch < countToFinalizePrevious+countToFinalizeCurrent {

		target := latest - countToFinalizeCurrent

		// target >= 211, target is finalized by current validator set
		if target >= currentCheckpoint {
			return pr.QueryHeader(int64(target))
		}

		// target < 211, The block signed by previous validator set is latest.
		// target + 11 <= latest, target is finalized by previous validator set.
		if target+countToFinalizePrevious <= latest {
			return pr.QueryHeader(int64(target))
		}

		// target is insufficient. The latest finalized block is latest - 11

	}

	// latest < 211, finalized by previous epoch count
	return pr.QueryHeader(int64(latest - countToFinalizePrevious))
}

// GetLatestLightHeight returns the latest height on the light client
func (pr *Prover) GetLatestLightHeight() (int64, error) {
	bn, err := pr.chain.LatestLightHeight(context.TODO())
	if err != nil {
		return 0, err
	}
	return int64(bn), nil
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(_ string, dstHeader core.HeaderI, _ sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	// get account proof from header
	header := dstHeader.(HeaderI)

	// recover account data from account proof
	account, err := header.Account(pr.chain.IBCHandlerAddress())
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
		IbcStoreAddress: pr.chain.IBCHandlerAddress().Bytes(),
	}
	anyClientState, err := codectypes.NewAnyWithValue(&clientState)
	if err != nil {
		return nil, err
	}

	// create initial consensus state
	validatorSet, err := header.ValidatorSet()
	if err != nil {
		return nil, err
	}
	consensusState := ConsensusState{
		Timestamp:    header.Target().Time,
		StateRoot:    account.Root.Bytes(),
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

// SetupHeader creates a new header based on a given header
func (pr *Prover) SetupHeader(dst core.LightClientIBCQueryierI, baseSrcHeader core.HeaderI) (core.HeaderI, error) {
	header := baseSrcHeader.(*defaultHeader)

	// get client state on destination chain
	dstHeight, err := dst.GetLatestLightHeight()
	if err != nil {
		return nil, err
	}
	csRes, err := dst.QueryClientState(dstHeight)
	if err != nil {
		return nil, err
	}
	var cs exported.ClientState
	if err = pr.chain.Codec().UnpackAny(csRes.ClientState, &cs); err != nil {
		return nil, err
	}

	// use the latest height of the client state on the destination chain as trusted height
	latestHeight := cs.GetLatestHeight()
	exportedLatestHeight := clienttypes.NewHeight(latestHeight.GetRevisionNumber(), latestHeight.GetRevisionHeight())
	header.TrustedHeight = &exportedLatestHeight
	return header, nil
}

// UpdateLightWithHeader updates a header on the light client and returns the header and height corresponding to the chain
func (pr *Prover) UpdateLightWithHeader() (core.HeaderI, int64, int64, error) {
	header, err := pr.QueryLatestHeader()
	if err != nil {
		return nil, 0, 0, err
	}
	height := int64(header.GetHeight().GetRevisionHeight())
	return header, height, height, nil
}

// QueryClientConsensusStateWithProof returns the ClientConsensusState and its proof
func (pr *Prover) QueryClientConsensusStateWithProof(height int64, dstClientConsHeight exported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	res, err := pr.chain.QueryClientConsensusState(height, dstClientConsHeight)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.FullConsensusStateKey(
		pr.chain.Path().ClientID,
		dstClientConsHeight,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryClientStateWithProof returns the ClientState and its proof
func (pr *Prover) QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error) {
	res, err := pr.chain.QueryClientState(height)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.FullClientStateKey(
		pr.chain.Path().ClientID,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryConnectionWithProof returns the Connection and its proof
func (pr *Prover) QueryConnectionWithProof(height int64) (*conntypes.QueryConnectionResponse, error) {
	res, err := pr.chain.QueryConnection(height)
	if err != nil {
		return nil, err
	}
	if res.Connection.State == conntypes.UNINITIALIZED {
		// connection not found
		return res, nil
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.ConnectionKey(
		pr.chain.Path().ConnectionID,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryChannelWithProof returns the Channel and its proof
func (pr *Prover) QueryChannelWithProof(height int64) (chanRes *chantypes.QueryChannelResponse, err error) {
	res, err := pr.chain.QueryChannel(height)
	if err != nil {
		return nil, err
	}
	if res.Channel.State == chantypes.UNINITIALIZED {
		// channel not found
		return res, nil
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.ChannelKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryPacketCommitmentWithProof returns the packet commitment and its proof
func (pr *Prover) QueryPacketCommitmentWithProof(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	res, err := pr.chain.QueryPacketCommitment(height, seq)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.PacketCommitmentKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
		seq,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// QueryPacketAcknowledgementCommitmentWithProof returns the packet acknowledgement commitment and its proof
func (pr *Prover) QueryPacketAcknowledgementCommitmentWithProof(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	res, err := pr.chain.QueryPacketAcknowledgementCommitment(height, seq)
	if err != nil {
		return nil, err
	}
	res.ProofHeight = pr.toHeight(height)
	res.Proof, err = pr.getStateCommitmentProof(host.PacketAcknowledgementKey(
		pr.chain.Path().PortID,
		pr.chain.Path().ChannelID,
		seq,
	), height)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pr *Prover) toHeight(height int64) clienttypes.Height {
	return clienttypes.NewHeight(pr.revisionNumber, uint64(height))
}
