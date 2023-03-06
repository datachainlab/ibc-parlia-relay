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
		Root:   common.HexToHash("40104466ea6ef5dff68b50f4141c30d8c53127004e79ebd0df873fd7f720cce0"),
		Number: big.NewInt(200),
		Extra:  common.Hex2Bytes("d983010111846765746889676f312e31362e3135856c696e75780000d5efc574b794a294a7e455058789c82f15e8ce00669d689e911972c75ce28656b550ef4f2f7d38b22f59a639a49cb5045d83afb593a3a7df42651f4c60a9e431bb2bebd52a1d81f9d0642356c1a1244a8789f1bf3edfbb6b01"),
	}
	ethHeader, err := ts.reader.newETHHeader(&rawHeader)
	ts.Require().NoError(err)

	accountProof := []string{
		"0xf901b1a06ed7d26d8a14a9cd8aeb3f01960cae49b7e95dbd549fff6556694ea3ee3dc173a016145981b89d8e5a79ffa53ba10853658d28b40eb5b4da876fe811b637fcf9f9a06c664a574aeaa739d85b114d13a20a15330b02565910dfb142b6715b5742a5628080a0d94bd2db341ab9d3d7f4fe0aa569bb21dfac0d5eb0ec008c7af23d7f2ed98ec1a0cee66e2515872d5f4b42ada7cc733288809c11ab99aa0d25eb941236074f9904a0f3aa8d1774f013d8af0fdd8364c7833f16b42ad377e485b754f5cdae6fedaa2fa0b6abdbc03e50b225ebfd66b4191ef79b84aad79ef7ec52343228635e8cea2d15a0da497f183531645a1d4d2adfdb1f4a7fb74bf1cdcae87152c294b55d2f7c0122a0be7fe4af9ad77800739f76f16237ba9e7c66ac584076cdcc48e7279474cdc1eda043c8135071601a572445190a416fddba291e6a9014d1c4641481b8e5e791945ba05367fa3ee8fab428d91416e2feb6f6cf02ff7ba408086e937d94d2b6849c78eb80a05911bcb62e4ba3117ca428f93305ebf06247d573f25bb0fff22681716c21744da06e17e34ff3a3afaceda8cbb41f832adad3563a0f1874bd34bb2dc836f3425c1380",
		"0xf871808080a086fafbf46048a86bdf57e311afb54e741a5e47048f24c4c5203e2b449bd84e3c80808080a0786d1f63e7f001000e9585bcb37f4d8e37937df4060f122e2d909f5dac42af0e8080a0c6fc7995fea5e16130da2e3d2cb01f4dadb6158ac6a0e83081dd8a3dbd269fad8080808080",
		"0xf869a020166e3e9140c3a544e909a14f1ed65f3ba3c03b8b68be3d32aec95ed171c8b5b846f8440780a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
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
	account, err := proxy.Account(common.HexToAddress("b794a294a7e455058789c82f15e8ce00669d689e"))
	ts.Require().NoError(err)
	ts.Require().Equal(account.Root, common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"))
	ts.Require().Equal(account.CodeHash, common.Hex2Bytes("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"))
	ts.Require().Equal(account.Nonce, uint64(7))
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
		assert.Equal(header.Number.Int64(), height+uint64(i))
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
