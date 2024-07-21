package module

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type ValidatorSetTestSuite struct {
	suite.Suite
}

func TestValidatorSetTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorSetTestSuite))
}

func (ts *ValidatorSetTestSuite) SetupTest() {
}

func (ts *ValidatorSetTestSuite) fromRlp(hex string) *types.Header {
	var h types.Header
	err := rlp.DecodeBytes(common.Hex2Bytes(hex), &h)
	ts.Require().NoError(err)
	return &h
}

func (ts *ValidatorSetTestSuite) TestSuccessExtractValidatorSet() {

	// 200
	block := ts.fromRlp("f90391a0844dee9abff97d261ae0049fe38246ac10aba49f2b8618f28f7c2d19e62eccf9a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a0ecf1aa30fa754576ac4abc3cf2a61d1babd41c7e5515855efd857b2d3f37866ba00f0ea7d212c4aaca329b03f5e9ed9c69d3641eb5e03a4edb69b61e6f9d8d51efa0c3372a1f332fc4245e1a9fdcb62580fc6dae741087a8029560f19216dd3d58b9b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000281c88402625a00826c7484669a6a91b90173d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6d9a13701eafb76870cb220843b8c6476824bfa158c66a3f3d2fba1d440da8edc79b59ed9a3a43db62bd7659f7d4e25073f9241dba560600b23e26c30d48ea0395eeeb4ede04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a567015cbc63c3d778cef5e8dbfdaf1fd8f758a764f2667aad2d3775954e4ac23e726226b66f0a94631bd0b6d937b22955d73eed65a31a6f535662f51cc7547143f6f201a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
	validators, turnTerm, err := extractValidatorSetAndTurnTerm(block)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 4)
	ts.Require().Equal(turnTerm, uint8(1))

	// 400
	block = ts.fromRlp("f90442a04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5da01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa0015ebe4a5d6cd56f0bf97db1d21746f59ab5cbecf216e34753920d815403ada2a03cd1ebc99cd975182c58de47be968c97658cff4c465e20654185f408a851403cb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201908402625a008229a884669a6ce9b90223d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274a7876ea32e7a748c697d01345145485561305b248cd0ede772633b8baea9958f9b602db36d78934d948244a13c2d66e998f987783276e9aee6facbff50b0d63574406b51b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6e04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a56701f8ae0fb8608b6dc552b410a6fa44fa31643850bcb314f1d4edb32c0c79ee3efef5397691f3685d80057d77510a00e77a39e8b2497419053c3b81a8901e85590a20a0a2dad529c82f6c175ec3ebca8a9112415aa94718af673c16c0e90e327e27709666e499f84882018ea0709f88597f05218c198818991cf5598c9280db30d5bfe899da9b7a8c963bff6c82018fa04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d8012518315e9c22a4a648f4d26efcf57f877a26498de6d53fe7a267e8d5ef01482009817fc9de90ca8008ef1f420aa606ddc0c56a975bace3906601fd5cde657d600a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
	validators, turnTerm, err = extractValidatorSetAndTurnTerm(block)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 4)
	ts.Require().Equal(turnTerm, uint8(1))

}

func (ts *ValidatorSetTestSuite) TestErrorExtractValidatorSet() {
	testnetHeader := &types.Header{
		Number: big.NewInt(0),
		Extra:  []byte{},
	}
	_, _, err := ExtractValidatorSetAndTurnTerm(testnetHeader)
	ts.Require().Equal(err.Error(), "invalid extra length : 0")

	testnetHeader.Extra = make([]byte, extraSeal+extraVanity)
	_, _, err = ExtractValidatorSetAndTurnTerm(testnetHeader)
	ts.Require().Equal(err.Error(), "invalid validator bytes length: 0")
}

func (ts *ValidatorSetTestSuite) TestSuccessQueryValidatorSet() {

	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return ts.fromRlp("f90442a04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5da01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa0015ebe4a5d6cd56f0bf97db1d21746f59ab5cbecf216e34753920d815403ada2a03cd1ebc99cd975182c58de47be968c97658cff4c465e20654185f408a851403cb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201908402625a008229a884669a6ce9b90223d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274a7876ea32e7a748c697d01345145485561305b248cd0ede772633b8baea9958f9b602db36d78934d948244a13c2d66e998f987783276e9aee6facbff50b0d63574406b51b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6e04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a56701f8ae0fb8608b6dc552b410a6fa44fa31643850bcb314f1d4edb32c0c79ee3efef5397691f3685d80057d77510a00e77a39e8b2497419053c3b81a8901e85590a20a0a2dad529c82f6c175ec3ebca8a9112415aa94718af673c16c0e90e327e27709666e499f84882018ea0709f88597f05218c198818991cf5598c9280db30d5bfe899da9b7a8c963bff6c82018fa04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d8012518315e9c22a4a648f4d26efcf57f877a26498de6d53fe7a267e8d5ef01482009817fc9de90ca8008ef1f420aa606ddc0c56a975bace3906601fd5cde657d600a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080"), nil
	}
	validators, turnTerm, err := QueryValidatorSetAndTurnTerm(fn, 400)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 4)
	ts.Require().Equal(turnTerm, uint8(1))
}

func (ts *ValidatorSetTestSuite) TestErrorQueryValidatorSet() {
	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return nil, errors.New("error")
	}
	_, _, err := QueryValidatorSetAndTurnTerm(fn, 200)
	ts.Require().Equal(err.Error(), "error")
}

func (ts *ValidatorSetTestSuite) TestCheckpoint() {
	validator := Validators(make([][]byte, 1))
	ts.Equal(int(validator.Checkpoint(1)), 1)
	ts.Equal(int(validator.Checkpoint(3)), 1)
	ts.Equal(int(validator.Checkpoint(9)), 1)

	validator = make([][]byte, 5)
	ts.Equal(int(validator.Checkpoint(1)), 3)
	ts.Equal(int(validator.Checkpoint(3)), 7)
	ts.Equal(int(validator.Checkpoint(9)), 19)

	validator = make([][]byte, 8)
	ts.Equal(int(validator.Checkpoint(1)), 5)
	ts.Equal(int(validator.Checkpoint(3)), 13)
	ts.Equal(int(validator.Checkpoint(9)), 37)

	validator = make([][]byte, 21)
	ts.Equal(int(validator.Checkpoint(1)), 11)
	ts.Equal(int(validator.Checkpoint(3)), 31)
	ts.Equal(int(validator.Checkpoint(9)), 91)
}

func (ts *ValidatorSetTestSuite) TestValidator() {
	trusted := Validators([][]byte{{1}, {2}, {3}, {4}, {5}})
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}, {4}, {5}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}, {4}, {5}, {10}, {11}, {12}, {13}, {14}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}, {4}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}, {4}, {10}, {11}, {12}, {13}, {14}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {3}, {10}, {11}, {12}, {13}, {14}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}}))
	ts.True(trusted.Contains([][]byte{{1}, {2}, {10}, {11}, {12}, {13}, {14}}))
	ts.False(trusted.Contains([][]byte{{1}}))
	ts.False(trusted.Contains([][]byte{{1}, {10}, {11}, {12}, {13}, {14}}))
	ts.False(trusted.Contains([][]byte{}))
	ts.False(trusted.Contains([][]byte{{10}, {11}, {12}, {13}, {14}}))

}
