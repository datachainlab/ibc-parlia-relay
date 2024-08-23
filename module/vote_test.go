package module

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"math/big"
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

func (ts *VoteTestSuite) TestErrorGetVoteAttestationFromHeaderEpochNoVote() {
	header := &types.Header{
		Extra:  common.Hex2Bytes("d98301040d846765746889676f312e32312e3132856c696e757800000299d9bc0808265da01e1a65d62b903c7b34c08cb389bf3d9996f763f030b1adcfb369c5a5df4a18e1529baffe7feaec66db3dbd1bc06810f7f6f88b7be6645418a7e2a2a3f40514c215a13e315cbfb9398a26d77a299963bf034c28f8b0183ea044211f468630233d2533b73307979c78a9486b33bb4ee04ca31a65f3e86fba804db7fe293fa643e6b72bb3821a3d9d7a717d64e6088ac937d5aacdd3e20ca963979974cd8ff90cbf097023dc8c448245ceff671e965d57d82eaf9be91478cfa0f24d2993e0c5f43a6c5a4cd99850023040d3256eb0babe89f0ea54edaa398513136612f5a334b49d766ebe3eb9f6bdc163bd2c19aa7e8cee1667851ae0c1651f01c4cf7cf2cfcf8475bff3e99cab25b05631472d53387f3321fd69d1e030bb921230dfb188826affaa39ebf1c38b190851e4db0588a3e90142c5299041fb8a0db3bb9a1fa4bdf0dae84ca37ee12a6b8c26caab775f0e007b76d76ee8823de52a1a431884c2ca930c5e72bff3803af79641cf964cc001671017f0b680f93b7dde085b24bbc67b2a562a216f903ac878c5477641328172a353f1e493cf7f5f2cf1aec83bf0c74df566a41aa7ed65ea84ea99e3849ef31887c0f880a0feb92f356f58fbd023a82f5311fc87a5883a662e9ebbbefc90bf13aa533c2438a4113804bfd447b49cd040d20bc21e49ffea6487f5638e4346ad9fc6d1ec30e28016d3892b51a7898bd354cfe78643453fd3868410da412de7f2883180d0a2840111ad2e043fa403eb04cc3c0ed356ea54a6e7015490240681b002cb63e12f65c456cafca335c730b123553e70df5322013812429e0bc31508e1f1fbf0ab312e4aaade9e022150071a1f00"),
		Number: big.NewInt(43198800),
	}
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Nil(vote)
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
