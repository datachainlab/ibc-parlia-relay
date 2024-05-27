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
	fn := func(h uint64) uint64 {
		if h%constant.BlocksPerEpoch != 0 {
			return toEpoch(h)
		}
		return h
	}
	ts.verifySetupHeadersForUpdate(fn, 5)
}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_beforeEpoch() {

	fn := func(h uint64) uint64 {
		if h%constant.BlocksPerEpoch == 0 {
			return h - 1
		}
		return h
	}
	ts.verifySetupHeadersForUpdate(fn, 6)
}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_afterEpoch() {
	fn := func(h uint64) uint64 {
		if h%constant.BlocksPerEpoch == 0 {
			return h + 1
		}
		return h
	}
	ts.verifySetupHeadersForUpdate(fn, 6)
}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_checkpoint() {
	fn := func(h uint64) uint64 {
		return ts.finalizedCheckpoint(h)
	}
	ts.verifySetupHeadersForUpdate(fn, 6)
}

func (ts *SetupNetworkTestSuite) TestSuccess_setupHeadersForUpdate_beforeCheckpoint() {
	fn := func(h uint64) uint64 {
		return ts.finalizedCheckpoint(h) - 1
	}
	ts.verifySetupHeadersForUpdate(fn, 6)
}

func (ts *SetupNetworkTestSuite) verifySetupHeadersForUpdate(editFinalizedHeight func(h uint64) uint64, expected int) {

	latestBlockNumber, err := ts.client.BlockNumber(context.Background())
	ts.Require().NoError(err)

	finalizedHeight, latestFinalizedHeader, err := queryLatestFinalizedHeader(ts.headerFn, latestBlockNumber)
	ts.Require().NoError(err)

	finalizedHeightAfter := editFinalizedHeight(finalizedHeight)
	if finalizedHeightAfter != finalizedHeight {
		finalizedHeight = finalizedHeightAfter
		latestFinalizedHeader, err = queryFinalizedHeader(ts.headerFn, finalizedHeight, latestBlockNumber)
		ts.Require().NoError(err)
	}

	verifiableLatestFinalizeHeader, err := withProofAndValidators(ts.headerFn, ts.accountProofFn, finalizedHeight, latestFinalizedHeader)
	ts.Require().NoError(err)

	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight, 0)
	ts.verify(verifiableLatestFinalizeHeader, latestBlockNumber, finalizedHeight-constant.BlocksPerEpoch*5, expected)

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

func (ts *SetupNetworkTestSuite) finalizedCheckpoint(h uint64) uint64 {
	prevEpoch := toEpoch(h) - constant.BlocksPerEpoch
	beforePrevEpoch := prevEpoch - constant.BlocksPerEpoch
	beforePrevValidatorSet, beforePrevTurnTerm, err := queryValidatorSetAndTurnTerm(ts.headerFn, beforePrevEpoch)
	ts.Require().NoError(err)
	checkpointValue := beforePrevValidatorSet.Checkpoint(beforePrevTurnTerm)
	log.GetLogger().Info("validator set", "len", len(beforePrevValidatorSet), "checkpoint", checkpointValue)
	checkpoint := checkpointValue + prevEpoch
	return checkpoint
}
