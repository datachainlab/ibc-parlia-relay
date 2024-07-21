package it

import (
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum/signers/hd"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/hyperledger-labs/yui-relayer/log"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
)

type dstChain struct {
	core.Chain
}

func (d dstChain) GetLatestFinalizedHeader() (latestFinalizedHeader core.Header, err error) {
	panic("implement me")
}

type ProverNetworkTestSuite struct {
	suite.Suite
	prover *module.Prover
	chain  module.Chain
}

func TestProverNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(ProverNetworkTestSuite))
}

func (ts *ProverNetworkTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)

	chain := ts.makeChain("http://localhost8545")

	err = chain.SetRelayInfo(&core.PathEnd{
		ClientID:     "xx-parlia-1",
		ConnectionID: "connection-0",
		ChannelID:    "channel-0",
		PortID:       "transfer",
		Order:        "UNORDERED",
	}, nil, nil)
	ts.Require().NoError(err)

	ts.chain = chain
	ts.prover = ts.makeProver(ts.chain)
}

func (ts *ProverNetworkTestSuite) TestQueryLatestFinalizedHeader() {
	header, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().NoError(err)
	ts.Require().NoError(header.ValidateBasic())
	ts.Require().Len(header.(*module.Header).Headers, 3)
	h := header.(*module.Header)
	_, err = h.Target()
	ts.Require().NoError(err)
	ts.Require().True(len(h.PreviousValidators) >= 1)
	ts.Require().True(len(h.CurrentValidators) >= 1)
	ts.Require().True(len(h.AccountProof) >= 1)
	ts.Require().True(h.CurrentTurnTerm >= 1 && h.CurrentTurnTerm <= 9)
	ts.Require().True(h.PreviousTurnTerm >= 1 && h.PreviousTurnTerm <= 9)
}

func (ts *ProverNetworkTestSuite) TestSetupHeadersForUpdate() {
	dst := dstChain{
		Chain: ts.makeChain("http://localhost:8645"),
	}
	header, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().NoError(err)
	setupDone, err := ts.prover.SetupHeadersForUpdate(dst, header)
	ts.Require().NoError(err)
	ts.Require().True(len(setupDone) > 0)
	for _, h := range setupDone {
		ts.Require().Len(h.(*module.Header).Headers, 3)
	}
}

func (ts *ProverNetworkTestSuite) TestSuccessCreateInitialLightClientState() {
	header, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().NoError(err)
	target, err := header.(*module.Header).Target()
	ts.Require().NoError(err)
	s1, s2, err := ts.prover.CreateInitialLightClientState(header.GetHeight())
	ts.Require().NoError(err)

	cs := s1.(*module.ClientState)
	ts.Require().Equal(cs.ChainId, uint64(9999))
	ts.Require().Equal(cs.TrustingPeriod, 86400*time.Second)
	ts.Require().Equal(cs.MaxClockDrift, 1*time.Second)
	ts.Require().False(cs.Frozen)
	ts.Require().Equal(common.Bytes2Hex(cs.IbcStoreAddress), ts.chain.IBCAddress())
	var commitment [32]byte
	ts.Require().Equal(common.Bytes2Hex(cs.IbcCommitmentsSlot), common.Bytes2Hex(commitment[:]))
	ts.Require().Equal(cs.GetLatestHeight(), header.GetHeight())

	cVal, cTurn, err := module.QueryValidatorSetAndTurnTerm(ts.chain.Header, module.GetCurrentEpoch(target.Number.Uint64()))
	ts.Require().NoError(err)
	pVal, pTurn, err := module.QueryValidatorSetAndTurnTerm(ts.chain.Header, module.GetPreviousEpoch(target.Number.Uint64()))
	ts.Require().NoError(err)
	consState := s2.(*module.ConsensusState)
	ts.Require().Equal(consState.CurrentValidatorsHash, module.MakeEpochHash(cVal, cTurn))
	ts.Require().Equal(consState.PreviousValidatorsHash, module.MakeEpochHash(pVal, pTurn))
	ts.Require().Equal(consState.Timestamp, target.Time)
	ts.Require().Equal(common.BytesToHash(consState.StateRoot), target.Root)
}

func (ts *ProverNetworkTestSuite) TestErrorCreateInitialLightClientState() {
	// No finalized header found
	_, _, err := ts.prover.CreateInitialLightClientState(clienttypes.NewHeight(0, 0))
	ts.Require().Equal(err.Error(), "no finalized headers were found up to 0")
}

func (ts *ProverNetworkTestSuite) makeChain(rpcAddr string) module.Chain {
	signerConfig := &hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	}
	anySignerConfig, err := codectypes.NewAnyWithValue(signerConfig)
	ts.Require().NoError(err)
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId: 9999,
		IbcAddress: "0x2F5703804E29F4252FA9405B8D357220d11b3bd9",
		Signer:     anySignerConfig,
		RpcAddr:    rpcAddr,
	})
	ts.Require().NoError(err)
	codec := core.MakeCodec()
	err = chain.Init("", 0, codec, false)
	ts.Require().NoError(err)
	return module.NewChain(chain)
}

func (ts *ProverNetworkTestSuite) makeProver(chain module.Chain) *module.Prover {
	config := module.ProverConfig{
		TrustingPeriod: 86400 * time.Second,
		MaxClockDrift:  1 * time.Second,
		RefreshThresholdRate: &module.Fraction{
			Numerator:   3,
			Denominator: 2,
		},
	}
	return module.NewProver(chain, &config).(*module.Prover)
}
