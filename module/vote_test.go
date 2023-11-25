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
	testnetHeader := &types.Header{
		Number: big.NewInt(29835600),
		Extra:  common.Hex2Bytes("d883010202846765746888676f312e31392e39856c696e7578000000110bea95071284214b9b9c85549ab3d2b972df0deef66ac2c9ab1757500d6f4fdee439b17cf8e43267f94bc759162fb68de676d2fe10cc4cde26dd06be7e345e9cbf4b1dbf86b262bc35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f296c5d20b2a975c050e4220be276ace4892f4b41a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b27bb860a140cc9c8cc07d4ddf366440d9784efc88743d26af40f8956dd1c3501e560f745910bb14a5ec392f53cf78ddc2d2d69a146af287f7e079c3cbbfd3d446836d9b9397aa9a803b6c6b4f1cfc50baddbe2378cf194da35b9f4a1a32850114f1c5d9f84c8401c7414ea049d2e0876f51ce4693892331f8344a102aad88eb9e9bcfaa247cc9f898d1f8008401c7414fa0cf8d34727ff1d895bb49ca4be60c3b24d98d8afa9ce78644924e4b9aa39df8548022dc981e8703d3ca8b23fc032089667cb631cb28c32731762813bbf9fdb7e7a56b3945d65f2d72402a2abb9fbaf4bf094a3e5a542e175ecc54b426ee366b2ba200"),
	}
	vote, err := getVoteAttestationFromHeader(testnetHeader)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(123))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0x49d2e0876f51ce4693892331f8344a102aad88eb9e9bcfaa247cc9f898d1f800"))
	ts.Require().Equal(vote.Data.SourceNumber, uint64(29835598))
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0xcf8d34727ff1d895bb49ca4be60c3b24d98d8afa9ce78644924e4b9aa39df854"))
	ts.Require().Equal(vote.Data.TargetNumber, uint64(29835599))
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "a140cc9c8cc07d4ddf366440d9784efc88743d26af40f8956dd1c3501e560f745910bb14a5ec392f53cf78ddc2d2d69a146af287f7e079c3cbbfd3d446836d9b9397aa9a803b6c6b4f1cfc50baddbe2378cf194da35b9f4a1a32850114f1c5d9")
}

func (ts *VoteTestSuite) TestSuccessGetVoteAttestationFromHeaderNotEpoch() {
	testnetHeader := &types.Header{
		Number: big.NewInt(31835601),
		Extra:  common.Hex2Bytes("d88301020b846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5831defffb860a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37f84c8401e5c5cfa0be938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c14648401e5c5d0a03658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc28027cb6d065d5a6d8749ca45a185add61b9ce470136898643170f8072513ca45f35d826f02cb2494f857beebdac9ec04196c8b30a65352ef155a28ac6a0057ff1601"),
	}
	vote, err := getVoteAttestationFromHeader(testnetHeader)
	ts.Require().NoError(err)
	ts.Require().Equal(vote.VoteAddressSet, uint64(1961983))
	ts.Require().Equal(vote.Data.SourceHash, common.HexToHash("0xbe938dfeafe5b932c2dcef0e2bebb1a05f31104a59b49d78b0b7746a483c1464"))
	ts.Require().Equal(vote.Data.SourceNumber, uint64(31835599))
	ts.Require().Equal(vote.Data.TargetHash, common.HexToHash("0x3658f0bb6692995a9dd3b72a69ec6e8e1b9af4361718d8a275c2b92d26eeffc2"))
	ts.Require().Equal(vote.Data.TargetNumber, uint64(31835600))
	ts.Require().Equal(common.Bytes2Hex(vote.AggSignature[:]), "a44482b16993815ff4903016ce83ef788b455e2c80ba9976e8e55ac6591b9f9965234a0a2c579269bc5e09577977322d07d17bb8d657ac621a1abfadcb35b9c9d4713dbdd3d47fd3cc6dc2475c989aa224fecd083101049ef1adea2718b00e37")
}

func (ts *VoteTestSuite) TestErrorGetVoteAttestationFromHeader() {
	testnetHeader := &types.Header{
		Extra: make([]byte, extraSeal+extraVanity),
	}
	vote, err := getVoteAttestationFromHeader(testnetHeader)
	ts.Require().Nil(vote)
	ts.Require().Nil(err)
}
