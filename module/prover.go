package module

import (
	"context"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
)

type Prover struct {
	chain          *ethereum.Chain
	config         ProverConfig
	headerReader   HeaderReader
	revisionNumber uint64
}

func NewProver(chain *ethereum.Chain, config ProverConfig) core.ProverI {
	return &Prover{
		chain:          chain,
		config:         config,
		revisionNumber: 0, //TODO upgrade
		headerReader:   NewHeaderReader(chain.Client().BlockByNumber),
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

	ethHeaders, err := pr.headerReader.QueryETHHeaders(height)
	if err != nil {
		return nil, err
	}
	// get RLP-encoded account proof
	rlpAccountProof, err := pr.getAccountProof(height)
	if err != nil {
		return nil, err
	}
	return NewHeaderProxy(pr.revisionNumber, &Header{
		AccountProof: rlpAccountProof,
		Headers:      ethHeaders,
	}), nil
}

// QueryLatestHeader returns the latest header from the chain
func (pr *Prover) QueryLatestHeader() (out core.HeaderI, err error) {
	bn, err := pr.chain.Client().BlockNumber(context.TODO())
	if err != nil {
		return nil, err
	}
	return pr.QueryHeader(int64(bn))
}

// GetLatestLightHeight returns the latest height on the light client
func (pr *Prover) GetLatestLightHeight() (int64, error) {
	//TODO LCPの方をとる
	bn, err := pr.chain.Client().BlockNumber(context.TODO())
	if err != nil {
		return 0, err
	}
	return int64(bn), nil
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(clientID string, dstHeader core.HeaderI, signer sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	// get account proof from header
	header := dstHeader.(HeaderI)

	// recover account data from account proof
	account, err := header.GetAccount(pr.chain.Config().IBCHandlerAddress())
	if err != nil {
		return nil, err
	}

	// get chain id
	chainID, err := pr.chain.Client().ChainID(context.TODO())
	if err != nil {
		return nil, err
	}

	// create initial client state
	target, err := header.GetTarget()
	if err != nil {
		return nil, err
	}
	height := header.GetHeight()
	latestHeight := clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionNumber())
	clientState := ClientState{
		TrustLevel: &Fraction{
			Numerator:   pr.config.TrustLevelNumerator,
			Denominator: pr.config.TrustLevelDenominator,
		},
		TrustingPeriod:  pr.config.TrustingPeriod,
		ChainId:         chainID.Uint64(),
		LatestHeight:    &latestHeight,
		Frozen:          false,
		IbcStoreAddress: pr.chain.Config().IBCHandlerAddress().Bytes(),
	}
	anyClientState, err := codectypes.NewAnyWithValue(&clientState)
	if err != nil {
		return nil, err
	}

	// create initial consensus state
	validatorSet, err := header.GetValidatorSet()
	if err != nil {
		return nil, err
	}
	consensusState := ConsensusState{
		Timestamp:    target.Time,
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
	return nil, nil
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

// QueryClientConsensusState returns the ClientConsensusState and its proof
func (pr *Prover) QueryClientConsensusStateWithProof(height int64, dstClientConsHeight exported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	return nil, nil
}

// QueryClientStateWithProof returns the ClientState and its proof
func (pr *Prover) QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error) {
	return nil, nil
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
