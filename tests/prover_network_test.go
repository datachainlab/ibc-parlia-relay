package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/datachainlab/ibc-hd-signer/pkg/hd"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/hyperledger-labs/yui-relayer/log"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
)

type dstChain struct {
	core.Chain
}

func (d dstChain) GetLatestFinalizedHeader(_ context.Context) (latestFinalizedHeader core.Header, err error) {
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
	err := log.InitLogger("DEBUG", "text", "stdout", false)
	ts.Require().NoError(err)

	chain := ts.makeChain("http://localhost:8545", "ibc1")

	ts.chain = chain
	ts.prover = ts.makeProver(ts.chain)
}

func (ts *ProverNetworkTestSuite) TestQueryLatestFinalizedHeader() {
	header, err := ts.prover.GetLatestFinalizedHeader(context.Background())
	ts.Require().NoError(err)
	ts.Require().NoError(header.ValidateBasic())
	ts.Require().Len(header.(*module.Header).Headers, 3)
	h := header.(*module.Header)
	t, err := h.Target()
	ts.Require().NoError(err)
	ts.Require().NotNil(t.RequestsHash) // Check Pargue Enabled
	ts.Require().True(len(h.PreviousValidators) >= 1)
	ts.Require().True(len(h.CurrentValidators) >= 1)
	ts.Require().True(h.CurrentTurnLength >= 1 && h.CurrentTurnLength <= 9)
	ts.Require().True(h.PreviousTurnLength >= 1 && h.PreviousTurnLength <= 9)
}

func (ts *ProverNetworkTestSuite) TestSetupHeadersForUpdate() {
	dst := dstChain{
		Chain: ts.makeChain("http://localhost:8645", "ibc0"),
	}
	header, err := ts.prover.GetLatestFinalizedHeader(context.Background())
	ts.Require().NoError(err)
	setupDone, err := ts.prover.SetupHeadersForUpdate(context.Background(), dst, header)
	ts.Require().NoError(err)
	ts.Require().True(len(setupDone) > 0)
	for _, h := range setupDone {
		ts.Require().Len(h.(*module.Header).Headers, 3)
	}
}

func (ts *ProverNetworkTestSuite) TestSuccessCreateInitialLightClientState() {
	ctx := context.Background()
	s1, s2, err := ts.prover.CreateInitialLightClientState(ctx, nil)
	ts.Require().NoError(err)

	cs := s1.(*module.ClientState)
	ts.Require().Equal(cs.ChainId, uint64(9999))
	ts.Require().Equal(cs.TrustingPeriod, 86400*time.Second)
	ts.Require().Equal(cs.MaxClockDrift, 1*time.Second)
	ts.Require().False(cs.Frozen)
	ts.Require().Equal(common.Bytes2Hex(cs.IbcStoreAddress), strings.ToLower(ts.chain.IBCAddress().String()[2:]))
	ts.Require().Equal(common.Bytes2Hex(cs.IbcCommitmentsSlot), common.Bytes2Hex(module.IBCCommitmentsSlot[:]))

	header, err := ts.chain.Header(ctx, cs.GetLatestHeight().GetRevisionHeight())
	ts.Require().NoError(err)
	ts.Require().Equal(cs.GetLatestHeight().GetRevisionHeight(), header.Number.Uint64())

	forkParams := module.GetForkParameters(module.Localnet)
	currentForkSpec, prevForkSpec, err := module.FindTargetForkSpec(forkParams, header.Number.Uint64(), module.MilliTimestamp(header))
	ts.Require().NoError(err)
	bs, err := module.GetBoundaryHeight(ctx, ts.chain.Header, header.Number.Uint64(), *currentForkSpec)
	ts.Require().NoError(err)
	be, err := bs.GetBoundaryEpochs(prevForkSpec)
	ts.Require().NoError(err)

	currentEpoch := be.CurrentEpochBlockNumber(header.Number.Uint64())
	cVal, cTurn, err := module.QueryValidatorSetAndTurnLength(ctx, ts.chain.Header, currentEpoch)
	ts.Require().NoError(err)
	pVal, pTurn, err := module.QueryValidatorSetAndTurnLength(ctx, ts.chain.Header, be.PreviousEpochBlockNumber(currentEpoch))
	ts.Require().NoError(err)
	consState := s2.(*module.ConsensusState)
	ts.Require().Equal(consState.CurrentValidatorsHash, module.MakeEpochHash(cVal, cTurn))
	ts.Require().Equal(consState.PreviousValidatorsHash, module.MakeEpochHash(pVal, pTurn))
	ts.Require().Equal(consState.Timestamp, module.MilliTimestamp(header))
	ts.Require().Equal(common.BytesToHash(consState.StateRoot), header.Root)
}

func (ts *ProverNetworkTestSuite) makeChain(rpcAddr string, ibcChainID string) module.Chain {
	signerConfig := &hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	}
	anySignerConfig, err := codectypes.NewAnyWithValue(signerConfig)
	ts.Require().NoError(err)
	chain, err := ethereum.NewChain(context.Background(), ethereum.ChainConfig{
		EthChainId: 9999,
		IbcAddress: "0xaa43d337145E8930d01cb4E60Abf6595C692921E",
		Signer:     anySignerConfig,
		RpcAddr:    rpcAddr,
	})
	ts.Require().NoError(err)
	codec := core.MakeCodec()
	modules := []config.ModuleI{ethereum.Module{}, module.Module{}, hd.Module{}}
	for _, m := range modules {
		m.RegisterInterfaces(codec.InterfaceRegistry())
	}
	err = chain.Init("", 0, codec, false)
	ts.Require().NoError(err)
	err = chain.SetRelayInfo(&core.PathEnd{
		ChainID:      ibcChainID,
		ClientID:     "xx-parlia-0",
		ConnectionID: "connection-0",
		ChannelID:    "channel-0",
		PortID:       "transfer",
		Order:        "UNORDERED",
	}, nil, nil)
	ts.Require().NoError(err)
	return module.NewChain(chain, chain.Config().IBCAddress(), chain.Client())
}

func (ts *ProverNetworkTestSuite) makeProver(chain module.Chain) *module.Prover {
	config := module.ProverConfig{
		TrustingPeriod: 86400 * time.Second,
		MaxClockDrift:  1 * time.Second,
		RefreshThresholdRate: &module.Fraction{
			Numerator:   3,
			Denominator: 2,
		},
		Network: string(module.Localnet),
	}
	ts.Require().NoError(config.Validate())
	return module.NewProver(chain, &config).(*module.Prover)
}
