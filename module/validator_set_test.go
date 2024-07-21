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

	block := previousEpochHeader()
	validators, turnTerm, err := extractValidatorSetAndTurnTerm(block)
	ts.Require().NoError(err)
	ts.Require().Len(validators, 4)
	ts.Require().Equal(turnTerm, uint8(1))

	block = epochHeader()
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
		return epochHeader(), nil
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
