package module

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"

	// contract address changes for each deployment
	ibcHandlerAddress = "aa43d337145E8930d01cb4E60Abf6595C692921E"
)

type mockChain struct {
	ChainI
}

func (r *mockChain) QueryClientState(height int64) (*clienttypes.QueryClientStateResponse, error) {
	cHeight := clienttypes.NewHeight(0, uint64(height))
	cs := ClientState{
		LatestHeight:    &cHeight,
		IbcStoreAddress: common.Hex2Bytes(ibcHandlerAddress),
	}
	anyClientState, err := codectypes.NewAnyWithValue(&cs)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryClientStateResponse(anyClientState, nil, cHeight), nil
}

type ProverTestSuite struct {
	suite.Suite
	prover *Prover
}

func TestProverTestSuite(t *testing.T) {
	suite.Run(t, new(ProverTestSuite))
}

func (ts *ProverTestSuite) SetupTest() {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		RpcAddr:           "http://localhost:8545",
		EthChainId:        9999,
		HdwMnemonic:       hdwMnemonic,
		HdwPath:           hdwPath,
		IbcHandlerAddress: ibcHandlerAddress,
	})
	ts.Require().NoError(err)
	codec := core.MakeCodec()

	err = chain.Init("", 0, codec, false)
	ts.Require().NoError(err)
	// call SetRelayInfo
	err = chain.SetRelayInfo(&core.PathEnd{
		ClientID:     "mock-client-0",
		ConnectionID: "connection-0",
		ChannelID:    "channel-0",
		PortID:       "transfer",
		Order:        "UNORDERED",
	}, nil, nil)
	ts.Require().NoError(err)

	config := ProverConfig{
		TrustLevelNumerator:   1,
		TrustLevelDenominator: 3,
		TrustingPeriod:        1_000_000_000,
	}
	testChain := mockChain{ChainI: NewChain(chain)}
	ts.prover = NewProver(&testChain, &config).(*Prover)
}

func (ts *ProverTestSuite) TestQueryHeader() {
	header, err := ts.prover.QueryHeader(200)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(200), header.GetHeight().GetRevisionHeight())
}

func (ts *ProverTestSuite) TestQueryLatestHeader() {
	_, err := ts.prover.QueryLatestHeader()
	ts.Require().NoError(err)
}

func (ts *ProverTestSuite) TestCreateMsgCreateClient() {
	header, err := ts.prover.QueryHeader(200)
	ts.Require().NoError(err)
	msg, err := ts.prover.CreateMsgCreateClient("", header, types.AccAddress{})
	ts.Require().NoError(err)
	ts.Require().Equal(msg.ClientState.TypeUrl, "/ibc.lightclients.parlia.v1.ClientState")
	var cs ClientState
	ts.Require().NoError(proto.Unmarshal(msg.ClientState.Value, &cs))
	ts.Require().Equal(cs.ChainId, uint64(9999))
	ts.Require().Equal(cs.TrustingPeriod, uint64(1_000_000_000))
	ts.Require().Equal(cs.TrustLevel.Numerator, uint64(1))
	ts.Require().Equal(cs.TrustLevel.Denominator, uint64(3))
	ts.Require().False(cs.Frozen)
	ts.Require().Equal(common.Bytes2Hex(cs.IbcStoreAddress), ibcHandlerAddress)
	ts.Require().Equal(cs.GetLatestHeight().GetRevisionHeight(), uint64(200))
	ts.Require().Equal(cs.GetLatestHeight().GetRevisionNumber(), ts.prover.revisionNumber)

	var cs2 ConsensusState
	ts.Require().NoError(err)
	ts.Require().NoError(proto.Unmarshal(msg.ConsensusState.Value, &cs2))
	rawHeader := header.(HeaderI)
	validatorSet, err := rawHeader.ValidatorSet()
	account, err := rawHeader.Account(common.HexToAddress(ibcHandlerAddress))
	ts.Require().Equal(cs2.ValidatorSet, validatorSet)
	ts.Require().Equal(cs2.Timestamp, rawHeader.Target().Time)
	ts.Require().Equal(common.BytesToHash(cs2.StateRoot), account.Root)
}

type dstMock struct {
	ChainI
	core.ProverI
}

func (r *dstMock) GetLatestLightHeight() (int64, error) {
	return 10, nil
}

func (ts *ProverTestSuite) TestSetupHeader() {
	dst := dstMock{
		ChainI:  ts.prover.chain,
		ProverI: ts.prover,
	}
	header := &defaultHeader{
		Header: &Header{},
	}
	setupDone, err := ts.prover.SetupHeader(&dst, header)
	done := setupDone.(*defaultHeader)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(10), done.GetTrustedHeight().GetRevisionHeight())
}

func (ts *ProverTestSuite) TestQueryClientStateWithProof() {
	res, err := ts.prover.QueryClientStateWithProof(21400)
	ts.Require().NoError(err)

	ts.Require().Equal(res.ProofHeight.GetRevisionNumber(), ts.prover.revisionNumber)
	ts.Require().Equal(res.ProofHeight.GetRevisionHeight(), uint64(21400))

	// storage_key is 0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae
	ts.Require().Equal(common.Bytes2Hex(res.Proof), "f853f8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080")
}
