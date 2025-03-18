package module

import (
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
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
	// 100
	header := epochHeader()
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(15))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0x302d35ff53c930401473e5650e8169f18f1156d0127cd1bd1a3e65fe365c5efc"))
	ts.Require().Equal(vote.Data.SourceNumber, header.Number.Uint64()-2)
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0x322d19e268300c0c825ffdc22a4376232406b925a7c4be8727f9a4425818ec8a"))
	ts.Require().Equal(vote.Data.TargetNumber, header.Number.Uint64()-1)
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "99f5bd42d7a4e11f283b9daa22fe1f7ae89e5260b61722ec3f57dc2b18a669e0040e42a6f81d114732c00609d4648b0d0fec3a07fb300cf49a7cf6116abb217b13a63ee6330663349e8a316c42dc04f16131445a331365d2b32bcb4d5b546c25")
}

func (ts *VoteTestSuite) TestErrorGetVoteAttestationFromHeaderEpochNoVote() {
	header := &types.Header{
		Extra:  common.Hex2Bytes("d98301040d846765746889676f312e32312e3132856c696e757800000299d9bc0808265da01e1a65d62b903c7b34c08cb389bf3d9996f763f030b1adcfb369c5a5df4a18e1529baffe7feaec66db3dbd1bc06810f7f6f88b7be6645418a7e2a2a3f40514c215a13e315cbfb9398a26d77a299963bf034c28f8b0183ea044211f468630233d2533b73307979c78a9486b33bb4ee04ca31a65f3e86fba804db7fe293fa643e6b72bb3821a3d9d7a717d64e6088ac937d5aacdd3e20ca963979974cd8ff90cbf097023dc8c448245ceff671e965d57d82eaf9be91478cfa0f24d2993e0c5f43a6c5a4cd99850023040d3256eb0babe89f0ea54edaa398513136612f5a334b49d766ebe3eb9f6bdc163bd2c19aa7e8cee1667851ae0c1651f01c4cf7cf2cfcf8475bff3e99cab25b05631472d53387f3321fd69d1e030bb921230dfb188826affaa39ebf1c38b190851e4db0588a3e90142c5299041fb8a0db3bb9a1fa4bdf0dae84ca37ee12a6b8c26caab775f0e007b76d76ee8823de52a1a431884c2ca930c5e72bff3803af79641cf964cc001671017f0b680f93b7dde085b24bbc67b2a562a216f903ac878c5477641328172a353f1e493cf7f5f2cf1aec83bf0c74df566a41aa7ed65ea84ea99e3849ef31887c0f880a0feb92f356f58fbd023a82f5311fc87a5883a662e9ebbbefc90bf13aa533c2438a4113804bfd447b49cd040d20bc21e49ffea6487f5638e4346ad9fc6d1ec30e28016d3892b51a7898bd354cfe78643453fd3868410da412de7f2883180d0a2840111ad2e043fa403eb04cc3c0ed356ea54a6e7015490240681b002cb63e12f65c456cafca335c730b123553e70df5322013812429e0bc31508e1f1fbf0ab312e4aaade9e022150071a1f00"),
		Number: big.NewInt(0).SetUint64(constant.BlocksPerEpoch),
	}
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Nil(vote)
}

func (ts *VoteTestSuite) TestSuccessGetVoteAttestationFromHeaderNotEpoch() {
	header := epochHeaderPlus1()
	vote, err := getVoteAttestationFromHeader(header)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(15))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0x322d19e268300c0c825ffdc22a4376232406b925a7c4be8727f9a4425818ec8a"))
	ts.Require().Equal(vote.Data.SourceNumber, header.Number.Uint64()-2)
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0x3a302bedfa30dd88b82a95136a99d93ea8863a741c2201ad77a63d0f9c0c329c"))
	ts.Require().Equal(vote.Data.TargetNumber, header.Number.Uint64()-1)
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "a51854c31fb60a02ba70c07eeb467be677b9548c828607f99dfd0edc80a9b25be05670b86485dd71d8fb8e19d7458a9103d942ea6b84070ed47adcd3a3f284385fc538a5f692289c3abc25372e461a54ef23100718aedf80224a1e4fe26671d3")
}

func (ts *VoteTestSuite) TestErrorGetVoteAttestationFromHeader() {
	testnetHeader := &types.Header{
		Extra: make([]byte, extraSeal+extraVanity),
	}
	vote, err := getVoteAttestationFromHeader(testnetHeader)
	ts.Require().Nil(vote)
	ts.Require().Nil(err)
}
