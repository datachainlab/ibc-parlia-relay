package module

import (
	"context"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"os"
	"testing"
)

type SetupNetworkTestSuite struct {
	suite.Suite
	client         *ethclient.Client
	headerFn       getHeaderFn
	accountProofFn getAccountProof
}

func TestSetupNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(SetupNetworkTestSuite))
}

func (ts *SetupNetworkTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)

	rpcNode := os.Getenv("BSC_RPC_NODE")
	if rpcNode == "" {
		// https://docs.bscscan.com/misc-tools-and-utilities/public-rpc-nodes
		rpcNode = "https://data-seed-prebsc-1-s1.binance.org:8545"
	}
	client, err := ethclient.Dial(rpcNode)
	ts.Require().NoError(err)

	ts.client = client

	ts.headerFn = func(ctx context.Context, height uint64) (*types2.Header, error) {
		return ts.client.HeaderByNumber(ctx, big.NewInt(0).SetUint64(height))
	}
	ts.accountProofFn = func(height int64) ([]byte, common.Hash, error) {
		return nil, common.Hash{}, nil
	}
}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_epoch() {

	latestBlockNumber, err := ts.client.BlockNumber(context.Background())
	ts.Require().NoError(err)

	finalizedHeight, latestFinalizedHeader, err := queryLatestFinalizedHeader(ts.headerFn, latestBlockNumber)
	ts.Require().NoError(err)

	// force epoch
	if finalizedHeight%constant.BlocksPerEpoch != 0 {
		finalizedHeight = toEpoch(finalizedHeight)
		latestFinalizedHeader, err = queryFinalizedHeader(ts.headerFn, finalizedHeight, latestBlockNumber)
	}

	verifiableLatestFinalizeHeader, err := withProofAndValidators(ts.headerFn, ts.accountProofFn, finalizedHeight, latestFinalizedHeader)
	ts.Require().NoError(err)

	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight, 0)
	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight-constant.BlocksPerEpoch*5, 5)

}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_notEpoch() {

	latestBlockNumber, err := ts.client.BlockNumber(context.Background())
	ts.Require().NoError(err)

	finalizedHeight, latestFinalizedHeader, err := queryLatestFinalizedHeader(ts.headerFn, latestBlockNumber)
	ts.Require().NoError(err)

	// force epoch - 1
	if finalizedHeight%constant.BlocksPerEpoch == 0 {
		finalizedHeight = finalizedHeight - 1
		latestFinalizedHeader, err = queryFinalizedHeader(ts.headerFn, finalizedHeight, latestBlockNumber)
	}

	verifiableLatestFinalizeHeader, err := withProofAndValidators(ts.headerFn, ts.accountProofFn, finalizedHeight, latestFinalizedHeader)
	ts.Require().NoError(err)

	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight, 0)
	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight-constant.BlocksPerEpoch*5, 6)

}

func (ts *SetupNetworkTestSuite) verify(verifiableLatestFinalizeHeader core.Header, latestBlockNumber, trustedHeight uint64, expected int) {
	clientStateLatestHeight := clienttypes.NewHeight(0, trustedHeight)
	queryVerifiableNeighboringEpochHeader := func(height uint64, limitHeight uint64) (core.Header, error) {
		ethHeaders, err := queryFinalizedHeader(ts.headerFn, height, limitHeight)
		if err != nil {
			return nil, err
		}
		// No finalized header found
		if ethHeaders == nil {
			return nil, nil
		}
		return withProofAndValidators(ts.headerFn, ts.accountProofFn, height, ethHeaders)
	}
	targets, err := setupHeadersForUpdate(queryVerifiableNeighboringEpochHeader, ts.headerFn, clientStateLatestHeight, verifiableLatestFinalizeHeader.(*Header), clienttypes.NewHeight(0, latestBlockNumber))
	ts.Require().NoError(err)
	ts.Require().Len(targets, expected)
	for i, h := range targets {
		trusted := h.(*Header).TrustedHeight
		if i == 0 {
			ts.Require().Equal(trusted.RevisionHeight, trustedHeight)
		} else {
			ts.Require().Equal(*trusted, targets[i-1].GetHeight())
		}
	}
}
