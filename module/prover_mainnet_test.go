package module

import (
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ProverMainnetTestSuite struct {
	suite.Suite
	prover *Prover
}

func TestProverMainnetTestSuite(t *testing.T) {
	suite.Run(t, new(ProverMainnetTestSuite))
}

func (ts *ProverMainnetTestSuite) SetupTest() {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  56,
		RpcAddr:     "https://bsc-dataseed1.binance.org",
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		// TODO change address after starting mainnet test
		IbcAddress: "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
	})
	ts.Require().NoError(err)

	config := ProverConfig{}
	ts.prover = NewProver(NewChain(chain), &config).(*Prover)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_BeforeEpoch() {
	latest := uint64(2770999)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	_ = ts.assertSufficient(latest, iHeader)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_Epoch() {
	latest := uint64(2771000)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	_ = ts.assertSufficient(latest, iHeader)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_BeforeCheckpoint() {
	latest := uint64(27710010)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	header := ts.assertSufficient(latest, iHeader)

	// assert epoch
	target, _ := header.Target()
	validators, err := extractValidatorSet(target)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 21)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_Checkpoint() {
	latest := uint64(27710011)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	_ = ts.assertSufficient(latest, iHeader)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_AfterCheckpoint() {
	latest := uint64(27710012)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	_ = ts.assertSufficient(latest, iHeader)
}

func (ts *ProverMainnetTestSuite) assertSufficient(latest uint64, iHeader core.Header) *Header {
	requiredBlocksToFinalizeInCurrentMainnet := 11
	header := iHeader.(*Header)
	ts.Require().Len(header.Headers, requiredBlocksToFinalizeInCurrentMainnet)

	target, err := header.Target()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number.Uint64(), latest-uint64(requiredBlocksToFinalizeInCurrentMainnet-1))

	ethHeaders, err := header.decodeEthHeaders()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number, ethHeaders[0].Number)

	for i, eth := range ethHeaders {
		if i > 0 {
			ts.Require().Equal(eth.Number.Uint64()-1, ethHeaders[i-1].Number.Uint64())
		}
	}
	return header
}
