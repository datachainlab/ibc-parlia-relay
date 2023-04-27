package module

import (
	"github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
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

	config := ProverConfig{
		Debug: true,
	}
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

	clientStateLatestHeight := types.NewHeight(0, iHeader.GetHeight().GetRevisionHeight()-constant.BlocksPerEpoch-1)
	headers, err := ts.prover.setupHeadersForUpdate(clientStateLatestHeight, header)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 2)
	ts.Require().Equal(headers[0].GetHeight().GetRevisionHeight(), uint64(27709800))
	ts.Require().Equal(headers[1].GetHeight(), header.GetHeight())
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_Checkpoint() {
	latest := uint64(27710011)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	header := ts.assertSufficient(latest, iHeader)

	clientStateLatestHeight := types.NewHeight(0, iHeader.GetHeight().GetRevisionHeight()-1)
	headers, err := ts.prover.setupHeadersForUpdate(clientStateLatestHeight, header)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 1)
	ts.Require().Equal(headers[0].GetHeight(), header.GetHeight())
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader_AfterCheckpoint() {
	latest := uint64(27710012)
	iHeader, err := ts.prover.getLatestFinalizedHeader(latest)
	ts.Require().NoError(err)
	header := ts.assertSufficient(latest, iHeader)

	clientStateLatestHeight := types.NewHeight(0, iHeader.GetHeight().GetRevisionHeight()-1)
	headers, err := ts.prover.setupHeadersForUpdate(clientStateLatestHeight, header)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 1)
	ts.Require().Equal(headers[0].GetHeight().GetRevisionHeight(), uint64(27710001))

	clientStateLatestHeight = types.NewHeight(0, latest-400)
	headers, err = ts.prover.setupHeadersForUpdate(clientStateLatestHeight, header)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 3)
	ts.Require().Equal(headers[0].GetHeight().GetRevisionHeight(), uint64(27709800))
	ts.Require().Equal(headers[2].GetHeight().GetRevisionHeight(), uint64(27710000))
	ts.Require().Equal(headers[2].GetHeight().GetRevisionHeight(), uint64(27710002))
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
