package module

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
		header := &types.Header{
			Number: big.NewInt(int64(height)),
		}
		if height == 31835601 || height == 31835602 {
			header.Extra = common.Hex2Bytes("d88301020b846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5831defffb860a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37f84c8401e5c5cfa0be938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c14648401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28027cb6d065d5a6d8749ca45a185add61b9ce470136898643170f8072513ca45f35d826f02cb2494f857beebdac9ec04196c8b30a65352ef155a28ac6a0057ff1601")
		}
		return header, nil
	}

	// No finalized header found ( invalid relation )
	headers, err = queryFinalizedHeader(fn, 31835592, 31835602)
	ts.Require().NoError(err)
	ts.Require().Nil(headers)
}

func (ts *HeaderQueryTestSuite) TestSuccessQueryFinalizedHeader() {
	ts.Require().NoError(log.InitLogger("INFO", "json", "stdout"))
	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		header := &types.Header{
			Number: big.NewInt(int64(height)),
		}
		if height == 31835601 {
			header.Extra = common.Hex2Bytes("d88301020b846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5831defffb860a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37f84c8401e5c5cfa0be938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c14648401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28027cb6d065d5a6d8749ca45a185add61b9ce470136898643170f8072513ca45f35d826f02cb2494f857beebdac9ec04196c8b30a65352ef155a28ac6a0057ff1601")
		} else if height == 31835602 {
			header.Extra = common.Hex2Bytes("d88301020b846765746888676f312e31392e38856c696e7578000000b19df4a2f8b5831defffb860a244628caa7b3002a245b677c419c5991d9ba62e7d298e96565b72f8ccc6587510f8827c00783d0a13326bfc72bbcbb90e6bdf988ef662b286158296e0f270f21568fdb75210f631d53b81e74f0fa9a5c591dc46cbeceb28952264d8863b7812f84c8401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28401e5c5d1a06b3b459206a5b6b1963e686318b0261b9c0888e1a253f77d109c60c6734c84c28031c42276b8ebf15bb5b843865147ea9435be29a83afeae646fc156b45832e0016bb3fa7119db6fe5dfe5d99733b6f7dd38ac4d7aeb7882cd4b6c576faf6951a901")
		}
		return header, nil
	}

	headers, err := queryFinalizedHeader(fn, 31835592, 31835602)
	ts.Require().NoError(err)
	ts.Require().Len(headers, 11)
}

func (ts *HeaderQueryTestSuite) TestSuccessQueryLatestFinalizedHeader() {

	verify := func(latestBlockNumber uint64) {
		getHeader := func(ctx context.Context, height uint64) (*types.Header, error) {
			header := &types.Header{
				Number: big.NewInt(int64(height)),
			}
			if height == 31835601 {
				header.Extra = common.Hex2Bytes("d88301020b846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5831defffb860a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37f84c8401e5c5cfa0be938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c14648401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28027cb6d065d5a6d8749ca45a185add61b9ce470136898643170f8072513ca45f35d826f02cb2494f857beebdac9ec04196c8b30a65352ef155a28ac6a0057ff1601")
			} else if height == 31835602 {
				header.Extra = common.Hex2Bytes("d88301020b846765746888676f312e31392e38856c696e7578000000b19df4a2f8b5831defffb860a244628caa7b3002a245b677c419c5991d9ba62e7d298e96565b72f8ccc6587510f8827c00783d0a13326bfc72bbcbb90e6bdf988ef662b286158296e0f270f21568fdb75210f631d53b81e74f0fa9a5c591dc46cbeceb28952264d8863b7812f84c8401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28401e5c5d1a06b3b459206a5b6b1963e686318b0261b9c0888e1a253f77d109c60c6734c84c28031c42276b8ebf15bb5b843865147ea9435be29a83afeae646fc156b45832e0016bb3fa7119db6fe5dfe5d99733b6f7dd38ac4d7aeb7882cd4b6c576faf6951a901")
			}
			return header, nil
		}
		height, h, err := queryLatestFinalizedHeader(getHeader, latestBlockNumber)
		ts.Require().NoError(err)
		ts.Require().Len(h, 3)
		ts.Require().Equal(int(height), 31835600)
	}
	for i := 31835602; i < 31835602+100; i++ {
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

func TestEth(t *testing.T) {
	client, err := ethclient.Dial("https://data-seed-prebsc-1-s1.binance.org:8545")
	if err != nil {
		t.Fatal(err)
	}
	header, err := client.HeaderByNumber(context.Background(), big.NewInt(39548009))
	if err != nil {
		t.Fatal(err)
	}
	h, err := newETHHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	wh := Header{
		Headers: []*ETHHeader{h},
	}
	decoded, err := wh.decodeEthHeaders()
	if err != nil {
		t.Fatal(err)
	}
	for _, hh := range decoded {
		hash := hh.Hash()
		println(hh.WithdrawalsHash.String(), hash.String())
	}

}
