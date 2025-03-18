package module

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"strings"
	"testing"
)

type HeaderQueryTestSuite struct {
	suite.Suite
}

func TestHeaderQueryTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderQueryTestSuite))
}

func (ts *HeaderQueryTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)
}

func (ts *HeaderQueryTestSuite) TestErrorQueryFinalizedHeader() {
	ts.Require().NoError(log.InitLogger("INFO", "json", "stdout"))
	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{
			Number: big.NewInt(int64(height)),
		}, nil
	}

	// No finalized header found
	headers, err := queryFinalizedHeader(fn, 1, 10)
	ts.Require().NoError(err)
	ts.Require().Nil(headers)

	fn = func(ctx context.Context, height uint64) (*types.Header, error) {
		h := headerByHeight(int64(height))
		if h != nil {
			return h, nil
		}
		return &types.Header{Number: big.NewInt(int64(height))}, nil
	}

	headers, err = queryFinalizedHeader(fn, 760, 1000)
	ts.Require().NoError(err)
	ts.Require().Nil(headers)
}

func (ts *HeaderQueryTestSuite) TestSuccessQueryFinalizedHeader() {
	ts.Require().NoError(log.InitLogger("INFO", "json", "stdout"))
	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		h := headerByHeight(int64(height))
		if h != nil {
			return h, nil
		}
		return &types.Header{Number: big.NewInt(int64(height))}, nil
	}

	headers, err := queryFinalizedHeader(fn, 760, 1002)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 1002-760)
}

func (ts *HeaderQueryTestSuite) TestSuccessQueryLatestFinalizedHeader() {

	verify := func(latestBlockNumber uint64) {
		getHeader := func(ctx context.Context, height uint64) (*types.Header, error) {
			h := headerByHeight(int64(height))
			if h != nil {
				return h, nil
			}
			return &types.Header{Number: big.NewInt(int64(height))}, nil
		}
		height, h, err := queryLatestFinalizedHeader(getHeader, latestBlockNumber)
		ts.Require().NoError(err)
		ts.Require().Len(h, 3)
		ts.Require().Equal(int(height), 1001)
	}
	for i := 1003; i < 1003+100; i++ {
		verify(uint64(i))
	}
}

func (ts *HeaderQueryTestSuite) TestErrorQueryLatestFinalizedHeader_NoVote() {

	verify := func(latestBlockNumber uint64, extra []byte) {
		getHeader := func(ctx context.Context, height uint64) (*types.Header, error) {
			return &types.Header{
				Number: big.NewInt(int64(height)),
				Extra:  extra,
			}, nil
		}
		_, _, err := queryLatestFinalizedHeader(getHeader, latestBlockNumber)
		ts.Require().True(strings.Contains(err.Error(), "no finalized header found"))
	}

	for i := 0; i < 100; i++ {
		verify(uint64(i), make([]byte, 0))
		verify(uint64(i), common.Hex2Bytes("d88301020b846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5831defffb860a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37f84c8401e5c5cfa0be938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c14648401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28027cb6d065d5a6d8749ca45a185add61b9ce470136898643170f8072513ca45f35d826f02cb2494f857beebdac9ec04196c8b30a65352ef155a28ac6a0057ff1601"))
	}
}
