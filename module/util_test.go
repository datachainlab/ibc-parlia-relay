package module

import (
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UtilTestSuite struct {
	suite.Suite
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

func (ts *UtilTestSuite) SetupTest() {
}

func (ts *UtilTestSuite) TestGetPreviousEpoch() {
	ts.Require().Equal(constant.BlocksPerEpoch, getPreviousEpoch(2*constant.BlocksPerEpoch))
	ts.Require().Equal(uint64(0), getPreviousEpoch(2*constant.BlocksPerEpoch-1))
	ts.Require().Equal(uint64(0), getPreviousEpoch(constant.BlocksPerEpoch+1))
	ts.Require().Equal(uint64(0), getPreviousEpoch(constant.BlocksPerEpoch))
	ts.Require().Equal(uint64(0), getPreviousEpoch(constant.BlocksPerEpoch-1))
	ts.Require().Equal(uint64(0), getPreviousEpoch(0))
}
