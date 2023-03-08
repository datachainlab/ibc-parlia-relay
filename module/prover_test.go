package module

import (
	"context"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/client"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"

	// contract address changes for each deployment
	ibcHandlerAddress = "aa43d337145e8930d01cb4e60abf6595c692921e"
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

func (r *mockChain) Header(_ context.Context, height uint64) (*types2.Header, error) {
	header := &types2.Header{
		Root: common.HexToHash("c84307dfe4ccfec4a851a77755d63228d8e0b9ba3345d1eee37ed729ee16eaa1"),
	}
	header.Number = big.NewInt(int64(height))
	if header.Number.Int64()%epochBlockPeriod == 0 {
		if header.Number.Int64() == 0 {
			header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*4)
		} else {
			header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*21)
		}
	} else {
		header.Extra = make([]byte, extraVanity+extraSeal)
	}
	return header, nil
}

func (r *mockChain) GetStateProof(_ common.Address, _ [][]byte, _ *big.Int) (*client.StateProof, error) {
	// eth.getProof("0xaa43d337145E8930d01cb4E60Abf6595C692921E",["0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae"], 21400)
	// storageHash: "0xc3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0",
	accountProof := []string{
		"0xf901f1a080679a623dfdd0dfa34cb5c1db80292abdc2a9e75f5026e3d24cd10ea58f8e0da04f4d7ef0e932874f07aec064ee1281cd6a3245fceab78bdd6a8d2d7a86d27451a0a715335e2de6e91c28910eff04e8709ff6ca93862121a0b52560071867a9f14080a0842db9556e659d64ca9d2d33229ebac6e7e2185f42bd07965464de8064d94ac8a0d94bd2db341ab9d3d7f4fe0aa569bb21dfac0d5eb0ec008c7af23d7f2ed98ec1a0cee66e2515872d5f4b42ada7cc733288809c11ab99aa0d25eb941236074f9904a0f3aa8d1774f013d8af0fdd8364c7833f16b42ad377e485b754f5cdae6fedaa2fa0bffc6b17aaf862725aaf4af4ecda3ed70d4102b875451eb965259ead260b06c7a026a29f57f5efaf83a8f098ed0ba0f53aac353364ce498a82d589e7bcf1f84e76a01a25f2cac2c6a021225ea182c3c391c0fafac96cb38896eb45648a5c33f31b6ca0d4d6b410f89044b335cc7b14221050035d87d390043bf6d84bc0f8005391f213a092dfa1004df4e71ccfaf3a6d682718f1fbb2d1e6411566e139f1efa74874c303a078455f6ef72aa4dc670e9b467fdbe29d37b5c4eb526ee07b372d2bcea57871eaa05911bcb62e4ba3117ca428f93305ebf06247d573f25bb0fff22681716c21744da0f47e1a054e1ee9ac18fd711b2571c2cab26e88d1a5be46d7078723076866265880",
		"0xf851808080808080808080a08ffa88d75a03fd29af8cb1a4ac016e32ef8e39631a6bf45d79a34adfc4ecb1448080a0a1161a49c0c7e7a92a2efe173abffdbb1ed91e5235688e2edbc4e38078dc5c5580808080",
		"0xf869a02012683435c076b898a6cac1c03e41900e379104fefd4219d99f7908cb59cfb3b846f8440180a0c3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0a07498e14000b8457a51de3cd583e9337cfa52aee2c2e9f945fac35a820e685904",
	}
	accountProofRLP, err := encodeRLP(accountProof)
	if err != nil {
		return nil, err
	}
	storageProof := []string{"0xf8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080"}
	storageProofRLP, err := encodeRLP(storageProof)
	if err != nil {
		return nil, err
	}
	return &client.StateProof{
		AccountProofRLP: accountProofRLP,
		StorageProofRLP: [][]byte{storageProofRLP},
	}, nil
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

func (ts *ProverTestSuite) TestQueryETHHeaders() {
	assertHeader := func(height uint64, count int) {
		ethHeaders, err := ts.prover.queryETHHeaders(height)
		assert := ts.Require()
		assert.NoError(err)
		assert.Len(ethHeaders, count) // only one validator
		var header types2.Header
		for i := 0; i < count; i++ {
			assert.NoError(rlp.DecodeBytes(ethHeaders[i].Header, &header))
			assert.Equal(header.Number.Uint64(), height+uint64(i))
		}
	}
	assertHeader(0, 2)
	assertHeader(1, 2)
	assertHeader(199, 2)
	assertHeader(200, 2)
	assertHeader(201, 2)
	assertHeader(202, 11)
}

func (ts *ProverTestSuite) TestRequireCountToFinalize() {
	header := &types2.Header{}
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*1)
	ts.Require().Equal(requiredCountToFinalize(header), 1)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*2)
	ts.Require().Equal(requiredCountToFinalize(header), 1)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*3)
	ts.Require().Equal(requiredCountToFinalize(header), 2)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*4)
	ts.Require().Equal(requiredCountToFinalize(header), 2)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*21)
	ts.Require().Equal(requiredCountToFinalize(header), 11)
}