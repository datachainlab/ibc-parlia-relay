package module

import (
	"context"
	"github.com/datachainlab/ibc-hd-signer/pkg/hd"
	"github.com/hyperledger-labs/yui-relayer/log"
	"math/big"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	types3 "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
)

type mockChain struct {
	Chain
	consensusStateTimestamp map[exported.Height]uint64
	chainTimestamp          map[exported.Height]uint64
	latestHeight            uint64
	trustedHeight           uint64
}

func (c *mockChain) GetProof(_ common.Address, _ [][]byte, _ *big.Int) (*client.StateProof, error) {
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

func (c *mockChain) QueryClientState(ctx core.QueryContext) (*clienttypes.QueryClientStateResponse, error) {
	cHeight := clienttypes.NewHeight(ctx.Height().GetRevisionNumber(), c.trustedHeight)
	cs := ClientState{
		LatestHeight: &cHeight,
	}
	anyClientState, err := codectypes.NewAnyWithValue(&cs)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryClientStateResponse(anyClientState, nil, cHeight), nil
}

func (c *mockChain) QueryClientConsensusState(_ core.QueryContext, height exported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	cHeight := clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
	cs := ConsensusState{
		Timestamp: c.consensusStateTimestamp[cHeight],
	}
	anyConsensusState, err := codectypes.NewAnyWithValue(&cs)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryConsensusStateResponse(anyConsensusState, nil, cHeight), nil
}

func (c *mockChain) Timestamp(height exported.Height) (time.Time, error) {
	return time.Unix(int64(c.chainTimestamp[height]), 0), nil
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
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)

	signerConfig := &hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	}
	anySignerConfig, err := codectypes.NewAnyWithValue(signerConfig)
	ts.Require().NoError(err)
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId: 9999,
		IbcAddress: common.Address{}.String(),
		Signer:     anySignerConfig,
	})
	ts.Require().NoError(err)
	codec := core.MakeCodec()

	err = chain.Init("", 0, codec, false)
	ts.Require().NoError(err)

	err = chain.SetRelayInfo(&core.PathEnd{
		ClientID:     "mock-client-0",
		ConnectionID: "connection-0",
		ChannelID:    "channel-0",
		PortID:       "transfer",
		Order:        "UNORDERED",
	}, nil, nil)
	ts.Require().NoError(err)

	config := ProverConfig{
		TrustingPeriod: 100 * time.Second,
		MaxClockDrift:  1 * time.Millisecond,
		RefreshThresholdRate: &Fraction{
			Numerator:   1,
			Denominator: 2,
		},
	}
	ts.chain = &mockChain{
		Chain:                   NewChain(chain),
		consensusStateTimestamp: make(map[exported.Height]uint64),
		chainTimestamp:          make(map[exported.Height]uint64),
		latestHeight:            0,
		trustedHeight:           0,
	}
	ts.prover = NewProver(ts.chain, &config).(*Prover)
}

func (ts *ProverTestSuite) TestQueryClientStateWithProof() {
	cHeight := clienttypes.NewHeight(0, 21400)
	anyClientState, err := codectypes.NewAnyWithValue(&ClientState{
		LatestHeight:    &cHeight,
		IbcStoreAddress: ts.prover.chain.IBCAddress().Bytes(),
	})
	ts.Require().NoError(err)
	cs := clienttypes.NewQueryClientStateResponse(anyClientState, nil, cHeight)

	bzCs, err := ts.prover.chain.Codec().Marshal(cs)
	ts.Require().NoError(err)

	ctx := core.NewQueryContext(context.TODO(), clienttypes.NewHeight(0, 21400))
	proof, proofHeight, err := ts.prover.ProveState(ctx, host.FullClientStatePath(ts.prover.chain.Path().ClientID), bzCs)
	ts.Require().NoError(err)

	ts.Require().Equal(proofHeight.GetRevisionNumber(), uint64(0))
	ts.Require().Equal(proofHeight.GetRevisionHeight(), uint64(21400))

	decoded := ProveState{}
	ts.Require().NoError(decoded.Unmarshal(proof))

	expected, _ := ts.chain.GetProof(common.Address{}, nil, nil)
	ts.Require().Equal(common.Bytes2Hex(decoded.AccountProof), common.Bytes2Hex(expected.AccountProofRLP))
	// storage_key is 0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae
	ts.Require().Equal(common.Bytes2Hex(decoded.CommitmentProof), "f853f8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080")
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

func (ts *ProverTestSuite) TestCheckRefreshRequired() {
	type dstMock struct {
		Chain
		core.Prover
	}
	dst := dstMock{
		Chain:  ts.prover.chain,
		Prover: ts.prover,
	}
	defer func() {
		ts.chain.latestHeight = 0
		ts.chain.trustedHeight = 0
	}()

	now := time.Now()
	chainHeight := clienttypes.NewHeight(0, 0)
	csHeight := clienttypes.NewHeight(0, 0)
	ts.chain.chainTimestamp[chainHeight] = uint64(now.Unix())

	// should refresh by trusting_period
	ts.chain.consensusStateTimestamp[csHeight] = uint64(now.Add(-51 * time.Second).UnixNano())
	required, err := ts.prover.CheckRefreshRequired(dst)
	ts.Require().NoError(err)
	ts.Require().True(required)

	// needless by trusting_period
	ts.chain.consensusStateTimestamp[csHeight] = uint64(now.Add(-50 * time.Second).UnixNano())
	required, err = ts.prover.CheckRefreshRequired(dst)
	ts.Require().NoError(err)
	ts.Require().False(required)

	// should refresh by block difference
	ts.chain.latestHeight = 2
	ts.prover.config.RefreshBlockDifferenceThreshold = 1
	required, err = ts.prover.CheckRefreshRequired(dst)
	ts.Require().NoError(err)
	ts.Require().True(required)

	// needless by block difference
	ts.prover.config.RefreshBlockDifferenceThreshold = 2
	required, err = ts.prover.CheckRefreshRequired(dst)
	ts.Require().NoError(err)
	ts.Require().False(required)

	// needless by invalid block difference
	ts.chain.latestHeight = 1
	ts.chain.trustedHeight = 3
	ts.prover.config.RefreshBlockDifferenceThreshold = 1
	required, err = ts.prover.CheckRefreshRequired(dst)
	ts.Require().NoError(err)
	ts.Require().False(required)
}

func (ts *ProverTestSuite) TestProveHostConsensusState() {
	cs := ConsensusState{
		StateRoot:              common.Hash{}.Bytes(),
		Timestamp:              1,
		CurrentValidatorsHash:  common.Hash{}.Bytes(),
		PreviousValidatorsHash: common.Hash{}.Bytes(),
	}

	ts.prover.chain.Codec().InterfaceRegistry().RegisterImplementations(
		(*exported.ConsensusState)(nil),
		&ConsensusState{},
	)
	ctx := core.NewQueryContext(context.TODO(), clienttypes.NewHeight(0, 0))
	prove, err := ts.prover.ProveHostConsensusState(ctx, clienttypes.NewHeight(0, 0), &cs)
	ts.Require().NoError(err)
	ts.Require().Len(prove, 150)
}
