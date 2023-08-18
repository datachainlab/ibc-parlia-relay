package module

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	types3 "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"

	// contract address changes for each deployment
	ibcHandlerAddress = "aa43d337145e8930d01cb4e60abf6595c692921e"
)

type mockChain struct {
	Chain
	latestHeight uint64
	chainID      uint64
	blockMap     map[uint64]*types2.Header
}

func (r *mockChain) CanonicalChainID(ctx context.Context) (uint64, error) {
	return r.chainID, nil
}

func (r *mockChain) QueryClientState(ctx core.QueryContext) (*clienttypes.QueryClientStateResponse, error) {
	cHeight := clienttypes.NewHeight(ctx.Height().GetRevisionNumber(), ctx.Height().GetRevisionHeight())
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
	if v, ok := r.blockMap[height]; ok {
		return v, nil
	}
	header := &types2.Header{
		Root: common.HexToHash("c84307dfe4ccfec4a851a77755d63228d8e0b9ba3345d1eee37ed729ee16eaa1"),
	}
	header.Number = big.NewInt(int64(height))
	if header.Number.Uint64()%constant.BlocksPerEpoch == 0 {
		if header.Number.Int64() == 0 {
			header.Extra = append(header.Extra, make([]byte, extraVanity)...)
			for i := 1; i <= 4; i++ {
				header.Extra = append(header.Extra, common.Hex2Bytes(fmt.Sprintf("100000000000000000000000000000000000000%d", i))...)
			}
			header.Extra = append(header.Extra, make([]byte, extraSeal)...)
		} else {
			header.Extra = make([]byte, extraVanity)
			for i := 1; i <= 9; i++ {
				header.Extra = append(header.Extra, common.Hex2Bytes(fmt.Sprintf("200000000000000000000000000000000000000%d", i))...)
			}
			for i := 10; i <= 21; i++ {
				header.Extra = append(header.Extra, common.Hex2Bytes(fmt.Sprintf("20000000000000000000000000000000000000%d", i))...)
			}
			header.Extra = append(header.Extra, make([]byte, extraSeal)...)
		}
	} else {
		header.Extra = make([]byte, extraVanity+extraSeal)
		if header.Number.Uint64()/constant.BlocksPerEpoch < 203 {
			header.Coinbase = common.BytesToAddress(common.Hex2Bytes(fmt.Sprintf("100000000000000000000000000000000000000%d", 1+header.Number.Uint64()%4)))
		} else {
			header.Coinbase = common.BytesToAddress(common.Hex2Bytes(fmt.Sprintf("200000000000000000000000000000000000000%d", 1+header.Number.Uint64()%21)))
		}
	}
	return header, nil
}

func (r *mockChain) GetProof(_ common.Address, _ [][]byte, _ *big.Int) (*client.StateProof, error) {
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

func (c *mockChain) LatestHeight() (exported.Height, error) {
	return clienttypes.NewHeight(0, c.latestHeight), nil
}

type ProverTestSuite struct {
	suite.Suite
	prover *Prover
	chain  *mockChain
}

func TestProverTestSuite(t *testing.T) {
	suite.Run(t, new(ProverTestSuite))
}

func (ts *ProverTestSuite) SetupTest() {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  9999,
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  ibcHandlerAddress,
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
		TrustLevelDenominator: 2,
		TrustingPeriod:        100,
		Debug:                 true,
	}
	ts.chain = &mockChain{
		Chain:        NewChain(chain),
		latestHeight: 21400,
		chainID:      9999,
	}
	ts.prover = NewProver(ts.chain, &config).(*Prover)
}

func (ts *ProverTestSuite) TestQueryVerifyingHeader() {
	header, err := ts.prover.queryVerifyingHeader(200, 10)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(200), header.GetHeight().GetRevisionHeight())
}

func (ts *ProverTestSuite) TestQueryLatestFinalizedHeader() {
	currentLatest := ts.chain.latestHeight
	defer func() {
		ts.chain.latestHeight = currentLatest
	}()
	ts.chain.latestHeight = 0
	_, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().Error(err, "no finalized header found : latest = 0")

	ts.chain.latestHeight = 1
	_, err = ts.prover.GetLatestFinalizedHeader()
	ts.Require().Error(err, "no finalized header found : latest = 0")

	firstEpochBlock, _ := ts.chain.Header(context.TODO(), 0)
	firstValidators, _ := extractValidatorSet(firstEpochBlock)
	firstEpochFinalizing := ts.prover.requiredHeaderCountToFinalize(len(firstValidators))

	secondEpochBlock, _ := ts.chain.Header(context.TODO(), 200)
	secondValidators, _ := extractValidatorSet(secondEpochBlock)
	secondEpochFinalizing := ts.prover.requiredHeaderCountToFinalize(len(secondValidators))

	// finalized by previous epoch validators
	println("finalized by previous epoch validators")
	checkpoint := 200 + int(firstEpochFinalizing)
	for i := 2; i < checkpoint; i++ {
		ts.chain.latestHeight = uint64(i)
		header, terr := ts.prover.GetLatestFinalizedHeader()
		ts.Require().NoError(terr)
		height := header.GetHeight().GetRevisionHeight()
		ts.Require().Equal(height, ts.chain.latestHeight-(firstEpochFinalizing-1), "latest = ", i, "target =", int(height))
		downcast := header.(*Header)
		ts.Require().Len(downcast.PreviousValidators, 4, "latest =", i, "target =", int(height))
		if height%constant.BlocksPerEpoch == 0 {
			ts.Require().Nil(downcast.CurrentValidators, "latest =", i, "target =", int(height))
		} else {
			ts.Require().Len(downcast.CurrentValidators, 4, "latest =", i, "target =", int(height))
		}
	}
	println("across checkpoint")
	for i := checkpoint; i < checkpoint+int(secondEpochFinalizing)-1; i++ {
		ts.chain.latestHeight = uint64(i)
		header, terr := ts.prover.GetLatestFinalizedHeader()
		ts.Require().NoError(terr)
		height := header.GetHeight().GetRevisionHeight()
		downcast := header.(*Header)
		// Upper limit is checkpoint - 1 because there are insufficient blocks to finalize by current epoch validators.
		ts.Require().Len(downcast.PreviousValidators, 4, "latest =", i, "target =", int(height))
		ts.Require().Len(downcast.CurrentValidators, 21, "latest =", i, "target =", int(height))
		if i == checkpoint {
			ts.Require().Equal(int(height), i-int(firstEpochFinalizing-1), "latest =", i, "target =", int(height))
		} else {
			ts.Require().Equal(int(height), checkpoint-1, "latest =", i, "target =", int(height))
		}
	}

	// target is greater than current checkpoint
	println("finalized by current epoch validators")
	for i := checkpoint + int(secondEpochFinalizing) - 1; i < 400; i++ {
		ts.chain.latestHeight = uint64(i)
		header, terr := ts.prover.GetLatestFinalizedHeader()
		ts.Require().NoError(terr)
		height := header.GetHeight().GetRevisionHeight()
		ts.Require().Equal(height, ts.chain.latestHeight-(secondEpochFinalizing-1), i)
		downcast := header.(*Header)
		ts.Require().Len(downcast.PreviousValidators, 4, "latest =", i, "target =", int(height))
		ts.Require().Len(downcast.CurrentValidators, 21, "latest =", i, "target =", int(height))
	}

}

func (ts *ProverTestSuite) TestCreateMsgCreateClient() {

	previousEpochETHHeader, tErr := ts.prover.queryETHHeaders(uint64(200), 4)
	ts.Require().NoError(tErr)
	previousEpochHeader := &Header{Headers: previousEpochETHHeader}

	assertFn := func(finalizedHeight int64) {
		finalizedETHHeader, err := ts.prover.queryETHHeaders(uint64(finalizedHeight), 1)
		ts.Require().NoError(err)
		finalizedHeader := &Header{Headers: finalizedETHHeader}
		msg, err := ts.prover.CreateMsgCreateClient("", finalizedHeader, types.AccAddress{})
		ts.Require().NoError(err)
		ts.Require().Equal(msg.ClientState.TypeUrl, "/ibc.lightclients.parlia.v1.ClientState")
		var cs ClientState
		ts.Require().NoError(proto.Unmarshal(msg.ClientState.Value, &cs))
		ts.Require().Equal(cs.ChainId, uint64(9999))
		ts.Require().Equal(cs.TrustingPeriod, uint64(100))
		ts.Require().Equal(cs.TrustLevel.Numerator, uint64(1))
		ts.Require().Equal(cs.TrustLevel.Denominator, uint64(3))
		ts.Require().False(cs.Frozen)
		ts.Require().Equal(common.Bytes2Hex(cs.IbcStoreAddress), ibcHandlerAddress)
		var commitment [32]byte
		ts.Require().Equal(common.Bytes2Hex(cs.IbcCommitmentsSlot), common.Bytes2Hex(commitment[:]))
		ts.Require().Equal(int64(cs.GetLatestHeight().GetRevisionHeight()), int64(200))
		ts.Require().Equal(cs.GetLatestHeight().GetRevisionNumber(), uint64(0))

		// assert same epoch
		var cs2 ConsensusState
		ts.Require().NoError(err)
		ts.Require().NoError(proto.Unmarshal(msg.ConsensusState.Value, &cs2))
		target, err := previousEpochHeader.Target()
		ts.Require().NoError(err)
		validatorSet, err := extractValidatorSet(target)
		ts.Require().NoError(err)
		ts.Require().Equal(cs2.ValidatorsHash, crypto.Keccak256(validatorSet...))
		ts.Require().Equal(cs2.Timestamp, target.Time)
		ts.Require().Equal(cs2.StateRoot, common.HexToHash("0xc3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0").Bytes())
	}
	assertFn(400)
	assertFn(401)
	assertFn(599)
}

func (ts *ProverTestSuite) TestSetupHeader() {
	type dstMock struct {
		Chain
		core.Prover
	}
	dst := dstMock{
		Chain:  ts.prover.chain,
		Prover: ts.prover,
	}

	header, err := ts.prover.queryVerifyingHeader(21800, 11)
	ts.Require().NoError(err)
	setupDone, err := ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 2)
	e := setupDone[0].(*Header)
	ts.Require().Equal(uint64(21600), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21400), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[1].(*Header)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(21800), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21600), e.GetTrustedHeight().GetRevisionHeight())

	header, err = ts.prover.queryVerifyingHeader(21401, 11)
	ts.Require().NoError(err)
	setupDone, err = ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 1)
	e = setupDone[0].(*Header)
	ts.Require().Equal(uint64(21401), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21400), e.GetTrustedHeight().GetRevisionHeight())

	header, err = ts.prover.queryVerifyingHeader(21400, 11)
	ts.Require().NoError(err)
	setupDone, err = ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 0)

	header, err = ts.prover.queryVerifyingHeader(22005, 11)
	ts.Require().NoError(err)
	setupDone, err = ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 4)
	e = setupDone[0].(*Header)
	ts.Require().Equal(uint64(21600), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21400), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[1].(*Header)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(21800), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21600), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[2].(*Header)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(22000), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(21800), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[3].(*Header)
	ts.Require().NoError(err)
	ts.Require().Equal(uint64(22005), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(22000), e.GetTrustedHeight().GetRevisionHeight())

	currentLatest := ts.chain.latestHeight
	defer func() {
		ts.chain.latestHeight = currentLatest
	}()
	ts.chain.latestHeight = e.GetHeight().GetRevisionHeight()

	// for next update client
	header, err = ts.prover.queryVerifyingHeader(22006, 11)
	ts.Require().NoError(err)
	setupDone, err = ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 1)
	e = setupDone[0].(*Header)
	ts.Require().Equal(uint64(22006), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(22005), e.GetTrustedHeight().GetRevisionHeight())

	// relayer had been stopped
	ts.chain.latestHeight = e.GetHeight().GetRevisionHeight()
	header, err = ts.prover.queryVerifyingHeader(22510, 11)
	ts.Require().NoError(err)
	setupDone, err = ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 3)
	e = setupDone[0].(*Header)
	ts.Require().Nil(e.CurrentValidators)
	ts.Require().Equal(uint64(22200), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(22006), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[1].(*Header)
	ts.Require().Nil(e.CurrentValidators)
	ts.Require().Equal(uint64(22400), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(22200), e.GetTrustedHeight().GetRevisionHeight())
	e = setupDone[2].(*Header)
	ts.Require().Equal(uint64(22510), e.GetHeight().GetRevisionHeight())
	ts.Require().Equal(uint64(22400), e.GetTrustedHeight().GetRevisionHeight())

}

func (ts *ProverTestSuite) TestQueryClientStateWithProof() {
	ctx := core.NewQueryContext(context.TODO(), clienttypes.NewHeight(0, 21400))
	cs, err := ts.prover.chain.QueryClientState(ctx)
	ts.Require().NoError(err)

	bzCs, err := ts.prover.chain.Codec().Marshal(cs)
	ts.Require().NoError(err)

	proof, proofHeight, err := ts.prover.ProveState(ctx, host.FullClientStatePath(ts.prover.chain.Path().ClientID), bzCs)
	ts.Require().NoError(err)

	ts.Require().Equal(proofHeight.GetRevisionNumber(), uint64(0))
	ts.Require().Equal(proofHeight.GetRevisionHeight(), uint64(21400))

	// storage_key is 0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae
	ts.Require().Equal(common.Bytes2Hex(proof), "f853f8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080")
}

func (ts *ProverTestSuite) TestRequireCountToFinalize() {
	header := &types2.Header{
		Number: big.NewInt(1),
	}
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*1)
	validators, _ := extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(1))
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*2)
	validators, _ = extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(2))
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*3)
	validators, _ = extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(2))
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*4)
	validators, _ = extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(3))
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*5)
	validators, _ = extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(3))
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLengthBeforeLuban*21)
	validators, _ = extractValidatorSet(header)
	ts.Require().Equal(ts.prover.requiredHeaderCountToFinalize(len(validators)), uint64(11))
}

func (ts *ProverTestSuite) TestConnection() {
	res := &conntypes.QueryConnectionResponse{
		Connection: &conntypes.ConnectionEnd{
			ClientId: "99-parlia-0",
			Versions: []*conntypes.Version{
				{
					Identifier: "1",
					Features:   []string{"ORDER_ORDERED", "ORDER_UNORDERED"},
				},
			},
			State: conntypes.OPEN,
			Counterparty: conntypes.Counterparty{
				ClientId:     "99-parlia-0",
				ConnectionId: "connection-0",
				Prefix: types3.MerklePrefix{
					KeyPrefix: []byte("ibc"),
				},
			},
			DelayPeriod: 0,
		},
		Proof: []byte{249, 2, 108, 249, 1, 177, 160, 243, 2, 132, 113, 118, 63, 160, 241, 161, 149, 174, 195, 18, 210, 53, 140, 244, 55, 106, 61, 135, 92, 126, 3, 174, 227, 145, 76, 246, 158, 163, 237, 128, 160, 127, 209, 245, 74, 140, 45, 22, 54, 65, 152, 69, 181, 239, 59, 177, 124, 160, 102, 90, 184, 251, 217, 5, 60, 213, 213, 82, 239, 90, 170, 6, 2, 160, 41, 212, 235, 101, 41, 88, 83, 242, 202, 249, 194, 236, 70, 87, 205, 86, 210, 185, 20, 24, 165, 108, 78, 217, 227, 185, 171, 69, 147, 24, 214, 229, 160, 145, 96, 113, 245, 236, 179, 190, 225, 105, 241, 251, 65, 3, 235, 190, 98, 50, 95, 13, 58, 158, 126, 255, 126, 200, 182, 162, 184, 82, 48, 67, 136, 128, 160, 175, 124, 86, 245, 185, 249, 125, 146, 23, 9, 218, 185, 15, 109, 124, 33, 250, 59, 89, 96, 116, 82, 243, 65, 10, 193, 8, 40, 144, 139, 38, 64, 160, 224, 191, 86, 228, 105, 21, 42, 129, 130, 172, 228, 96, 248, 83, 25, 223, 99, 214, 201, 190, 202, 139, 42, 196, 142, 81, 92, 44, 50, 172, 251, 42, 160, 67, 76, 154, 154, 112, 58, 176, 167, 174, 126, 79, 134, 194, 208, 154, 245, 161, 106, 236, 125, 64, 136, 202, 72, 61, 70, 170, 12, 109, 132, 68, 213, 160, 170, 218, 158, 181, 234, 137, 42, 205, 212, 206, 113, 31, 185, 40, 158, 248, 185, 203, 175, 103, 31, 6, 150, 105, 26, 169, 115, 42, 94, 238, 154, 22, 160, 209, 8, 0, 140, 126, 171, 172, 12, 93, 82, 67, 64, 234, 3, 152, 165, 245, 137, 166, 131, 218, 2, 177, 29, 84, 166, 186, 8, 42, 245, 54, 145, 160, 183, 120, 101, 29, 90, 126, 76, 66, 215, 15, 21, 193, 218, 17, 65, 15, 9, 145, 242, 3, 203, 163, 150, 91, 77, 134, 86, 62, 207, 117, 71, 143, 160, 90, 98, 164, 183, 88, 177, 161, 231, 114, 25, 237, 70, 112, 69, 253, 90, 125, 202, 100, 255, 155, 200, 174, 225, 111, 199, 221, 194, 180, 124, 109, 50, 160, 187, 51, 102, 98, 64, 251, 30, 166, 130, 29, 10, 59, 50, 19, 246, 48, 184, 197, 144, 98, 5, 83, 71, 101, 160, 145, 11, 13, 122, 129, 16, 210, 128, 160, 67, 199, 95, 200, 128, 34, 48, 39, 12, 122, 115, 104, 117, 172, 182, 198, 69, 116, 151, 124, 143, 65, 129, 117, 79, 249, 190, 133, 168, 70, 52, 10, 128, 248, 113, 128, 160, 14, 147, 143, 255, 173, 177, 239, 236, 164, 203, 229, 21, 75, 174, 164, 236, 137, 188, 190, 203, 85, 8, 192, 11, 104, 183, 162, 207, 238, 101, 38, 67, 128, 128, 128, 160, 123, 103, 52, 98, 145, 109, 110, 134, 48, 20, 137, 241, 181, 253, 251, 6, 99, 206, 99, 49, 92, 213, 63, 76, 18, 22, 72, 175, 130, 0, 232, 129, 128, 128, 128, 128, 128, 128, 128, 160, 86, 19, 50, 171, 19, 198, 195, 87, 230, 246, 175, 58, 22, 123, 6, 172, 13, 14, 227, 136, 240, 15, 9, 31, 226, 129, 35, 102, 111, 56, 184, 89, 128, 128, 128, 248, 67, 160, 32, 63, 196, 45, 223, 108, 27, 91, 178, 24, 206, 36, 225, 76, 64, 175, 158, 14, 177, 39, 165, 215, 96, 80, 211, 125, 115, 105, 226, 252, 74, 71, 161, 160, 204, 217, 3, 153, 193, 37, 239, 8, 122, 231, 131, 230, 12, 221, 239, 188, 60, 214, 40, 96, 20, 190, 116, 18, 211, 1, 38, 31, 98, 207, 103, 255},
	}
	path := host.ConnectionPath("connection-0")
	commitment, err := messageToCommitment(res.Connection)
	ts.Require().NoError(err)
	storageRoot := common.BytesToHash([]byte{82, 151, 170, 160, 133, 205, 75, 144, 49, 43, 13, 172, 81, 2, 52, 123, 17, 51, 253, 55, 100, 124, 234, 205, 131, 149, 248, 211, 22, 210, 2, 68})
	ts.Require().NoError(verifyMembership(storageRoot, res.Proof, path, commitment))
}

func (ts *ProverTestSuite) TestConnectionStateProofAsLCPCommitment() {
	proofInit := common.Hex2Bytes("f8ccb872f8708369626398636f6e6e656374696f6e732f636f6e6e656374696f6e2d30a022ab576a7df38bb4860ffbc65f30d5a66536fb2d8ec3d5d7d4ab9a3ead0e4312900000000000000000000000000000013da0ee0b5f32ae2bff0d82149ea22b02e350fbbe467a514ba80bbadd89007df1d167949c406cd64ce7fce93eb9d7baf1288c41c921521db84153509dc20ccebff5b9b436cf108737c6bdc24782569dff36a376353407cbe19b2a8fbaf045755a52c674612a1274da7363cd84ff9f5272abcd0ae4e7043f3a3b00")
	var rawValueProof [][]byte
	if err := rlp.DecodeBytes(proofInit, &rawValueProof); err != nil {
		panic(err)
	}
	// same validation as LCPCommitment.sol#parseStateCommitmentAndProof
	ts.Require().Len(rawValueProof, 3)
	commitmentBytes := rawValueProof[0]
	signer := common.BytesToAddress(rawValueProof[1])
	signature := rawValueProof[2]
	ts.Require().Len(signature, 65)
	ts.Require().Equal(signer.String(), "0x9C406cD64Ce7fce93Eb9d7bAf1288C41C921521D")

	var rawCommitmentProof [][]byte
	if err := rlp.DecodeBytes(commitmentBytes, &rawCommitmentProof); err != nil {
		panic(err)
	}
	ts.Require().Len(rawCommitmentProof, 5)

	// assert commitment
	commitmentPrefix := rawCommitmentProof[0]
	commitmentPath := rawCommitmentProof[1]
	commitmentValue := common.BytesToHash(rawCommitmentProof[2])
	commitmentHeight := new(big.Int).SetBytes(rawCommitmentProof[3]).Uint64()
	commitmentStateId := common.BytesToHash(rawCommitmentProof[4])
	ts.Require().Equal(string(commitmentPrefix), "ibc")
	ts.Require().Equal(string(commitmentPath), "connections/connection-0")
	ts.Require().Equal(commitmentValue.String(), "0x22ab576a7df38bb4860ffbc65f30d5a66536fb2d8ec3d5d7d4ab9a3ead0e4312")
	ts.Require().Equal(commitmentHeight, uint64(317))
	ts.Require().Equal(commitmentStateId.String(), "0xee0b5f32ae2bff0d82149ea22b02e350fbbe467a514ba80bbadd89007df1d167")
}

func (ts *ProverTestSuite) TestQueryVerifyingHeaderReverse_Unique() {
	blockMap := map[uint64]*types2.Header{}
	validatorCount := 21
	requiredCount := ts.prover.requiredHeaderCountToFinalize(validatorCount)
	checkpoint := 200 + checkpointHeight(validatorCount)
	for i := uint64(200); i <= checkpoint+uint64(validatorCount); i++ {
		blockMap[i] = &types2.Header{
			Number:   big.NewInt(int64(i)),
			Extra:    make([]byte, extraVanity+extraSeal),
			Coinbase: common.BytesToAddress([]byte{byte(i)}),
		}
	}
	ts.chain.blockMap = blockMap
	defer func() {
		ts.chain.blockMap = nil
	}()
	for latest := checkpoint + requiredCount - 1; latest > checkpoint; latest-- {
		result, err := ts.prover.queryVerifyingHeaderReverse(requiredCount, latest)
		ts.Require().NoError(err)
		ts.Require().Equal(len(result.(*Header).Headers), int(requiredCount), fmt.Sprintf("latest=%d", latest))
	}
}

func (ts *ProverTestSuite) TestQueryVerifyingHeaderReverse_AllSame() {
	previousValidator := [][]byte{{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}, {20}, {21}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}
	currentValidator := [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}, {20}, {21}}
	validatorCount := len(previousValidator)
	requiredCount := ts.prover.requiredHeaderCountToFinalize(validatorCount)
	checkpoint := 200 + checkpointHeight(validatorCount)
	blockMap := map[uint64]*types2.Header{}
	for i := checkpoint - uint64(len(previousValidator)); i < checkpoint; i++ {
		blockMap[i] = &types2.Header{
			Number:   big.NewInt(int64(i)),
			Extra:    make([]byte, extraVanity+extraSeal),
			Coinbase: common.BytesToAddress(previousValidator[i-(checkpoint-uint64(len(previousValidator)))]),
		}
	}
	for i := checkpoint; i < checkpoint+uint64(len(currentValidator)); i++ {
		blockMap[i] = &types2.Header{
			Number:   big.NewInt(int64(i)),
			Extra:    make([]byte, extraVanity+extraSeal),
			Coinbase: common.BytesToAddress(currentValidator[i-checkpoint]),
		}
	}
	ts.chain.blockMap = blockMap
	defer func() {
		ts.chain.blockMap = nil
	}()
	// cur={10-1}: prev={10-1} used, {21} unused = 10 + 10 + 1 = 11
	result, err := ts.prover.queryVerifyingHeaderReverse(requiredCount, 220)
	ts.Require().NoError(err)
	ts.Require().Equal(int(requiredCount+10), len(result.(*Header).Headers))

	// cur={2-1}: prev={21}} unused = 10 + 1 = 11
	result, err = ts.prover.queryVerifyingHeaderReverse(requiredCount, 212)
	ts.Require().Equal(int(requiredCount+2), len(result.(*Header).Headers))
	ts.Require().NoError(err)
}

func (ts *ProverTestSuite) TestQueryVerifyingHeaderReverse_HalfUnique() {
	previousValidator := [][]byte{{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}, {20}, {21}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}
	currentValidator := [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {111}, {112}, {113}, {114}, {115}, {116}, {117}, {118}, {119}, {120}, {121}}
	validatorCount := len(previousValidator)
	requiredCount := ts.prover.requiredHeaderCountToFinalize(validatorCount)
	checkpoint := 200 + checkpointHeight(validatorCount)
	blockMap := map[uint64]*types2.Header{}
	for i := checkpoint - uint64(len(previousValidator)); i < checkpoint; i++ {
		blockMap[i] = &types2.Header{
			Number:   big.NewInt(int64(i)),
			Extra:    make([]byte, extraVanity+extraSeal),
			Coinbase: common.BytesToAddress(previousValidator[i-(checkpoint-uint64(len(previousValidator)))]),
		}
	}
	for i := checkpoint; i < checkpoint+uint64(len(currentValidator)); i++ {
		blockMap[i] = &types2.Header{
			Number:   big.NewInt(int64(i)),
			Extra:    make([]byte, extraVanity+extraSeal),
			Coinbase: common.BytesToAddress(currentValidator[i-checkpoint]),
		}
	}
	ts.chain.blockMap = blockMap
	defer func() {
		ts.chain.blockMap = nil
	}()

	// cur={10-1}: prev={10-1} used, {21} unused = 10 + 10 + 1 = 21
	result, err := ts.prover.queryVerifyingHeaderReverse(requiredCount, 220)
	ts.Require().NoError(err)
	ts.Require().Equal(int(requiredCount+10), len(result.(*Header).Headers))

	// cur={2-1}: prev={10-3} unused, {2-1} used, {21, 20} unused = 2 + 7 + 2 + 2 = 13
	result, err = ts.prover.queryVerifyingHeaderReverse(requiredCount, 212)
	ts.Require().Equal(int(requiredCount+2), len(result.(*Header).Headers))
	ts.Require().NoError(err)
}
