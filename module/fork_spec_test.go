package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type ForkSpecTestSuite struct {
	suite.Suite
}

func TestForkSpecTestSuite(t *testing.T) {
	suite.Run(t, new(ForkSpecTestSuite))
}

func (ts *ForkSpecTestSuite) SetupTest() {
	_ = log.InitLogger("DEBUG", "text", "stdout")
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_ValidHeight() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Height{Height: 100}},
		{HeightOrTimestamp: &ForkSpec_Height{Height: 200}},
	}
	height := uint64(150)
	timestamp := uint64(0)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.NoError(err)
	ts.Equal(forkSpecs[0], current)
	ts.Equal(forkSpecs[0], previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_ValidHeight2() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Height{Height: 100}},
		{HeightOrTimestamp: &ForkSpec_Height{Height: 200}},
	}
	height := uint64(200)
	timestamp := uint64(0)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.NoError(err)
	ts.Equal(forkSpecs[1], current)
	ts.Equal(forkSpecs[0], previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_ValidTimestamp() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}},
		{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 2000}},
	}
	height := uint64(0)
	timestamp := uint64(1500)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.NoError(err)
	ts.Equal(forkSpecs[0], current)
	ts.Equal(forkSpecs[0], previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_ValidTimestamp2() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}},
		{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 2000}},
	}
	height := uint64(0)
	timestamp := uint64(2000)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.NoError(err)
	ts.Equal(forkSpecs[1], current)
	ts.Equal(forkSpecs[0], previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_NoMatch() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Height{Height: 100}},
		{HeightOrTimestamp: &ForkSpec_Height{Height: 200}},
	}
	height := uint64(50)
	timestamp := uint64(0)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.Error(err)
	ts.Nil(current)
	ts.Nil(previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_Both() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Height{Height: 100}},
		{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 200}},
	}
	height := uint64(50)
	timestamp := uint64(200)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.NoError(err)
	ts.Equal(forkSpecs[1], current)
	ts.Equal(forkSpecs[0], previous)
}

func (ts *ForkSpecTestSuite) Test_FindTargetForkSpec_EmptyForkSpecs() {
	forkSpecs := []*ForkSpec{}
	height := uint64(100)
	timestamp := uint64(1000)

	current, previous, err := FindTargetForkSpec(forkSpecs, height, timestamp)

	ts.Error(err)
	ts.Nil(current)
	ts.Nil(previous)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidHeight() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height))}, nil
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 50}}

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(BoundaryHeight(50), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestamp() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: uint64(1000)}, nil
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000 * 1000}} // msec

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(BoundaryHeight(100), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestampMultiHeader() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: height}, nil
	}
	currentHeight := uint64(1100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000 * 1000}} // msec

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(BoundaryHeight(1000), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestampMultiHeaderNotJust() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: height}, nil
	}
	currentHeight := uint64(1100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 999_999}} // msec

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(BoundaryHeight(1000), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_TimestampNotFound() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: uint64(500)}, nil
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}}

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(BoundaryHeight(0), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_HeaderFnError() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return nil, fmt.Errorf("header not found")
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}}

	boundaryHeight, err := GetBoundaryHeight(headerFn, currentHeight, currentForkSpec)

	ts.Error(err)
	ts.Equal(BoundaryHeight(0), boundaryHeight)
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs() {
	forkSpecs := []*ForkSpec{
		{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200},
		{EpochLength: 500},
	}
	epochs, err := BoundaryHeight(1501).GetBoundaryEpochs(*forkSpecs[1], *forkSpecs[0])
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1400))
	ts.Require().Equal(epochs.Intermediates, []uint64{1600, 1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight(1600).GetBoundaryEpochs(*forkSpecs[1], *forkSpecs[0])
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1600))
	ts.Require().Equal(epochs.Intermediates, []uint64{1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight(1601).GetBoundaryEpochs(*forkSpecs[1], *forkSpecs[0])
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1600))
	ts.Require().Equal(epochs.Intermediates, []uint64{1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight(1800).GetBoundaryEpochs(*forkSpecs[1], *forkSpecs[0])
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1800))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight(2000).GetBoundaryEpochs(*forkSpecs[1], *forkSpecs[0])
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))
}
