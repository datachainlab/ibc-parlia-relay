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
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)
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

		targets, err := setupHeadersForUpdate(neighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000), GetForkParameters(Localnet))
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
	verify(skip+1, 10*skip-1, 10)
	verify(skip+1, 10*skip, 10)
	verify(skip+1, 10*skip+1, 11)

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
		targets, err := setupHeadersForUpdate(neighboringEpochFn, headerFn, clientStateLatestHeight, latestFinalizedHeader,
			clienttypes.NewHeight(0,
				1000000), GetForkParameters(Localnet))
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
		forkSpecs := []*ForkSpec{
			{
				// Must Set Milli timestamp
				HeightOrTimestamp:         &ForkSpec_Height{Height: 0},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: uint64(hftime * 1000)},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
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
		neighborFn := func(height uint64, _ uint64) (core.Header, error) {
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

		targets, err := setupHeadersForUpdate(neighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000), forkSpecs)
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
	verify(skip+1, 10*skip-1, 10+1)
	verify(skip+1, 10*skip, 10+1)
	verify(skip+1, 10*skip+1, 11+1)

}

func (ts *SetupTestSuite) Test_makeSubmittingHeights() {
	rq := ts.Require()
	msec := uint64(0)
	rq.Len(makeSubmittingHeights(10, 11, nil, 0), 0)
	rq.Len(makeSubmittingHeights(10, 11, &msec, 11), 0)
	rq.Len(makeSubmittingHeights(10, 11, &msec, 9), 1)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 99),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 100),
	)
	rq.Equal(
		[]uint64{100, 101, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 101),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, nil, 101),
	)
	rq.Equal(
		[]uint64{100, 200, 201, 300, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 201),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 301, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 301),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 401, 500},
		makeSubmittingHeights(501, 100, &msec, 401),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, nil, 401),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(501, 100, &msec, 501),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500, 501},
		makeSubmittingHeights(502, 100, &msec, 501),
	)
	rq.Equal(
		[]uint64{100, 200, 300, 400, 500},
		makeSubmittingHeights(502, 100, nil, 501),
	)

}
