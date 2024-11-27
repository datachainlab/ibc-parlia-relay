package module

import (
	"context"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type SetupTestSuite struct {
	suite.Suite
}

func TestSetupTestSuite(t *testing.T) {
	suite.Run(t, new(SetupTestSuite))
}

func (ts *SetupTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)
}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_neighboringEpoch() {

	verify := func(latestHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		neighborFn := func(height uint64, _ uint64) (core.Header, error) {
			h, e := newETHHeader(&types2.Header{
				Number: big.NewInt(int64(height)),
			})
			return &Header{
				Headers: []*ETHHeader{h},
			}, e
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  epochHeader().Extra,
			}, nil
		}

		targets, err := setupHeadersForUpdate(neighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000))
		ts.Require().NoError(err)
		ts.Require().Len(targets, expected)
		for i, h := range targets {
			trusted := h.(*Header).TrustedHeight
			if i == 0 {
				ts.Require().Equal(trusted.RevisionHeight, latestHeight)
			} else {
				ts.Require().Equal(*trusted, targets[i-1].GetHeight())
			}
		}
	}

	verify(0, constant.BlocksPerEpoch-1, 1)
	verify(0, constant.BlocksPerEpoch, 1)
	verify(0, constant.BlocksPerEpoch+1, 2)
	verify(0, 10*constant.BlocksPerEpoch-1, 10)
	verify(0, 10*constant.BlocksPerEpoch, 10)
	verify(0, 10*constant.BlocksPerEpoch+1, 11)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, 1)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, 2)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, 10)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, 10)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, 11)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, 1)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, 9)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, 9)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, 10)
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, 9)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, 9)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, 10)

}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_allEmpty() {

	verify := func(latestHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers: []*ETHHeader{target},
		}
		neighboringEpochFn := func(height uint64, _ uint64) (core.Header, error) {
			// insufficient vote attestation
			return nil, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  epochHeader().Extra,
			}, nil
		}
		targets, err := setupHeadersForUpdate(neighboringEpochFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 1000000))
		ts.Require().NoError(err)
		ts.Require().Len(targets, expected)
	}

	verify(0, constant.BlocksPerEpoch-1, 1)
	verify(0, constant.BlocksPerEpoch, 1)
	verify(0, constant.BlocksPerEpoch+1, 0) // non neighboring
	verify(0, 10*constant.BlocksPerEpoch-1, 0)
	verify(0, 10*constant.BlocksPerEpoch, 0)                        // non neighboring
	verify(0, 10*constant.BlocksPerEpoch+1, 0)                      // non neighboring
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, 0) // same
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, 1)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, 0) // non neighboring
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, 0)   // non neighboring
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, 0) // non neighboring
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, 0)        // same
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, 1)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, 0)   // non neighboring
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, 0)     // non neighboring
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, 0)   // non neighboring
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, 0)    // same
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, 0) // non neighboring
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, 0)   // non neighboring
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, 0) // non neighboring
}

func (ts *SetupTestSuite) TestSuccess_setupNeighboringEpochHeader() {

	epochHeight := uint64(400)
	trustedEpochHeight := uint64(200)

	neighboringEpochFn := func(height uint64, limit uint64) (core.Header, error) {
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(limit)),
		})
		ts.Require().NoError(err)
		return &Header{
			Headers: []*ETHHeader{target},
		}, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		return headerByHeight(int64(height)), nil
	}
	hs, err := setupNeighboringEpochHeader(headerFn, neighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 10000))
	ts.Require().NoError(err)
	target, err := hs.(*Header).Target()
	ts.Require().NoError(err)

	// next checkpoint - 1
	ts.Require().Equal(int64(602), target.Number.Int64())
}
