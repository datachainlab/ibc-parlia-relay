package module

import (
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
}

func TestHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(HeaderTestSuite))
}

func (ts *HeaderTestSuite) SetupTest() {
}

func (ts *HeaderTestSuite) TestNewHeaderSuccess() {
	rawHeader := types.Header{
		Root:   common.HexToHash("c84307dfe4ccfec4a851a77755d63228d8e0b9ba3345d1eee37ed729ee16eaa1"),
		Number: big.NewInt(21400),
		Extra:  common.Hex2Bytes("d983010111846765746889676f312e31362e3135856c696e75780000d5efc574b794a294a7e455058789c82f15e8ce00669d689e2090e08cc7f42b0bd007728dea17979718bc99e2261bfe892a7ac9b849545ffc597338db43a9cb00564971a8f282cfe1741b0294ba45a21ac35168138e2c2e8200"),
	}
	ethHeader, err := newETHHeader(&rawHeader)
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
	target, err := header.Target()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number, rawHeader.Number)
	validator, err := extractValidatorSet(target)
	ts.Require().NoError(err)
	ts.Require().Equal(len(validator), 1)
	ts.Require().NoError(header.ValidateBasic())
	ts.Require().Equal(header.GetHeight().GetRevisionHeight(), target.Number.Uint64())
	account, err := header.Account(common.HexToAddress(ibcHandlerAddress))
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
	ts.Require().Error(header.ValidateBasic())
}

func (ts *HeaderTestSuite) TestExtractValidatorSetLuban() {
	testnetHeader := &types.Header{
		// after luban in testnet:
		Number: big.NewInt(29835600),
		Extra:  common.Hex2Bytes("d883010202846765746888676f312e31392e39856c696e7578000000110bea95071284214b9b9c85549ab3d2b972df0deef66ac2c9ab1757500d6f4fdee439b17cf8e43267f94bc759162fb68de676d2fe10cc4cde26dd06be7e345e9cbf4b1dbf86b262bc35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f296c5d20b2a975c050e4220be276ace4892f4b41a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b27bb860a140cc9c8cc07d4ddf366440d9784efc88743d26af40f8956dd1c3501e560f745910bb14a5ec392f53cf78ddc2d2d69a146af287f7e079c3cbbfd3d446836d9b9397aa9a803b6c6b4f1cfc50baddbe2378cf194da35b9f4a1a32850114f1c5d9f84c8401c7414ea049d2e0876f51ce4693892331f8344a102aad88eb9e9bcfaa247cc9f898d1f8008401c7414fa0cf8d34727ff1d895bb49ca4be60c3b24d98d8afa9ce78644924e4b9aa39df8548022dc981e8703d3ca8b23fc032089667cb631cb28c32731762813bbf9fdb7e7a56b3945d65f2d72402a2abb9fbaf4bf094a3e5a542e175ecc54b426ee366b2ba200"),
	}
	validators, err := extractValidatorSet(testnetHeader)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 7)
	ts.Require().Equal(common.Bytes2Hex(validators[0]), "1284214b9b9c85549ab3d2b972df0deef66ac2c9")
	ts.Require().Equal(common.Bytes2Hex(validators[1]), "35552c16704d214347f29fa77f77da6d75d7c752")
	ts.Require().Equal(common.Bytes2Hex(validators[2]), "96c5d20b2a975c050e4220be276ace4892f4b41a")
	ts.Require().Equal(common.Bytes2Hex(validators[3]), "980a75ecd1309ea12fa2ed87a8744fbfc9b863d5")
	ts.Require().Equal(common.Bytes2Hex(validators[4]), "a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0")
	ts.Require().Equal(common.Bytes2Hex(validators[5]), "b71b214cb885500844365e95cd9942c7276e7fd8")
	ts.Require().Equal(common.Bytes2Hex(validators[6]), "f474cf03cceff28abc65c9cbae594f725c80e12d")
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
