package module

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func (ts *ValidatorSetTestSuite) TestSuccessExtractValidatorSet() {
	testnetHeader := &types.Header{
		// after luban in testnet:
		Number: big.NewInt(29835600),
		Extra:  common.Hex2Bytes("d883010202846765746888676f312e31392e39856c696e7578000000110bea95071284214b9b9c85549ab3d2b972df0deef66ac2c9ab1757500d6f4fdee439b17cf8e43267f94bc759162fb68de676d2fe10cc4cde26dd06be7e345e9cbf4b1dbf86b262bc35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f296c5d20b2a975c050e4220be276ace4892f4b41a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b27bb860a140cc9c8cc07d4ddf366440d9784efc88743d26af40f8956dd1c3501e560f745910bb14a5ec392f53cf78ddc2d2d69a146af287f7e079c3cbbfd3d446836d9b9397aa9a803b6c6b4f1cfc50baddbe2378cf194da35b9f4a1a32850114f1c5d9f84c8401c7414ea049d2e0876f51ce4693892331f8344a102aad88eb9e9bcfaa247cc9f898d1f8008401c7414fa0cf8d34727ff1d895bb49ca4be60c3b24d98d8afa9ce78644924e4b9aa39df8548022dc981e8703d3ca8b23fc032089667cb631cb28c32731762813bbf9fdb7e7a56b3945d65f2d72402a2abb9fbaf4bf094a3e5a542e175ecc54b426ee366b2ba200"),
	}
	validators, err := ExtractValidatorSet(testnetHeader)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 7)
	ts.Require().Equal(common.Bytes2Hex(validators[0]), "1284214b9b9c85549ab3d2b972df0deef66ac2c9ab1757500d6f4fdee439b17cf8e43267f94bc759162fb68de676d2fe10cc4cde26dd06be7e345e9cbf4b1dbf86b262bc")
	ts.Require().Equal(common.Bytes2Hex(validators[1]), "35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f2")
	ts.Require().Equal(common.Bytes2Hex(validators[2]), "96c5d20b2a975c050e4220be276ace4892f4b41a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	ts.Require().Equal(common.Bytes2Hex(validators[3]), "980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8")
	ts.Require().Equal(common.Bytes2Hex(validators[4]), "a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01b")
	ts.Require().Equal(common.Bytes2Hex(validators[5]), "b71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5f")
	ts.Require().Equal(common.Bytes2Hex(validators[6]), "f474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878")
}

func (ts *ValidatorSetTestSuite) TestErrorExtractValidatorSet() {
	testnetHeader := &types.Header{
		Number: big.NewInt(0),
		Extra:  []byte{},
	}
	_, err := ExtractValidatorSet(testnetHeader)
	ts.Require().Equal(err.Error(), "invalid extra length : 0")

	testnetHeader.Extra = make([]byte, extraSeal+extraVanity)
	_, err = ExtractValidatorSet(testnetHeader)
	ts.Require().Equal(err.Error(), "invalid validator bytes length: 0")
}

func (ts *ValidatorSetTestSuite) TestSuccessQueryValidatorSet() {

	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return &types.Header{
			Number: big.NewInt(0),
			Extra:  common.Hex2Bytes("d883010202846765746888676f312e31392e39856c696e7578000000110bea95071284214b9b9c85549ab3d2b972df0deef66ac2c9ab1757500d6f4fdee439b17cf8e43267f94bc759162fb68de676d2fe10cc4cde26dd06be7e345e9cbf4b1dbf86b262bc35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f296c5d20b2a975c050e4220be276ace4892f4b41a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b27bb860a140cc9c8cc07d4ddf366440d9784efc88743d26af40f8956dd1c3501e560f745910bb14a5ec392f53cf78ddc2d2d69a146af287f7e079c3cbbfd3d446836d9b9397aa9a803b6c6b4f1cfc50baddbe2378cf194da35b9f4a1a32850114f1c5d9f84c8401c7414ea049d2e0876f51ce4693892331f8344a102aad88eb9e9bcfaa247cc9f898d1f8008401c7414fa0cf8d34727ff1d895bb49ca4be60c3b24d98d8afa9ce78644924e4b9aa39df8548022dc981e8703d3ca8b23fc032089667cb631cb28c32731762813bbf9fdb7e7a56b3945d65f2d72402a2abb9fbaf4bf094a3e5a542e175ecc54b426ee366b2ba200"),
		}, nil
	}
	validators, err := QueryValidatorSet(fn, 200)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 7)
}

func (ts *ValidatorSetTestSuite) TestErrorQueryValidatorSet() {
	fn := func(ctx context.Context, height uint64) (*types.Header, error) {
		return nil, errors.New("error")
	}
	_, err := QueryValidatorSet(fn, 200)
	ts.Require().Equal(err.Error(), "error")
}

func (ts *ValidatorSetTestSuite) TestValidator() {
	trusted := Validators([][]byte{{1}, {2}, {3}, {4}, {5}})
	ts.Equal(trusted.Checkpoint(5), uint64(8))

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
