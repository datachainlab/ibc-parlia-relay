package module

import (
	"context"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
	"time"
)

type SetupTestSuite struct {
	suite.Suite
}

func TestSetupTestSuite(t *testing.T) {
	suite.Run(t, new(SetupTestSuite))
}

func (ts *SetupTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout", false)
	ts.Require().NoError(err)
}

var forkSpecsAfterMaxwell = []*ForkSpec{
	{
		// Must Set Milli timestamp
		HeightOrTimestamp:         &ForkSpec_Height{Height: 0},
		AdditionalHeaderItemCount: 1,
		EpochLength:               1000,
	},
}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_neighboringEpoch() {

	verify := func(trustedHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, trustedHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		neighborFn := func(_ context.Context, height uint64, _ uint64) (core.Header, error) {
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

		targets, err := setupHeadersForUpdate(context.Background(), neighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000), forkSpecsAfterMaxwell)
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

	verify(0, skip-1, 1)
	verify(0, skip, 1)
	verify(0, skip+1, 2)
	verify(0, 10*skip-1, 10)
	verify(0, 10*skip, 10)
	verify(0, 10*skip+1, 11)
	verify(skip-1, skip-1, 0)
	verify(skip-1, skip, 1)
	verify(skip-1, skip+1, 2)
	verify(skip-1, 10*skip-1, 10)
	verify(skip-1, 10*skip, 10)
	verify(skip-1, 10*skip+1, 11)
	verify(skip, skip, 0)
	verify(skip, skip+1, 1)
	verify(skip, 10*skip-1, 9)
	verify(skip, 10*skip, 9)
	verify(skip, 10*skip+1, 10)
	verify(skip+1, skip+1, 0)
	verify(skip+1, 10*skip-1, 9)
	verify(skip+1, 10*skip, 9)
	verify(skip+1, 10*skip+1, 10)

}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_allEmpty() {

	verify := func(trustedHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, trustedHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers: []*ETHHeader{target},
		}
		neighboringEpochFn := func(_ context.Context, height uint64, _ uint64) (core.Header, error) {
			// insufficient vote attestation
			return nil, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  epochHeader().Extra,
			}, nil
		}
		targets, err := setupHeadersForUpdate(context.Background(), neighboringEpochFn, headerFn, clientStateLatestHeight, latestFinalizedHeader,
			clienttypes.NewHeight(0,
				1000000), forkSpecsAfterMaxwell)
		ts.Require().NoError(err)
		ts.Require().Len(targets, expected)
	}

	verify(0, skip-1, 1)
	verify(0, skip, 1)
	verify(0, skip+1, 0) // non neighboring
	verify(0, 10*skip-1, 0)
	verify(0, 10*skip, 0)     // non neighboring
	verify(0, 10*skip+1, 0)   // non neighboring
	verify(skip-1, skip-1, 0) // same
	verify(skip-1, skip, 1)
	verify(skip-1, skip+1, 0) // non neighboring
	verify(skip-1, 10*skip-1, 0)
	verify(skip-1, 10*skip, 0)   // non neighboring
	verify(skip-1, 10*skip+1, 0) // non neighboring
	verify(skip, skip, 0)        // same
	verify(skip, skip+1, 1)
	verify(skip, 10*skip-1, 0)   // non neighboring
	verify(skip, 10*skip, 0)     // non neighboring
	verify(skip, 10*skip+1, 0)   // non neighboring
	verify(skip+1, skip+1, 0)    // same
	verify(skip+1, 10*skip-1, 0) // non neighboring
	verify(skip+1, 10*skip, 0)   // non neighboring
	verify(skip+1, 10*skip+1, 0) // non neighboring
}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_withHFBoundary() {

	verify := func(trustedHeight, nextHeight uint64, expected int) {
		now := time.Now()
		getTime := func(height uint64) time.Time {
			return now.Add(time.Duration(height) * time.Second)
		}

		hftime := now.Add(time.Duration(trustedHeight+(nextHeight-trustedHeight)/2) * time.Second).Unix()
		forkSpecs := append(forkSpecsAfterMaxwell, &ForkSpec{
			HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: uint64(hftime * 1000)},
			AdditionalHeaderItemCount: 1,
			EpochLength:               2000,
		})
		clientStateLatestHeight := clienttypes.NewHeight(0, trustedHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
			Time:   uint64(now.Unix()),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		neighborFn := func(ctx context.Context, height uint64, _ uint64) (core.Header, error) {
			h, e := newETHHeader(&types2.Header{
				Number: big.NewInt(int64(height)),
				Time:   uint64(getTime(height).Unix()),
			})
			return &Header{
				Headers: []*ETHHeader{h},
			}, e
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  epochHeader().Extra,
				Time:   uint64(getTime(height).Unix()),
			}, nil
		}

		targets, err := setupHeadersForUpdate(context.Background(), neighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000), forkSpecs)
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

	verify(0, 10*skip-1, 10+1)
	verify(0, 10*skip, 10+1)
	verify(0, 10*skip+1, 11+1)
	verify(skip-1, skip-1, 0)
	verify(skip-1, skip, 1)
	verify(skip-1, skip+1, 2)
	verify(skip-1, 10*skip-1, 10+1)
	verify(skip-1, 10*skip, 10+1)
	verify(skip-1, 10*skip+1, 11+1)
	verify(skip, skip, 0)
	verify(skip, skip+1, 1)
	verify(skip, 10*skip-1, 9+1)
	verify(skip, 10*skip, 9+1)
	verify(skip, 10*skip+1, 10+1)
	verify(skip+1, skip+1, 0)
	verify(skip+1, 10*skip-1, 9+1)
	verify(skip+1, 10*skip, 9+1)
	verify(skip+1, 10*skip+1, 10+1)

}

func (ts *SetupTestSuite) Test_makeSubmittingHeights() {
	rq := ts.Require()
	msec := uint64(0)
	rq.Len(makeSubmittingHeights(10, 1, 11, nil, 0), 0)
	rq.Len(makeSubmittingHeights(10, 1, 11, &msec, 11), 0)
	rq.Len(makeSubmittingHeights(10, 1, 11, &msec, 9), 1)
	rq.Len(makeSubmittingHeights(10, 9, 11, &msec, 9), 0)
	rq.Equal(
		[]uint64{skip - 1, skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, skip-1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, skip-1, skip, &msec, skip-1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, skip),
	)
	rq.Equal(
		[]uint64{skip, skip + 1, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, nil, skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 2*skip + 1, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, 2*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 3*skip + 1, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, 3*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 4*skip + 1, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, 4*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, nil, 4*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+1, 0, skip, &msec, 5*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip, 5*skip + 1},
		makeSubmittingHeights(5*skip+2, 0, skip, &msec, 5*skip+1),
	)
	rq.Equal(
		[]uint64{skip, 2 * skip, 3 * skip, 4 * skip, 5 * skip},
		makeSubmittingHeights(5*skip+2, 0, skip, nil, 5*skip+1),
	)
}
