package module

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type HeaderTestSuite struct {
	suite.Suite
	reader *headerReader
}

func TestHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderTestSuite))
}

func (ts *HeaderTestSuite) SetupTest() {
	ts.reader = NewHeaderReader(func(ctx context.Context, number uint64) (*types.Header, error) {
		header := &types.Header{}
		header.Number = big.NewInt(int64(number))
		if header.Number.Int64()%epochBlockPeriod == 0 {
			if header.Number.Int64() == 0 {
				header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*4)
			} else {
				header.Extra = make([]byte, extraVanity+extraSeal+validatorBytesLength*21)
			}
		} else {
			header.Extra = make([]byte, extraVanity+extraSeal)
		}
		return header, nil
	}).(*headerReader)
}

func (ts *HeaderTestSuite) TestQueryETHHeaders() {
	ts.assertHeader(0, 2)
	ts.assertHeader(1, 2)
	ts.assertHeader(199, 2)
	ts.assertHeader(200, 2)
	ts.assertHeader(201, 2)
	ts.assertHeader(202, 11)
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

func (ts *HeaderTestSuite) TestNewHeaderSuccess() {
	rawHeader := types.Header{
		Root:   common.HexToHash("c84307dfe4ccfec4a851a77755d63228d8e0b9ba3345d1eee37ed729ee16eaa1"),
		Number: big.NewInt(21400),
		Extra:  common.Hex2Bytes("d983010111846765746889676f312e31362e3135856c696e75780000d5efc574b794a294a7e455058789c82f15e8ce00669d689e2090e08cc7f42b0bd007728dea17979718bc99e2261bfe892a7ac9b849545ffc597338db43a9cb00564971a8f282cfe1741b0294ba45a21ac35168138e2c2e8200"),
	}
	ethHeader, err := ts.reader.newETHHeader(&rawHeader)
	ts.Require().NoError(err)

	accountProof := []string{
		"0xf901f1a080679a623dfdd0dfa34cb5c1db80292abdc2a9e75f5026e3d24cd10ea58f8e0da04f4d7ef0e932874f07aec064ee1281cd6a3245fceab78bdd6a8d2d7a86d27451a0a715335e2de6e91c28910eff04e8709ff6ca93862121a0b52560071867a9f14080a0842db9556e659d64ca9d2d33229ebac6e7e2185f42bd07965464de8064d94ac8a0d94bd2db341ab9d3d7f4fe0aa569bb21dfac0d5eb0ec008c7af23d7f2ed98ec1a0cee66e2515872d5f4b42ada7cc733288809c11ab99aa0d25eb941236074f9904a0f3aa8d1774f013d8af0fdd8364c7833f16b42ad377e485b754f5cdae6fedaa2fa0bffc6b17aaf862725aaf4af4ecda3ed70d4102b875451eb965259ead260b06c7a026a29f57f5efaf83a8f098ed0ba0f53aac353364ce498a82d589e7bcf1f84e76a01a25f2cac2c6a021225ea182c3c391c0fafac96cb38896eb45648a5c33f31b6ca0d4d6b410f89044b335cc7b14221050035d87d390043bf6d84bc0f8005391f213a092dfa1004df4e71ccfaf3a6d682718f1fbb2d1e6411566e139f1efa74874c303a078455f6ef72aa4dc670e9b467fdbe29d37b5c4eb526ee07b372d2bcea57871eaa05911bcb62e4ba3117ca428f93305ebf06247d573f25bb0fff22681716c21744da0f47e1a054e1ee9ac18fd711b2571c2cab26e88d1a5be46d7078723076866265880",
		"0xf851808080808080808080a08ffa88d75a03fd29af8cb1a4ac016e32ef8e39631a6bf45d79a34adfc4ecb1448080a0a1161a49c0c7e7a92a2efe173abffdbb1ed91e5235688e2edbc4e38078dc5c5580808080",
		"0xf869a02012683435c076b898a6cac1c03e41900e379104fefd4219d99f7908cb59cfb3b846f8440180a0c3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0a07498e14000b8457a51de3cd583e9337cfa52aee2c2e9f945fac35a820e685904",
	}
	accountProofRLP, err := encodeRLP(accountProof)
	ts.Require().NoError(err)

	header := Header{
		Headers:      []*ETHHeader{ethHeader},
		AccountProof: accountProofRLP,
	}
	proxy, err := NewHeader(1, &header)
	ts.Require().NoError(err)
	ts.Require().Equal(proxy.Target().Number, rawHeader.Number)
	validator, err := proxy.ValidatorSet()
	ts.Require().NoError(err)
	ts.Require().Equal(len(validator), 1)
	ts.Require().NoError(proxy.ValidateBasic())
	ts.Require().Equal(proxy.GetHeight().GetRevisionHeight(), proxy.Target().Number.Uint64())
	account, err := proxy.Account(common.HexToAddress(ibcHandlerAddress))
	ts.Require().NoError(err)
	ts.Require().Equal(account.Root, common.HexToHash("c3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0"))
	ts.Require().Equal(account.CodeHash, common.Hex2Bytes("7498e14000b8457a51de3cd583e9337cfa52aee2c2e9f945fac35a820e685904"))
	ts.Require().Equal(account.Nonce, uint64(1))
	ts.Require().Equal(account.Balance.Uint64(), uint64(0))
}

func (ts *HeaderTestSuite) TestNewHeaderError() {
	header := Header{
		Headers:      []*ETHHeader{},
		AccountProof: []byte{},
	}
	_, err := NewHeader(1, &header)
	ts.Require().Error(err)
}

func (ts *HeaderTestSuite) assertHeader(height uint64, count int) {
	ethHeaders, err := ts.reader.QueryETHHeaders(height)
	assert := ts.Require()
	assert.NoError(err)
	assert.Len(ethHeaders, count) // only one validator
	var header types.Header
	for i := 0; i < count; i++ {
		assert.NoError(rlp.DecodeBytes(ethHeaders[i].Header, &header))
		assert.Equal(header.Number.Uint64(), height+uint64(i))
	}
}

// see yui-ibc-solidity
func encodeRLP(proof []string) ([]byte, error) {
	var target [][][]byte
	for _, p := range proof {
		bz, err := hex.DecodeString(p[2:])
		if err != nil {
			return nil, err
		}
		var val [][]byte
		if err := rlp.DecodeBytes(bz, &val); err != nil {
			return nil, err
		}
		target = append(target, val)
	}
	bz, err := rlp.EncodeToBytes(target)
	if err != nil {
		return nil, err
	}
	return bz, nil
}
