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
	_ = log.InitLogger("DEBUG", "text", "stdout", false)
	boundaryHeightCache = make(map[uint64]uint64)
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
	ts.Equal(forkSpecs[0], previous[0])
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
	ts.Equal(forkSpecs[0], previous[0])
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
	ts.Equal(forkSpecs[0], previous[0])
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
	ts.Equal(forkSpecs[0], previous[0])
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
	ts.Equal(forkSpecs[0], previous[0])
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

	boundaryHeight, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(uint64(50), boundaryHeight.Height)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestamp() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: uint64(1000)}, nil
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000 * 1000}} // msec

	boundaryHeight, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(uint64(100), boundaryHeight.Height)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestampMultiHeader() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: height}, nil
	}
	currentHeight := uint64(1100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000_000}} // msec

	boundaryHeight, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(uint64(1000), boundaryHeight.Height)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_ValidTimestampMultiHeaderNotJust() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: height}, nil
	}
	currentHeight := uint64(1100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 999_999}} // msec

	boundaryHeight, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(uint64(1000), boundaryHeight.Height)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_TimestampNotFound() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{Number: big.NewInt(int64(height)), Time: uint64(500)}, nil
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}}

	boundaryHeight, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.NoError(err)
	ts.Equal(uint64(0), boundaryHeight.Height)
}

func (ts *ForkSpecTestSuite) Test_GetBoundaryHeight_HeaderFnError() {
	headerFn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return nil, fmt.Errorf("header not found")
	}
	currentHeight := uint64(100)
	currentForkSpec := ForkSpec{HeightOrTimestamp: &ForkSpec_Timestamp{Timestamp: 1000}}

	_, err := GetBoundaryHeight(context.Background(), headerFn, currentHeight, currentForkSpec)

	ts.Error(err)
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs() {
	pascalHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200}
	lorentzHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 500}
	maxwellHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}
	postMaxwellHF1 := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}
	postMaxwellHF2 := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 2000}

	// Check lorentzHF
	epochs, err := BoundaryHeight{Height: 1501, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1400))
	ts.Require().Equal(epochs.Intermediates, []uint64{1600, 1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 1600, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1600))
	ts.Require().Equal(epochs.Intermediates, []uint64{1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 1601, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1600))
	ts.Require().Equal(epochs.Intermediates, []uint64{1800})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 1800, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1800))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2000, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	// Check maxwell
	epochs, err = BoundaryHeight{Height: 1501, CurrentForkSpec: *maxwellHF}.GetBoundaryEpochs([]*ForkSpec{lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1500))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2000, CurrentForkSpec: *maxwellHF}.GetBoundaryEpochs([]*ForkSpec{lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2001, CurrentForkSpec: *maxwellHF}.GetBoundaryEpochs([]*ForkSpec{lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{2500})
	ts.Require().Equal(epochs.CurrentFirst, uint64(3000))

	epochs, err = BoundaryHeight{Height: 3000, CurrentForkSpec: *maxwellHF}.GetBoundaryEpochs([]*ForkSpec{lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(3000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(3000))

	// Check post maxwell 1
	epochs, err = BoundaryHeight{Height: 1501, CurrentForkSpec: *postMaxwellHF1}.GetBoundaryEpochs([]*ForkSpec{maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2000, CurrentForkSpec: *postMaxwellHF1}.GetBoundaryEpochs([]*ForkSpec{maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2001, CurrentForkSpec: *postMaxwellHF1}.GetBoundaryEpochs([]*ForkSpec{maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(3000))

	epochs, err = BoundaryHeight{Height: 3000, CurrentForkSpec: *postMaxwellHF1}.GetBoundaryEpochs([]*ForkSpec{maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(3000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(3000))

	// Check post maxwell 2
	epochs, err = BoundaryHeight{Height: 1501, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(1000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2000, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))

	epochs, err = BoundaryHeight{Height: 2001, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(2000))
	ts.Require().Equal(epochs.Intermediates, []uint64{3000})
	ts.Require().Equal(epochs.CurrentFirst, uint64(4000))

	epochs, err = BoundaryHeight{Height: 3000, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(3000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(4000))

	epochs, err = BoundaryHeight{Height: 4000, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(4000))
	ts.Require().Equal(epochs.Intermediates, []uint64{})
	ts.Require().Equal(epochs.CurrentFirst, uint64(4000))

	epochs, err = BoundaryHeight{Height: 4001, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(4000))
	ts.Require().Equal(epochs.Intermediates, []uint64{5000})
	ts.Require().Equal(epochs.CurrentFirst, uint64(6000))
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs_After_Lorentz() {
	pascalHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200}
	lorentzHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 500}

	epochs, err := BoundaryHeight{Height: 1, CurrentForkSpec: *lorentzHF}.GetBoundaryEpochs([]*ForkSpec{pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(0))
	ts.Require().Equal(epochs.Intermediates, []uint64{200, 400})
	ts.Require().Equal(epochs.CurrentFirst, uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(199), uint64(0))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(200), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(399), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(400), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(499), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(500), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(501), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(999), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1000), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1001), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1499), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1500), uint64(1500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1501), uint64(1500))

	ts.Require().Equal(epochs.PreviousEpochBlockNumber(0), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(200), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(400), uint64(200))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(500), uint64(400))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(1000), uint64(500))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(1500), uint64(1000))
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs_After_Maxwell() {
	pascalHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200}
	lorentzHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 500}
	maxwellHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}

	epochs, err := BoundaryHeight{Height: 1, CurrentForkSpec: *maxwellHF}.GetBoundaryEpochs([]*ForkSpec{lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(0))
	ts.Require().Equal(epochs.Intermediates, []uint64{200, 400, 500})
	ts.Require().Equal(epochs.CurrentFirst, uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(199), uint64(0))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(200), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(399), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(400), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(499), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(500), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(501), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(999), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1000), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1001), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1499), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1500), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1501), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1999), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2000), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2001), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2999), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(3000), uint64(3000))

	ts.Require().Equal(epochs.PreviousEpochBlockNumber(0), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(200), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(400), uint64(200))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(500), uint64(400))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(1000), uint64(500))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(2000), uint64(1000))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(3000), uint64(2000))
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs_After_Maxwell_1() {
	pascalHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200}
	lorentzHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 500}
	maxwellHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}
	postMaxwellHF1 := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}

	epochs, err := BoundaryHeight{Height: 1, CurrentForkSpec: *postMaxwellHF1}.GetBoundaryEpochs([]*ForkSpec{maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(0))
	ts.Require().Equal(epochs.Intermediates, []uint64{200, 400, 500})
	ts.Require().Equal(epochs.CurrentFirst, uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(199), uint64(0))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(200), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(399), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(400), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(499), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(500), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(501), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(999), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1000), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1001), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1499), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1500), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1501), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1999), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2000), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2001), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2999), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(3000), uint64(3000))

	ts.Require().Equal(epochs.PreviousEpochBlockNumber(0), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(200), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(400), uint64(200))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(500), uint64(400))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(1000), uint64(500))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(2000), uint64(1000))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(3000), uint64(2000))
}

func (ts *ForkSpecTestSuite) Test_Success_GetBoundaryEpochs_After_Maxwell_2() {
	pascalHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 0}, EpochLength: 200}
	lorentzHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 500}
	maxwellHF := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}
	postMaxwellHF1 := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 1000}
	postMaxwellHF2 := &ForkSpec{HeightOrTimestamp: &ForkSpec_Height{Height: 1}, EpochLength: 2000}

	epochs, err := BoundaryHeight{Height: 1, CurrentForkSpec: *postMaxwellHF2}.GetBoundaryEpochs([]*ForkSpec{postMaxwellHF1, maxwellHF, lorentzHF, pascalHF})
	ts.Require().NoError(err)
	ts.Require().Equal(epochs.PrevLast, uint64(0))
	ts.Require().Equal(epochs.Intermediates, []uint64{200, 400, 500, 1000})
	ts.Require().Equal(epochs.CurrentFirst, uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(199), uint64(0))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(200), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(399), uint64(200))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(400), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(499), uint64(400))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(500), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(501), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(999), uint64(500))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1000), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1001), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1499), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1500), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1501), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(1999), uint64(1000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2000), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2001), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(2999), uint64(2000))
	ts.Require().Equal(epochs.CurrentEpochBlockNumber(3000), uint64(2000))

	ts.Require().Equal(epochs.PreviousEpochBlockNumber(0), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(200), uint64(0))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(400), uint64(200))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(500), uint64(400))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(1000), uint64(500))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(2000), uint64(1000))
	ts.Require().Equal(epochs.PreviousEpochBlockNumber(3000), uint64(1000))
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_NoPrevious() {
	currentHeader := &types.Header{Number: big.NewInt(int64(123)), Time: 123_456}
	targetTs := 123_456
	distance := estimateDistance(nil, currentHeader, uint64(targetTs))
	ts.Require().Equal(uint64(1), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_MoveLowAfterLow() {
	previousHeader := &types.Header{Number: big.NewInt(int64(200)), Time: 60_000}
	currentHeader := &types.Header{Number: big.NewInt(int64(100)), Time: 20_000}
	distance := estimateDistance(previousHeader, currentHeader, 10_987_000)
	ts.Require().Equal(uint64(23), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_MoveHighAfterLow() {
	previousHeader := &types.Header{Number: big.NewInt(int64(200)), Time: 60_000}
	currentHeader := &types.Header{Number: big.NewInt(int64(100)), Time: 20_000}
	distance := estimateDistance(previousHeader, currentHeader, 20_720_000)
	ts.Require().Equal(uint64(2), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_MoveLowAfterHigh() {
	previousHeader := &types.Header{Number: big.NewInt(int64(100)), Time: 85_000}
	currentHeader := &types.Header{Number: big.NewInt(int64(200)), Time: 90_000}
	distance := estimateDistance(previousHeader, currentHeader, 97_123_000)
	ts.Require().Equal(uint64(143), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_MoveHighAfterHigh() {
	previousHeader := &types.Header{Number: big.NewInt(int64(100)), Time: 85_000}
	currentHeader := &types.Header{Number: big.NewInt(int64(200)), Time: 90_000}
	distance := estimateDistance(previousHeader, currentHeader, 87_123_000)
	ts.Require().Equal(uint64(58), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_SameTimestamp() {
	previousHeader := &types.Header{Number: big.NewInt(int64(100)), Time: 99_900}
	currentHeader := &types.Header{Number: big.NewInt(int64(200)), Time: 99_900}
	distance := estimateDistance(previousHeader, currentHeader, 50_000)
	ts.Require().Equal(uint64(1), distance)
}

func (ts *ForkSpecTestSuite) Test_estimateDistance_MinimumRate() {
	previousHeader := &types.Header{Number: big.NewInt(17), Time: math.MaxUint64}
	currentHeader := &types.Header{Number: big.NewInt(16), Time: 0}
	distance := estimateDistance(previousHeader, currentHeader, 50_000)
	ts.Require().Equal(uint64(1), distance)
}
