package module

import (
	"log"
	"testing"

	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
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
		EthChainId: 56,
		// We can get accountProof by eth_geProof only from AllThatNode
		RpcAddr:     "https://bsc-mainnet-rpc.allthatnode.com",
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		// TODO change address after starting mainnet test
		IbcAddress: "0x151f3951FA218cac426edFe078fA9e5C6dceA500",
	})
	ts.Require().NoError(err)

	config := ProverConfig{
		Debug: true,
	}
	ts.prover = NewProver(NewChain(chain), &config).(*Prover)
}

func (ts *ProverMainnetTestSuite) TestQueryLatestFinalizedHeader() {

	latestHeight, err := ts.prover.chain.LatestHeight()
	ts.Require().NoError(err)
	latest := latestHeight.GetRevisionHeight()
	println(latest)
	iHeader, err := ts.prover.GetLatestFinalizedHeaderByLatestHeight(latest)
	ts.Require().NoError(err)
	ts.Require().NoError(iHeader.ValidateBasic())

	requiredBlocksToFinalizeInCurrentMainnet := 11
	header := iHeader.(*Header)
	ts.Require().Len(header.Headers, requiredBlocksToFinalizeInCurrentMainnet)

	// target header
	target, err := header.Target()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number.Uint64(), latest-uint64(requiredBlocksToFinalizeInCurrentMainnet-1))

	// headers to verify
	ethHeaders, err := header.decodeEthHeaders()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number, ethHeaders[0].Number)

	for i, eth := range ethHeaders {
		if i > 0 {
			ts.Require().Equal(eth.Number.Uint64()-1, ethHeaders[i-1].Number.Uint64())
		}
		if eth.Number.Uint64()%200 == 0 {
			validators, err := extractValidatorSet(eth)
			ts.Require().NoError(err)
			ts.Require().Len(validators, 21)
		}
	}

	// account proof
	account, err := header.Account(ts.prover.chain.IBCAddress())
	ts.Require().NoError(err)
	ts.Require().NotEqual(account.Root, common.Hash{})
	log.Println(account.Root)

	// setup
	updating, err := ts.prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header)
	ts.Require().NoError(err)
	ts.Require().Len(updating, 1)
	ts.Require().Equal(updating[0].(*Header).GetHeight(), header.GetHeight())

	// updating msg
	pack, err := types.PackClientMessage(updating[0])
	ts.Require().NoError(err)
	marshal, err := pack.Marshal()
	ts.Require().NoError(err)
	log.Println(common.Bytes2Hex(marshal))

}
