package module

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type VoteTestSuite struct {
	suite.Suite
}

func TestVoteTestSuite(t *testing.T) {
	suite.Run(t, new(VoteTestSuite))
}

func (ts *VoteTestSuite) SetupTest() {
}

func (ts *VoteTestSuite) TestSuccessGetVoteAttestationFromHeaderEpoch() {
	// 400
	header := epochHeader()
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(15))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0x709f88597f05218c198818991cf5598c9280db30d5bfe899da9b7a8c963bff6c"))
	ts.Require().Equal(vote.Data.SourceNumber, uint64(398))
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0x4ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d"))
	ts.Require().Equal(vote.Data.TargetNumber, uint64(399))
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "8b6dc552b410a6fa44fa31643850bcb314f1d4edb32c0c79ee3efef5397691f3685d80057d77510a00e77a39e8b2497419053c3b81a8901e85590a20a0a2dad529c82f6c175ec3ebca8a9112415aa94718af673c16c0e90e327e27709666e499")
}

func (ts *VoteTestSuite) TestSuccessGetVoteAttestationFromHeaderNotEpoch() {
	// 401
	header := epochHeaderPlus1()
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(15))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0x4ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d"))
	ts.Require().Equal(vote.Data.SourceNumber, uint64(399))
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0xe256fac4dd62cc71eaefd8d6c24ae5209c0e48f5c0b62bcced06dfa838c2ad31"))
	ts.Require().Equal(vote.Data.TargetNumber, uint64(400))
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "9338bf42b6ef715e9c887e1b285e706355c2a993cd227497b447f8aad4b7fa44d18cd895862e1a2b961b78656d620f9c015e777cf9bcb6c50e1db2783818bd91f647f6879f8bd199f266f1166f9241f00f955fb5210e7e89e7678680900d1cc1")
}

func (ts *VoteTestSuite) TestErrorGetVoteAttestationFromHeader() {
	testnetHeader := &types.Header{
		Extra: make([]byte, extraSeal+extraVanity),
	}
	vote, err := getVoteAttestationFromHeader(testnetHeader)
	ts.Require().Nil(vote)
	ts.Require().Nil(err)
}
