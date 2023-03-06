package module

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"

	// contract address changes for each deployment
	ibcHandlerAddress = "b794a294a7e455058789c82f15e8ce00669d689e"
)

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
	config := ProverConfig{
		TrustLevelNumerator:   1,
		TrustLevelDenominator: 3,
		TrustingPeriod:        1_000_000_000,
	}
	ts.prover = NewProver(NewChain(chain), &config).(*Prover)
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
