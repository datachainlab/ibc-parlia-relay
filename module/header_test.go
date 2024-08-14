package module

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
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
	// 400
	rawHeader := fromRlp("f90442a04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5da01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa0015ebe4a5d6cd56f0bf97db1d21746f59ab5cbecf216e34753920d815403ada2a03cd1ebc99cd975182c58de47be968c97658cff4c465e20654185f408a851403cb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201908402625a008229a884669a6ce9b90223d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274a7876ea32e7a748c697d01345145485561305b248cd0ede772633b8baea9958f9b602db36d78934d948244a13c2d66e998f987783276e9aee6facbff50b0d63574406b51b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6e04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a56701f8ae0fb8608b6dc552b410a6fa44fa31643850bcb314f1d4edb32c0c79ee3efef5397691f3685d80057d77510a00e77a39e8b2497419053c3b81a8901e85590a20a0a2dad529c82f6c175ec3ebca8a9112415aa94718af673c16c0e90e327e27709666e499f84882018ea0709f88597f05218c198818991cf5598c9280db30d5bfe899da9b7a8c963bff6c82018fa04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d8012518315e9c22a4a648f4d26efcf57f877a26498de6d53fe7a267e8d5ef01482009817fc9de90ca8008ef1f420aa606ddc0c56a975bace3906601fd5cde657d600a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
	ethHeader, err := newETHHeader(rawHeader)
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
	validator, turnLength, err := extractValidatorSetAndTurnLength(target)
	ts.Require().NoError(err)
	ts.Require().Equal(len(validator), 4)
	ts.Require().Equal(turnLength, uint8(1))
	ts.Require().NoError(header.ValidateBasic())
	ts.Require().Equal(header.GetHeight().GetRevisionHeight(), target.Number.Uint64())
	account, err := header.Account(common.HexToAddress("aa43d337145e8930d01cb4e60abf6595c692921e"))
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
