package module

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type HeaderTestSuite struct {
	suite.Suite
	client *ethclient.Client
	reader *headerReader
}

func TestHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderTestSuite))
}

func (ts *HeaderTestSuite) SetupTest() {
	// https://github.com/datachainlab/lcp-bridge/tree/ddcbe10f7ec1cf302005892a87e2ae726535af33
	// - only one validator
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		panic(err)
	}
	ts.client = client
	ts.reader = NewHeaderReader(ts.client.BlockByNumber).(*headerReader)
}

func (ts *HeaderTestSuite) TestQueryETHHeaders() {
	// canonical chain
	ts.assertHeader(ts.reader, 0, 1)

	reader := NewHeaderReader(func(ctx context.Context, number *big.Int) (*types.Block, error) {
		header := &types.Header{}
		header.Number = number
		if header.Number.Int64()%int64(epochBlockPeriod) == 0 {
			if header.Number.Int64() == 0 {
				header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*4)
			} else {
				header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*21)
			}
		} else {
			header.Extra = make([]byte, extraVanity+extraSeal)
		}
		return types.NewBlock(header, nil, nil, nil, nil), nil
	})
	ts.assertHeader(reader, 0, 2)
	ts.assertHeader(reader, 1, 2)
	ts.assertHeader(reader, 199, 2)
	ts.assertHeader(reader, 200, 2)
	ts.assertHeader(reader, 201, 2)
	ts.assertHeader(reader, 202, 11)
}

func (ts *HeaderTestSuite) TestRequireCountToFinalize() {
	header := &types.Header{}
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*1)
	ts.Require().Equal(ts.reader.requiredCountToFinalize(header), 1)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*2)
	ts.Require().Equal(ts.reader.requiredCountToFinalize(header), 1)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*3)
	ts.Require().Equal(ts.reader.requiredCountToFinalize(header), 2)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*4)
	ts.Require().Equal(ts.reader.requiredCountToFinalize(header), 2)
	header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*21)
	ts.Require().Equal(ts.reader.requiredCountToFinalize(header), 11)
}

func (ts *HeaderTestSuite) assertHeader(reader HeaderReader, height int64, count int) {
	ethHeaders, err := reader.QueryETHHeaders(height)
	assert := ts.Require()
	assert.NoError(err)
	assert.Len(ethHeaders, count) // only one validator
	var header types.Header
	for i := 0; i < count; i++ {
		assert.NoError(rlp.DecodeBytes(ethHeaders[i].Header, &header))
		assert.Equal(header.Number.Int64(), height+int64(i))
	}
}
