package module

import (
	"context"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/stretchr/testify/suite"
	"math/big"
	"strings"
	"testing"
)

type SetupTestSuite struct {
	suite.Suite
}

func TestSetupTestSuite(t *testing.T) {
	suite.Run(t, new(SetupTestSuite))
}

func (ts *SetupTestSuite) SetupTest() {
	err := log.InitLogger("DEBUG", "text", "stdout")
	ts.Require().NoError(err)
}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_neighboringEpoch() {

	verify := func(latestHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		neighborFn := func(height uint64, _ uint64) (core.Header, error) {
			h, e := newETHHeader(&types2.Header{
				Number: big.NewInt(int64(height)),
			})
			return &Header{
				Headers: []*ETHHeader{h},
			}, e
		}
		nonNeighborFn := func(height uint64, _ uint64, _ uint64) (core.Header, error) {
			return nil, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300"),
			}, nil
		}

		targets, err := setupHeadersForUpdate(neighborFn, nonNeighborFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 100000))
		ts.Require().NoError(err)
		ts.Require().Len(targets, expected)
		for i, h := range targets {
			trusted := h.(*Header).TrustedHeight
			if i == 0 {
				ts.Require().Equal(trusted.RevisionHeight, latestHeight)
			} else {
				ts.Require().Equal(*trusted, targets[i-1].GetHeight())
			}
		}
	}

	verify(0, constant.BlocksPerEpoch-1, 1)
	verify(0, constant.BlocksPerEpoch, 1)
	verify(0, constant.BlocksPerEpoch+1, 2)
	verify(0, 10*constant.BlocksPerEpoch-1, 10)
	verify(0, 10*constant.BlocksPerEpoch, 10)
	verify(0, 10*constant.BlocksPerEpoch+1, 11)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, 1)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, 2)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, 10)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, 10)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, 11)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, 1)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, 9)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, 9)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, 10)
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, 9)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, 9)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, 10)

}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_nonNeighboringEpoch() {

	verify := func(latestHeight, nextHeight uint64, expectedEpochs []int64, enableFast bool) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		queryVerifyingNeighboringEpochHeader := func(_ uint64, _ uint64) (core.Header, error) {
			return nil, nil
		}
		t := true
		canVerify := &enableFast
		queryVerifyingNonNeighboringEpochHeader := func(height, limit, checkpoint uint64) (core.Header, error) {
			if !*canVerify {
				canVerify = &t
				return nil, nil
			}
			hs := make([]*ETHHeader, 0)
			for i := height; i <= checkpoint+2; i++ {
				if i > limit {
					break
				}
				h, e := newETHHeader(&types2.Header{
					Number: big.NewInt(int64(i)),
				})
				ts.Require().NoError(e)
				hs = append(hs, h)
			}
			return &Header{
				Headers: hs,
			}, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300"),
			}, nil
		}

		targets, err := setupHeadersForUpdate(queryVerifyingNeighboringEpochHeader, queryVerifyingNonNeighboringEpochHeader, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 1000000))
		ts.Require().NoError(err)
		ts.Require().Len(targets, len(expectedEpochs))
		for i, h := range targets {
			targetHeader, _ := h.(*Header).Target()
			ts.Require().Equal(targetHeader.Number.Int64(), expectedEpochs[i])

			trusted := h.(*Header).TrustedHeight
			if i == 0 {
				ts.Require().Equal(trusted.RevisionHeight, latestHeight)
			} else {
				ts.Require().Equal(*trusted, targets[i-1].GetHeight())
			}
		}
	}

	verify(0, constant.BlocksPerEpoch-1, []int64{int64(constant.BlocksPerEpoch - 1)}, false)
	verify(0, constant.BlocksPerEpoch-1, []int64{int64(constant.BlocksPerEpoch - 1)}, true)
	verify(0, constant.BlocksPerEpoch, []int64{}, false)
	verify(0, constant.BlocksPerEpoch, []int64{}, true)
	verify(0, constant.BlocksPerEpoch+1, []int64{}, false)
	verify(0, constant.BlocksPerEpoch+1, []int64{}, true)
	verify(0, 10*constant.BlocksPerEpoch-1, []int64{400, 800, 1200, 1600}, false)
	verify(0, 10*constant.BlocksPerEpoch-1, []int64{1800, 1999}, true)
	verify(0, 10*constant.BlocksPerEpoch, []int64{400, 800, 1200, 1600, 2000}, false)
	verify(0, 10*constant.BlocksPerEpoch, []int64{2000}, true)
	verify(0, 10*constant.BlocksPerEpoch+1, []int64{400, 800, 1200, 1600, 2000, 2001}, false)
	verify(0, 10*constant.BlocksPerEpoch+1, []int64{2000, 2001}, true)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, []int64{}, false)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, []int64{}, true)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, []int64{}, false)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, []int64{}, true)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, []int64{}, false)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, []int64{}, true)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, []int64{400, 800, 1200, 1600}, false)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, []int64{1800, 1999}, true)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, []int64{400, 800, 1200, 1600, 2000}, false)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, []int64{2000}, true)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, []int64{400, 800, 1200, 1600, 2000, 2001}, false)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, []int64{2000, 2001}, true)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, []int64{}, false)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, []int64{}, true)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, []int64{int64(constant.BlocksPerEpoch + 1)}, false)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, []int64{int64(constant.BlocksPerEpoch + 1)}, true)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, []int64{600, 1000, 1400, 1800, 1999}, false)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, []int64{1800, 1999}, true)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, []int64{600, 1000, 1400, 1800}, false)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, []int64{2000}, true)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, []int64{600, 1000, 1400, 1800}, false)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, []int64{2000, 2001}, true)
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, []int64{}, false)
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, []int64{}, true)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, []int64{600, 1000, 1400, 1800, 1999}, false)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, []int64{1800, 1999}, true)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, []int64{600, 1000, 1400, 1800}, false)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, []int64{2000}, true)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, []int64{600, 1000, 1400, 1800}, false)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, []int64{2000, 2001}, true)

}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_nonNeighboringEpoch_checkpointOverLimit() {

	verify := func(latestHeight, nextHeight uint64, expectedEpochs []int64) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers:            []*ETHHeader{target},
			CurrentValidators:  [][]byte{{1}},
			PreviousValidators: [][]byte{{1}},
		}
		queryVerifyingNeighboringEpochHeader := func(_ uint64, _ uint64) (core.Header, error) {
			return nil, nil
		}
		queryVerifyingNonNeighboringEpochHeader := func(height, limit, checkpoint uint64) (core.Header, error) {
			hs := make([]*ETHHeader, 0)
			for i := height; i <= checkpoint+2; i++ {
				if i > limit {
					break
				}
				h, e := newETHHeader(&types2.Header{
					Number: big.NewInt(int64(i)),
				})
				ts.Require().NoError(e)
				hs = append(hs, h)
			}
			return &Header{
				Headers: hs,
			}, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300"),
			}, nil
		}

		targets, err := setupHeadersForUpdate(queryVerifyingNeighboringEpochHeader, queryVerifyingNonNeighboringEpochHeader, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 1))
		ts.Require().NoError(err)
		ts.Require().Len(targets, len(expectedEpochs))
		for i, h := range targets {
			targetHeader, _ := h.(*Header).Target()
			ts.Require().Equal(targetHeader.Number.Int64(), expectedEpochs[i])

			trusted := h.(*Header).TrustedHeight
			if i == 0 {
				ts.Require().Equal(trusted.RevisionHeight, latestHeight)
			} else {
				ts.Require().Equal(*trusted, targets[i-1].GetHeight())
			}
		}
	}

	verify(0, constant.BlocksPerEpoch-1, []int64{int64(constant.BlocksPerEpoch - 1)})
	verify(0, constant.BlocksPerEpoch, []int64{})
	verify(0, constant.BlocksPerEpoch+1, []int64{})
	verify(0, 10*constant.BlocksPerEpoch-1, []int64{})
	verify(0, 10*constant.BlocksPerEpoch, []int64{})
	verify(0, 10*constant.BlocksPerEpoch+1, []int64{})
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, []int64{})
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, []int64{})
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, []int64{})
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, []int64{})
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, []int64{})
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, []int64{})
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, []int64{})
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, []int64{int64(constant.BlocksPerEpoch + 1)})
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, []int64{})
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, []int64{})
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, []int64{})
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, []int64{})
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, []int64{})
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, []int64{})
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, []int64{})

}

func (ts *SetupTestSuite) TestSuccess_setupHeadersForUpdate_allEmpty() {

	verify := func(latestHeight, nextHeight uint64, expected int) {
		clientStateLatestHeight := clienttypes.NewHeight(0, latestHeight)
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(nextHeight)),
		})
		ts.Require().NoError(err)
		latestFinalizedHeader := &Header{
			Headers: []*ETHHeader{target},
		}
		neighboringEpochFn := func(_ uint64, _ uint64) (core.Header, error) {
			return nil, nil
		}
		nonNeighboringEpochFn := func(_ uint64, _, _ uint64) (core.Header, error) {
			return nil, nil
		}
		headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
			return &types2.Header{
				Number: big.NewInt(int64(height)),
				Extra:  common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300"),
			}, nil
		}
		targets, err := setupHeadersForUpdate(neighboringEpochFn, nonNeighboringEpochFn, headerFn, clientStateLatestHeight, latestFinalizedHeader, clienttypes.NewHeight(0, 1000000))
		ts.Require().NoError(err)
		ts.Require().Len(targets, expected)
	}

	verify(0, constant.BlocksPerEpoch-1, 1) // latest non epoch finalized header exists
	verify(0, constant.BlocksPerEpoch, 0)
	verify(0, constant.BlocksPerEpoch+1, 0)
	verify(0, 10*constant.BlocksPerEpoch-1, 0)
	verify(0, 10*constant.BlocksPerEpoch, 0)
	verify(0, 10*constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch-1, constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch-1, 10*constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch, constant.BlocksPerEpoch+1, 1) // latest non epoch finalized header exists
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch, 10*constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch+1, constant.BlocksPerEpoch+1, 0)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch-1, 0)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch, 0)
	verify(constant.BlocksPerEpoch+1, 10*constant.BlocksPerEpoch+1, 0)
}

func (ts *SetupTestSuite) TestSuccess_setupNeighboringEpochHeader_notContainTrusted() {

	epochHeight := uint64(400)
	trustedEpochHeight := uint64(200)

	neighboringEpochFn := func(height uint64, limit uint64) (core.Header, error) {
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(limit)),
		})
		ts.Require().NoError(err)
		return &Header{
			Headers: []*ETHHeader{target},
		}, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		extra := common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300")
		if height == trustedEpochHeight {
			// testnet validator (size = 6)
			extra = common.Hex2Bytes("d983010306846765746889676f312e32302e3131856c696e7578000053474aa9061284214b9b9c85549ab3d2b972df0deef66ac2c98e82934ca974fdcd97f3309de967d3c9c43fa711a8d673af5d75465844bf8969c8d1948d903748ac7b8b1720fa64e50c35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f2980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b23fb860a2e980e217c681cb143f623fd5f9f621f6ff6744aef8e8eac63c68750700d0fc90e764516a9eaf069dae86e8f9db5c32037e33b610b88e180abe6c7cb44fe7291bbbf502d4a93b45b19214a6135d5b043c74d9b040969eb8a0ed038f3283173ff84c840231b4dea02c4e09f3cb878f41d53efbb15dcf08cab13455b16b9bbcfc7bd4e35de0a63e17840231b4dfa0c0931c8edab5ab5979a0762d3516367c2f44cecb5070db0ff7a4af46fc5073ee80672779b27abd01b3cff88b18684bbdc4a42009d1c8335e309dde3049c1ab0b0f457f75667781c9380994a6a92bb47c216f986252b2a8e82874307243c15e7f1b01")
		}
		return &types2.Header{
			Number: big.NewInt(int64(height)),
			Extra:  extra,
		}, nil
	}
	hs, err := setupNeighboringEpochHeader(headerFn, neighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 1000000))
	ts.Require().NoError(err)
	target, err := hs.(*Header).Target()
	ts.Require().NoError(err)

	// checkpoint - 1
	ts.Require().Equal(int64(403), target.Number.Int64())
}

func (ts *SetupTestSuite) TestSuccess_setupNeighboringEpochHeader_containTrusted() {

	epochHeight := uint64(400)
	trustedEpochHeight := uint64(200)

	neighboringEpochFn := func(height uint64, limit uint64) (core.Header, error) {
		target, err := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(limit)),
		})
		ts.Require().NoError(err)
		return &Header{
			Headers: []*ETHHeader{target},
		}, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		extra := common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300")
		return &types2.Header{
			Number: big.NewInt(int64(height)),
			Extra:  extra,
		}, nil
	}
	hs, err := setupNeighboringEpochHeader(headerFn, neighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 1000000))
	ts.Require().NoError(err)
	target, err := hs.(*Header).Target()
	ts.Require().NoError(err)

	// next checkpoint - 1
	ts.Require().Equal(int64(610), target.Number.Int64())
}

func (ts *SetupTestSuite) TestSuccess_setupNonNeighboringEpochHeader_containTrusted() {

	epochHeight := uint64(600)
	trustedEpochHeight := uint64(200)

	nonNeighboringEpochFn := func(height uint64, limit uint64, checkpoint uint64) (core.Header, error) {
		h1, e := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(checkpoint)),
		})
		ts.Require().NoError(e)
		h2, e := newETHHeader(&types2.Header{
			Number: big.NewInt(int64(limit)),
		})
		ts.Require().NoError(e)
		return &Header{
			Headers: []*ETHHeader{h1, h2},
		}, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		extra := common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300")
		return &types2.Header{
			Number: big.NewInt(int64(height)),
			Extra:  extra,
		}, nil
	}
	hs, err := setupNonNeighboringEpochHeader(headerFn, nonNeighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 1000000))
	ts.Require().NoError(err)
	target, err := hs.(*Header).Target()
	ts.Require().NoError(err)
	last, err := hs.(*Header).Last()
	ts.Require().NoError(err)

	// checkpoint
	ts.Require().Equal(21, len(hs.(*Header).TrustedCurrentValidators))
	ts.Require().Equal(int64(611), target.Number.Int64())
	// next checkpoint
	ts.Require().Equal(int64(810), last.Number.Int64())
}

func (ts *SetupTestSuite) TestSuccess_setupNonNeighboringEpochHeader_containTrusted_notFinalized() {

	epochHeight := uint64(600)
	trustedEpochHeight := uint64(200)

	nonNeighboringEpochFn := func(height uint64, limit uint64, checkpoint uint64) (core.Header, error) {
		return nil, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		extra := common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300")
		return &types2.Header{
			Number: big.NewInt(int64(height)),
			Extra:  extra,
		}, nil
	}
	hs, err := setupNonNeighboringEpochHeader(headerFn, nonNeighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 1))
	ts.Require().NoError(err)
	ts.Require().Nil(hs)
}

func (ts *SetupTestSuite) TestError_setupNonNeighboringEpochHeader_notContainTrusted() {

	epochHeight := uint64(600)
	trustedEpochHeight := uint64(200)

	nonNeighboringEpochFn := func(height uint64, limit uint64, checkpoint uint64) (core.Header, error) {
		return nil, nil
	}
	headerFn := func(_ context.Context, height uint64) (*types2.Header, error) {
		extra := common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300")
		if height == trustedEpochHeight {
			// testnet validator (size = 6)
			extra = common.Hex2Bytes("d983010306846765746889676f312e32302e3131856c696e7578000053474aa9061284214b9b9c85549ab3d2b972df0deef66ac2c98e82934ca974fdcd97f3309de967d3c9c43fa711a8d673af5d75465844bf8969c8d1948d903748ac7b8b1720fa64e50c35552c16704d214347f29fa77f77da6d75d7c752b742ad4855bae330426b823e742da31f816cc83bc16d69a9134be0cfb4a1d17ec34f1b5b32d5c20440b8536b1e88f0f2980a75ecd1309ea12fa2ed87a8744fbfc9b863d589037a9ace3b590165ea1c0c5ac72bf600b7c88c1e435f41932c1132aae1bfa0bb68e46b96ccb12c3415e4d82af717d8a2959d3f95eae5dc7d70144ce1b73b403b7eb6e0b973c2d38487e58fd6e145491b110080fb14ac915a0411fc78f19e09a399ddee0d20c63a75d8f930f1694544ad2dc01bb71b214cb885500844365e95cd9942c7276e7fd8a2750ec6dded3dcdc2f351782310b0eadc077db59abca0f0cd26776e2e7acb9f3bce40b1fa5221fd1561226c6263cc5ff474cf03cceff28abc65c9cbae594f725c80e12d96c9b86c3400e529bfe184056e257c07940bb664636f689e8d2027c834681f8f878b73445261034e946bb2d901b4b878f8b23fb860a2e980e217c681cb143f623fd5f9f621f6ff6744aef8e8eac63c68750700d0fc90e764516a9eaf069dae86e8f9db5c32037e33b610b88e180abe6c7cb44fe7291bbbf502d4a93b45b19214a6135d5b043c74d9b040969eb8a0ed038f3283173ff84c840231b4dea02c4e09f3cb878f41d53efbb15dcf08cab13455b16b9bbcfc7bd4e35de0a63e17840231b4dfa0c0931c8edab5ab5979a0762d3516367c2f44cecb5070db0ff7a4af46fc5073ee80672779b27abd01b3cff88b18684bbdc4a42009d1c8335e309dde3049c1ab0b0f457f75667781c9380994a6a92bb47c216f986252b2a8e82874307243c15e7f1b01")
		}
		return &types2.Header{
			Number: big.NewInt(int64(height)),
			Extra:  extra,
		}, nil
	}
	_, err := setupNonNeighboringEpochHeader(headerFn, nonNeighboringEpochFn, epochHeight, trustedEpochHeight, clienttypes.NewHeight(0, 1000000))
	ts.Require().True(strings.Contains(err.Error(), "invalid untrusted validator set"))
}
