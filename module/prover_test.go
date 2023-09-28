package module

import (
	"context"
	"math/big"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	types3 "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/stretchr/testify/suite"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"

	// contract address changes for each deployment
	ibcHandlerAddress = "aa43d337145e8930d01cb4e60abf6595c692921e"
)

type mockChain struct {
	Chain
	latestHeight uint64
	chainID      uint64
}

func (c *mockChain) CanonicalChainID(ctx context.Context) (uint64, error) {
	return c.chainID, nil
}

func (c *mockChain) QueryClientState(ctx core.QueryContext) (*clienttypes.QueryClientStateResponse, error) {
	cHeight := clienttypes.NewHeight(ctx.Height().GetRevisionNumber(), ctx.Height().GetRevisionHeight())
	cs := ClientState{
		LatestHeight:    &cHeight,
		IbcStoreAddress: common.Hex2Bytes(ibcHandlerAddress),
	}
	anyClientState, err := codectypes.NewAnyWithValue(&cs)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryClientStateResponse(anyClientState, nil, cHeight), nil
}

func (c *mockChain) Header(_ context.Context, height uint64) (*types2.Header, error) {
	headerMap := map[uint64]types2.Header{
		31297221: {
			ParentHash:  common.HexToHash("da01d8fede81f2840ea4bd2d5586d4303f68c73becf0a23d1d6dddd4890bd274"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("ea0a6e3c511bbd10f4519ece37dc24887e11b55d"),
			Root:        common.HexToHash("97528e3167d013309f8585c8db78f4af02b0acbcf92e6b3aac530f9972b18ded"),
			TxHash:      common.HexToHash("ed65d042cdbf107e3803eb2aaa1932b1ae28ab8b2fc4df9bafb9217b390aa9c9"),
			ReceiptHash: common.HexToHash("b8394040a65867ce80c2f4fd34d3bba3095b80ff9ccaea1e58669ca564a1e103"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("91a876410d0c7012c91502d6845200a6d208a629cdcc06a0053861b2b9760ba2c8293640420712c5329374810216e098cfeba05d250267e0d8421025952c6a30242ee00309c60244e5338ccbca3104be2518206551e6580b0c5d81b0c8a33c05e02cb13e7b22c3cb248d041493691c51087c8acb1b0bb68e884ec552e443b4041e0b07914679c2eb02f4240458036fae4aa6fe9dd421313b75513b74aa505c30067648177ea9d9c2332d5deebe9f38d046eac9840301613312c0252a7813aaf8ee39440e3088e446cc5d0a9e1ca2fb00a0c8b1421d1dc03611ba406a2a34e5fd16399c1d89dec60741b05462e1110945781a845e7db1796c9bca3e909b61205d")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297221),
			GasLimit:    0x8583b00,
			GasUsed:     0xaa7b54,
			Time:        0x64eee0d4,
			Extra:       common.Hex2Bytes("d88301020b846765746888676f312e31392e38856c696e7578000000b19df4a2f8b5830ddfffb860b4de956fe39e1e7e6b535f75054f2efea8ac185ac2b76384c749cc45f7decfea9999b9b3b41e5666e36408069cdd06870e266fe868b9128ce0a0aecf4ecc6b0e42d10c5e646e9571a862dee7112e88f39ddc476b67be9cf32ccb337d3496f901f84c8401dd8ec3a035be0876c21a5a5726193d1adb8ce9ebb6eb95757407cafb9106a57f0d3d12228401dd8ec4a0da01d8fede81f2840ea4bd2d5586d4303f68c73becf0a23d1d6dddd4890bd27480939d982a7958bf5912c1105da4673d035df33b685ad798e6f589ada1b40f63fa27edbbcea731e17928a54c07847a70d7e262a0bc41e68e5fcda124a37f4340b301"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297220: {
			ParentHash:  common.HexToHash("35be0876c21a5a5726193d1adb8ce9ebb6eb95757407cafb9106a57f0d3d1222"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("d93dbfb27e027f5e9e6da52b9e1c413ce35adc11"),
			Root:        common.HexToHash("1283c0ddaf635ba45c04c7dcd3ee83370027d55fee45e9bd2a8f121074808ece"),
			TxHash:      common.HexToHash("62c9236c31a88ecf2d5ecb4666f06b78fad90bc6cde486714c92ac123b26e56b"),
			ReceiptHash: common.HexToHash("babb28a01a3e0fe3657bd296123c54e8e6368588e943ab3b966d97effdcc1ae2"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("dd6dd27f6ff5373d2ef7e4edba7714d6f3daf37bd7d8bfffba5fd73b32befff7ffc75fcbfbb955cfeab33ff77d1fd0b7fbd3d1fc05f5f02f7faf3efff3ed7afcffcffbf65d52ad3c9b75cedcb8d672fafdb1ffbf52f7af38e2fce0bffa6b0e3f5ffd7dfeaf4fcd630acd8ffc917b3dbf0ca36c63784d76b65eead7b9f6ef1fbf5aafff6e2e5fba9ff4ffbd57eeebdcbe76f67db77c36bdc86dabb3d3bccbdead97e726e3ebbefbf6effef7efcffebec3f7afcfef4ebaf9b5f7eaf77e6f5bebebe5e7f59ea35efbf7f6ea5d2fc36ffdeeadffb5eec3ba5e3dbbd774cef627f77ff1bafa3dbdf3d9de2735e7c685a723b6fc3e27fb767d7179bf2ffadfb9b9d8fb")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297220),
			GasLimit:    0x85832a7,
			GasUsed:     0x142a9ef,
			Time:        0x64eee0d1,
			Extra:       common.Hex2Bytes("d88301020a846765746888676f312e32302e34856c696e7578000000b19df4a2f8b5830ddfffb860a4332ba47e6728f4b0e52e490cf4bd3cc4fc742c12bb8a8e5912095c71d9c845f5ddc5fdd4128dbed0e1bcaf4da8236b05ded0aa4a4c81f4a4fb2250f790fe1b21edb1aed0bc7bb7d4a78b26c6dd9af3b65a2bacb5b44e4a7b1dc3284e9d7d6bf84c8401dd8ec2a00f0fc4bae073bbeae37d8ff5395c95480fe5a9149bad4f0cd4f5e8e32d18b32e8401dd8ec3a035be0876c21a5a5726193d1adb8ce9ebb6eb95757407cafb9106a57f0d3d1222805bd7bc1a83768b5c7a505eed50eca8fdddd8a583a46d37f80a2d714be9f8ca2b04b0ba1398956512d18b30e3a56755245df898b98f8d5f6131eb7527f7dc65b101"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297219: {
			ParentHash:  common.HexToHash("0f0fc4bae073bbeae37d8ff5395c95480fe5a9149bad4f0cd4f5e8e32d18b32e"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("d1d6bf74282782b0b3eb1413c901d6ecf02e8e28"),
			Root:        common.HexToHash("a6923c4f9bce5b7cc15a16d85222ce2618aa0f12bca0fa6b703d9aa5aa412648"),
			TxHash:      common.HexToHash("57dc9bdab942a85992bf2bb645594c35d9aa1df083fee8207df409886d313d6d"),
			ReceiptHash: common.HexToHash("3bfa2879642e0273e765670993bff0ec2f22e589317782599cc6b79453441096"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("eea73ff54e3fd7191d7f53f49ca3c29eb5b9688d0fafc2d33e7ee12d3a35cc69ef7ebd97cfc31b177a6b73ccda3622f8f9b6d95b59976d306e3ecc5f7431eed49e5cd9f70ff7cd29df7de1ebfd3bae347bbd98a770559ede746e6f93c9fb4e23eb3db0a53e57bf4eb77d7c36d3f37944cd99ddd55bb79f7e1d57f5b7ee1b6eeaf987c3a4da37134c5fa5be1fd2107bce442ea68f7c34fc9f6beba6df46fb78e63fd7323f57c723718edfddce9bd9fd425a978f65a7bed42fff41f666aefb58f76dd3ff561332ffaeb7fac28d55e99b7cf1f7b1a768b7a8bcc13e7ce715f4affed53eb4e4b1c177bd5f5d2de36bf39bd6b8d6976e5dec7cccb4f4fd932a677ecd")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297219),
			GasLimit:    0x84fe2c6,
			GasUsed:     0xf43a7c,
			Time:        0x64eee0ce,
			Extra:       common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2f8b58305dfffb860b7ab6f15df58e54befa47c75d64d32161823c59edf5bade334b55623eceb68b45752b058206ab7805f277362b7df153601ff6c28a4359f22a74b39304f929af5ca3192f58af57955bd33002402c6824834448ac33fb34d702d2e0d08d2bdcb8cf84c8401dd8ec1a0a493bd047c4c05912239fccde0dbd49730ec916afaabff66abd2e4d8bb351a938401dd8ec2a00f0fc4bae073bbeae37d8ff5395c95480fe5a9149bad4f0cd4f5e8e32d18b32e80515ed762f28c37e88f3d4f21bb33aa5fbd548f3bd168a4b4eeba9ed4c128f79f16886350f86283a571bd58d56cbe2ede5500d4261f8536677c4f24fd5542db0a00"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297218: {
			ParentHash:  common.HexToHash("a493bd047c4c05912239fccde0dbd49730ec916afaabff66abd2e4d8bb351a93"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("cc8e6d00c17eb431350c6c50d8b8f05176b90b11"),
			Root:        common.HexToHash("b111fd4d787edd2a9a54e9e76aabac9bedf7213ddcf44b9960f334c9303b0256"),
			TxHash:      common.HexToHash("c81295cf7fdfa201ef13477519a25fd0749c8085acc05f59386291fb87fac10b"),
			ReceiptHash: common.HexToHash("6024c6edc9ef851420121f565972736a182e40a2c5603c40de6afa8c1848b83b"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("66a0839004a3501048881050821442000900040000814969480300289840012088035a00108000010a8810000142020081618811240001212020801000201a0404140848034000200106100898884434249023000244010020441200808006000a0b20201a0200182402005810002a401800c80c007e36260000401083c8041c0020c39684308058138424001c0010002824a40524a0100810081440184028a0a2d8180110814006022c40083b1001004000a010400200043001192823100044932045028012804600700820a92612402820010000040838011451d20800a4250838080102003100032184089002111440200b0426a050480010300000011001")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297218),
			GasLimit:    0x8583b00,
			GasUsed:     0x35a2f2,
			Time:        0x64eee0cb,
			Extra:       common.Hex2Bytes("d98301020a846765746889676f312e31392e3131856c696e75780000b19df4a2f8b58305dfffb860afc7a23949201ffc50f92c39cb6a3a42f960edd1cafc424b852f511da8dbc052c0bb0222baa6b3a560cfe41a5a7b95c709c11e05b6471366ddb49ba73d0a547b085851d45b51dfbff2cc71f2ca93ff0d6402ecf1c534f240e4b67dfec13efc38f84c8401dd8ec0a03f20b11009e964211f4ef2e316ddf214066d6e0603b6fa2a283b14133c0aeb1b8401dd8ec1a0a493bd047c4c05912239fccde0dbd49730ec916afaabff66abd2e4d8bb351a9380740b7c6a2a1f2f5083f10dc247ed6e6223edd986368179556bd9dd3d78949149357dd8ed2793a0c701d6b7df744261439b71ef4cd2ef9b189afcc2f8105745ac01"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297217: {
			ParentHash:  common.HexToHash("3f20b11009e964211f4ef2e316ddf214066d6e0603b6fa2a283b14133c0aeb1b"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("35ebb5849518aff370ca25e19e1072cc1a9fabca"),
			Root:        common.HexToHash("158af1b358e1ab7c08fd65a742d3e36c372859047aa6d1f8b47f8c6ca1a21d13"),
			TxHash:      common.HexToHash("2e308aa7adf3de39000a56dfd3596435e357304504b9ecb1401492dcc101cf88"),
			ReceiptHash: common.HexToHash("698a86c1d0922df0705634891cbc82347d693ecd9e6ec0201ac9b9ffb4006cca"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("14244b5f06401457d30562ee82512903900dd1291845c7203f3f49517c14a7b08eeb1643764ed0295a328daf585e442a87c0dc3eef1ab1195200248805275b88847ad90a0f564c1a133a621dd3ed006af41494d545467e5bb5bc80f9b4492d36ce2c12a7aa064353340d18a5c11659c04904a952bb45fd06a344ee50cb4db65090eaa19787bd929d10bf6608d508b19b24651ec5f916309b31b68cf88841c8becf098432529e06423a2e83edae3fb40226ea99652ac09398fa626db2f116c4c8a47057b20a81292e13d0ee6674a0d92221d50e6a8126ea1d835fdbdab2b8fc2e9cb0ced009a6e37d29d5a842f5018932486c87294162fae82017883c5e6960c4")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297217),
			GasLimit:    0x8582a4f,
			GasUsed:     0xd5aba3,
			Time:        0x64eee0c8,
			Extra:       common.Hex2Bytes("d883010209846765746888676f312e32302e36856c696e7578000000b19df4a2f8b58305dfffb860887daa937dc51ef1bdc88a3f2e52f51ccc297a3439741cf23e3bc635ecca60413c070e136840b09112b15158b54887801215cd8e3354c552e03587b922c8aa3c444c55336eaefafbf88708863869316136f0a68e45d62c243eb5976dffe41656f84c8401dd8ebfa0d1cbcb756aa78425abb3fd8eb10e814c67b8cf6e7750458b8c597388bd8682928401dd8ec0a03f20b11009e964211f4ef2e316ddf214066d6e0603b6fa2a283b14133c0aeb1b80b2c9b0fdb99e7bf632bdb2ab30550cff337b15c9163fddc47c9cc20b9582c71343857e79a7dae2b38715d8cbecf111a2d1cfcfbad8f4b65abd6b67ad2eb97bb900"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297216: {
			ParentHash:  common.HexToHash("d1cbcb756aa78425abb3fd8eb10e814c67b8cf6e7750458b8c597388bd868292"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("b4dd66d7c2c7e57f628210187192fb89d4b99dd4"),
			Root:        common.HexToHash("57ede97b62b05af1187766de108267b2f27537ed9edd1a9f7889529db4cca965"),
			TxHash:      common.HexToHash("93775f7dad7f0678386501fd9419fbff9e902707687c7baa4c73478a74c4e839"),
			ReceiptHash: common.HexToHash("0e3fb454aa5489134df6457875bfde27c8eb43f092cf84cffb1f007d8acbe1fe"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("882cc625e0381010c60011c584004440a00b1121012c82100038612a16088082af04b07e28049011662125b1020680b08300148849480620d8262c280a640a081d7485d2514625203328e09802204128703622a52054820e002d0000c8148445c20908222a1605012229022681101ae2081010175041848624510818de623640110a2901466d5882229d258e1921042a8924158511203a8800c8c0508980102092050022078d1402270638cd3f0a1640cc058d0108128189308365a020c224d8242c1ce290c361864190899b1a0012542500061006104094031347462a02e424281808521004114021881c484c0041295950e26951e09448a002419000413b06")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297216),
			GasLimit:    0x84fda76,
			GasUsed:     0x847ded,
			Time:        0x64eee0c4,
			Extra:       common.Hex2Bytes("d98301020a846765746889676f312e31392e3131856c696e75780000b19df4a2f8b5830ddfffb860b01dbdb421b9fb5b1f71f196b5e1a8979b353ebfb25bf1a0ef31cc692139df7918b4d2db38211f91ee023bbdf78ac50b00d1497f7d90e42adb7826b21c129d5c706219f37d9216b27be379a4f2854149fa744ef734e8558070c5e8c4fd794c5df84c8401dd8ebea017148d821da270f09540075131924b934bc0352dd71d8b8f60274f39d73231dd8401dd8ebfa0d1cbcb756aa78425abb3fd8eb10e814c67b8cf6e7750458b8c597388bd868292808cc4af67170aaa12f7d2c44f6bec754eadaad71cf12df86ad87667fb2a9579940c3a579e4b5bb1c63c02b488429d683c9e3599b0583f0e4015e98d674f58281f00"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297215: {
			ParentHash:  common.HexToHash("17148d821da270f09540075131924b934bc0352dd71d8b8f60274f39d73231dd"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("b218c5d6af1f979ac42bc68d98a5a0d796c6ab01"),
			Root:        common.HexToHash("a6e3d060b775a05eaa251583fb248df682fa839bd575d254eb4330d6c92f0356"),
			TxHash:      common.HexToHash("4a81457bed752a8611a3fcf1409e6bfc34c8006651f64e6fdb5530a113d8972c"),
			ReceiptHash: common.HexToHash("a8539e4e87b5ce8ee15a3f9c52113062de8883c1ea062255959e3f3f5350ec00"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("1177e6c48819553fcb1800dcc612126e24206a13cd4d054bea0a07b81f00a03e8e43f8c2c280905102227a9355da28c18e1c000a071c61a140352328113d228007d640114bd00c69db01a449c835706ca8dcab8040744802234e841088042eb42a9d23365e262a370c0d001693eafc4348a6a440294fdf23010ac910944b6df8888ec3142b5440dc30aa5500a894fc9bf0b646857e6838de318cd1c300929c20828da412533ae674b610cb1e6f85e54008a4d7d0878286a57469a3b57752a2e7b38f14828276a80e30ab1c48f23650e0b41447e5c14f0417819e7d87833cefab519990032ef8e7274125026a8163152c278085a9d078c67c23b87d00da15f00b")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297215),
			GasLimit:    0x84792e5,
			GasUsed:     0xaac551,
			Time:        0x64eee0c1,
			Extra:       common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2f8b5830ddfffb860b9b1a84b0ce0243a5b8592ffc6562f1f9f6e95197a1ca1ea17eb300b4f101382aff91466a179e5b9ad4fa3423f6ed40f14edef5a8addd5590754049e9828001b3aa9e355abd098b01b059781512654f9dce636865e6dd98c9c41066e09f64af2f84c8401dd8ebda0e5ec815463a14fc0a652302d302e93c2a06280d115782a183f5378692e8638288401dd8ebea017148d821da270f09540075131924b934bc0352dd71d8b8f60274f39d73231dd80615eb294cc6f3c9f5a307dca23199a3844bf63b9078c41d457d7732a7956189a4b68c910d5ed0eb7fe4e7432061ab9ca6bccc150099a5ad554112aa34494cc6801"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297200: {
			ParentHash:  common.HexToHash("10c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc3"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("e9ae3261a475a27bb1028f140bc2a7c843318afd"),
			Root:        common.HexToHash("111504390f79560bc7fa5d1d5b8cdb109d0ee1aa76816dba43b88f1e5e152b6f"),
			TxHash:      common.HexToHash("2a2eb4b95af5c879078702bfccd6147bd47162385898d8c0b81fb94ac5456a60"),
			ReceiptHash: common.HexToHash("57ed46152fe61b432e682da7f0b82273e1fee09674b2b62c64952e214e84e7e1"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("8035164600431418e05400edd0a9650e8238c2833980eadd35538b33d3641594dc00b06042041004a398b01d821aa232cdb11211a044e339d0440056482c6ec80e61d00aa544881a8b61c2988220042db254432900c6c8410045114b80ba0021cd6c20a23a66042b04e7971817460d6c0e52e95a00cbd4809580a4d9804e35c6998a0246256d4ad18588363010a0d89030240cc77a03158a0f4d83c0ac841cb0871050c8f73f6650e6153206369a061a0c2d90835480d12374a021a1605a2278e3209746a6555c4a32a1b8b87a629968238cf242a088ac190b51fa72f200e56c579880051b2042824b1db2e7058003c1619cfc307864e1c8fc2a690299a1864a")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297200),
			GasLimit:    0x8583b00,
			GasUsed:     0xa136cd,
			Time:        0x64eee094,
			Extra:       common.Hex2Bytes("d88301020a846765746888676f312e32302e35856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e2465176c461afb316ebc773c61faee85a6515daa8a923564c6ffd37fb2fe9f118ef88092e8762c7addb526ab7eb1e772baef85181f892c731be0c1891a50e6b06262c816295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e9135ebb5849518aff370ca25e19e1072cc1a9fabcaa7f3e2c0b4b16ad183c473bafe30a36e39fa4a143657e229cd23c77f8fbc8e4e4e241695dd3d248d1e51521eee6619143f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2569c77a677c40c7fbea129d4b171a39b7a8ddabfab2317f59d86abfaf690850223d90e9e7593d91a29331dfc2f84d5adecc75fc39ecab4632c1b4400a3dd1e1298835bcca70f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d78869485a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be807dddb074639cd9fa61b47676c064fc50d62cb1f2c71577def3144fabeb75a8a1c8cb5b51d1d1b4a05eec67988b8685008baa17459ec425dbaebc852f496dc92196cdcc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419d1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ee226379db83cffc681495730c11fdde79ba4c0cae7bc6faa3f0cc3e6093b633fd7ee4f86970926958d0b7ec80437f936acf212b78f0cd095f4565fff144fd458d233a5bef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b5830aefffb86097bc63a64e8d730014c39dcaac8f3309e37a11c06f0f5c233b55ba19c1f6c34d2d08de4b030ce825bb21fd884bc0fcb811336857419f5ca42a92ac149a4661a248de10f4ca6496069fdfd10d43bc74ccb81806b6ecd384617d1006b16dead7e4f84c8401dd8eaea0e61c6075d2ab24fcdc423764c21771cac6b241cbff89718f9cc8fc6459b4e7578401dd8eafa010c8358490a494a40c5c92aff8628fa770860a9d34e7fb7df38dfb208b0ddfc380ff15abfc44495e4d4605458bb485f0cac5a152b380a8d0208b3f9ff6216230ec4dd67a73b72b1d17a888c68e111f806ef0b255d012b5185b7420b5fb529c9b9300"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
		31297000: {
			ParentHash:  common.HexToHash("a0dcc3ef1e710117565d6cef7043f133595a6f1f57d9b49ded7e1a6bfb72659f"),
			UncleHash:   common.HexToHash("1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    common.HexToAddress("7ae2f5b9e386cd1b50a4550696d957cb4900f03a"),
			Root:        common.HexToHash("e1772b39389c0b57e2a0913139235645fdcc84efd902ca5f04bf0fbd6743359b"),
			TxHash:      common.HexToHash("ea03b553afbb8246c565183867eda57cc6a61cbd92064943462a04ba48122f1f"),
			ReceiptHash: common.HexToHash("bd65f987e07936f6c672a1a4696f95777ed53e24cb0998e9e7f0753fb23adcc9"),
			Bloom:       types2.BytesToBloom(common.Hex2Bytes("00684a0053503034e8cc814484010402e200d85414009228001021810328923097109fe8460030048066d40814560200c930ca1082b86800655a8450002812210446842123e82900033012a842c08c2930d0092109604f0c060c1241802504c4042c0b607b2e08601204120013040a28384200d2006194a32c21381247430486102a210549594a842689948804b93588112e04852910908805d400c10081742023030c08d2a2021800080e17078e404848000cba05c2940208402479491360aa69e5d6020187480f90408900510098222201a20000586835a7d8d3323308a1c91c10a0020046a3b20b112445050057110b06840001a086ca2821b20098010018")),
			Difficulty:  big.NewInt(2),
			Number:      big.NewInt(31297000),
			GasLimit:    0x85832a7,
			GasUsed:     0x6c005b,
			Time:        0x64eede3a,
			Extra:       common.Hex2Bytes("d88301020b846765746888676f312e31392e38856c696e7578000000b19df4a2150bac492386862ad3df4b666bc096b0505bb694dab0bec348681af766751cb839576e9c515a09c8bffa30a46296ccc56612490eb480d03bf948e10005bbcc0421f90b3d4e295e26495cef6f69dfa69911d9d8e4f3bbadb89b977cf58294f7239d515e15b24cfeb82494056cf691eaf729b165f32c9757c429dba5051155903067e56ebe3698678e912d4c407bbe49438ed859fe965b140dcf1aab71a993c1f7f6929d1fe2a17b4e14614ef9fc5bdc713d6631d675403fbeefac55611bf612700b1b65f4744861b80b0f7d6ab03f349bbafec1551819b8be1efea2fc46ca749aa184248a459464eec1a21e7fc7b71a053d9644e9bb8da4853b8f872cd7c1d6b324bf1922829830646ceadfb658d3de009a61dd481a114a2e761c554b641742c973867899d38a80967d39e406a0a9642d41e9007a27fc1150a267d143a9f786cd2b5eecbdcc4036273705225b956d5e2f8f5eb95d2570f657164e5b75689b64b7fd1fa275f334f28e1896a26afa1295da81418593bd12814463d9f6e45c36a0e47eb4cd3e5b6af29c41e2a3a5636430155a466e216585af3ba772b61c6014342d914470ec7ac2975be345796c2b81db0422a5fd08e40db1fc2368d2245e4b18b1d0b85c921aaaafd2e341760e29fc613edd39f71254614e2055c3287a517ae2f5b9e386cd1b50a4550696d957cb4900f03ab84f83ff2df44193496793b847f64e9d6db1b3953682bb95edd096eb1e69bbd357c200992ca78050d0cbe180cfaa018e8b6c8fd93d6f4cea42bbb345dbc6f0dfdb5bec73a8a257074e82b881cfa06ef3eb4efeca060c2531359abd0eab8af1e3edfa2025fca464ac9c3fd123f6c24a0d788694859f8ccdafcc39f3c7d6ebf637c9151673cbc36b888819ec5ec3e97e1f03bbb4bb6055c7a5feac8f4f259df58349a32bb5cb377e2cb1f362b77f1dd398cfd3e9dba46138c3a6f79b60359f141df90a0c745125b131caaffd12b772e180fbf38a051c97dabc8aaa0126a233a9e828cdafcc7422c4bb1f4030a56ba364c54103f26bad91508b5220b741b218c5d6af1f979ac42bc68d98a5a0d796c6ab01b659ad0fbd9f515893fdd740b29ba0772dbde9b4635921dd91bd2963a0fc855e31f6338f45b211c4e9dedb7f2eb09de7b4dd66d7c2c7e57f628210187192fb89d4b99dd4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000cc8e6d00c17eb431350c6c50d8b8f05176b90b11b3a3d4feb825ae9702711566df5dbf38e82add4dd1b573b95d2466fa6501ccb81e9d26a352b96150ccbf7b697fd0a419ce2fd7544e0b2cc94692d4a704debef7bcb61328b64abe25614c9cfd32e456b4d521f29c8357f4af4606978296c9be93494072ac05fa86e3d27cc8d66e65000f8ba33fbbd1d6bf74282782b0b3eb1413c901d6ecf02e8e28939e8fb41b682372335be8070199ad3e8621d1743bcac4cc9d8f0f6e10f41e56461385c8eb5daac804fe3f2bca6ce739d93dbfb27e027f5e9e6da52b9e1c413ce35adc11000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e2d3a739effcd3a99387d015e260eefac72ebea1956c470ddff48cb49300200b5f83497f3a3ccb3aeb83c5edd9818569038e61d197184f4aa6939ea5e9911e3e98ac6d21e9ae3261a475a27bb1028f140bc2a7c843318afd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ea0a6e3c511bbd10f4519ece37dc24887e11b55db2d4c6283c44a1c7bd503aaba7666e9f0c830e0ff016c1c750a5e48757a713d0836b1cabfd5c281b1de3b77d1c192183ef0274e31810c9df02f98fafde0f841f4e66a1cd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f8b58305efffb8608605e361feaf30688803514a7f9071e61d3c992661f0f80671c72c571dccb7a3246759b622c815cd0fb166418e4e3ffe0a580db63d13c177927edebe460a7fd0ee1df6fc5e44f85c23f8fb01d413da8ced2c01e12e5c750dede91e845889837af84c8401dd8de6a01152160e4ada894f35233d509c56afb367fc0e43b7f5b4c4548d74bd48a519648401dd8de7a0a0dcc3ef1e710117565d6cef7043f133595a6f1f57d9b49ded7e1a6bfb72659f80ec5abb982536097426af57180bec01ec157d60c0e1c227d7b5f96c5e218c055834f7c5464206a5e49c638a1490ca924c38a6923909dccc9baae17f7a7d8011ee00"),
			MixDigest:   common.HexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Nonce:       [8]byte{},
			BaseFee:     big.NewInt(0),
		},
	}
	if v, ok := headerMap[height]; ok {
		return &v, nil
	}

	return &types2.Header{
		Root:   common.HexToHash("c84307dfe4ccfec4a851a77755d63228d8e0b9ba3345d1eee37ed729ee16eaa1"),
		Number: big.NewInt(int64(height)),
	}, nil
}

func (c *mockChain) GetProof(_ common.Address, _ [][]byte, _ *big.Int) (*client.StateProof, error) {
	// eth.getProof("0xaa43d337145E8930d01cb4E60Abf6595C692921E",["0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae"], 21400)
	// storageHash: "0xc3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0",
	accountProof := []string{
		"0xf901f1a080679a623dfdd0dfa34cb5c1db80292abdc2a9e75f5026e3d24cd10ea58f8e0da04f4d7ef0e932874f07aec064ee1281cd6a3245fceab78bdd6a8d2d7a86d27451a0a715335e2de6e91c28910eff04e8709ff6ca93862121a0b52560071867a9f14080a0842db9556e659d64ca9d2d33229ebac6e7e2185f42bd07965464de8064d94ac8a0d94bd2db341ab9d3d7f4fe0aa569bb21dfac0d5eb0ec008c7af23d7f2ed98ec1a0cee66e2515872d5f4b42ada7cc733288809c11ab99aa0d25eb941236074f9904a0f3aa8d1774f013d8af0fdd8364c7833f16b42ad377e485b754f5cdae6fedaa2fa0bffc6b17aaf862725aaf4af4ecda3ed70d4102b875451eb965259ead260b06c7a026a29f57f5efaf83a8f098ed0ba0f53aac353364ce498a82d589e7bcf1f84e76a01a25f2cac2c6a021225ea182c3c391c0fafac96cb38896eb45648a5c33f31b6ca0d4d6b410f89044b335cc7b14221050035d87d390043bf6d84bc0f8005391f213a092dfa1004df4e71ccfaf3a6d682718f1fbb2d1e6411566e139f1efa74874c303a078455f6ef72aa4dc670e9b467fdbe29d37b5c4eb526ee07b372d2bcea57871eaa05911bcb62e4ba3117ca428f93305ebf06247d573f25bb0fff22681716c21744da0f47e1a054e1ee9ac18fd711b2571c2cab26e88d1a5be46d7078723076866265880",
		"0xf851808080808080808080a08ffa88d75a03fd29af8cb1a4ac016e32ef8e39631a6bf45d79a34adfc4ecb1448080a0a1161a49c0c7e7a92a2efe173abffdbb1ed91e5235688e2edbc4e38078dc5c5580808080",
		"0xf869a02012683435c076b898a6cac1c03e41900e379104fefd4219d99f7908cb59cfb3b846f8440180a0c3608871098f21b59607ef3fb9412a091de9246ad1281a92f5b07dc2f465b7a0a07498e14000b8457a51de3cd583e9337cfa52aee2c2e9f945fac35a820e685904",
	}
	accountProofRLP, err := encodeRLP(accountProof)
	if err != nil {
		return nil, err
	}
	storageProof := []string{"0xf8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080"}
	storageProofRLP, err := encodeRLP(storageProof)
	if err != nil {
		return nil, err
	}
	return &client.StateProof{
		AccountProofRLP: accountProofRLP,
		StorageProofRLP: [][]byte{storageProofRLP},
	}, nil
}

func (c *mockChain) LatestHeight() (exported.Height, error) {
	return clienttypes.NewHeight(0, c.latestHeight), nil
}

type ProverTestSuite struct {
	suite.Suite
	prover *Prover
	chain  *mockChain
}

func TestProverTestSuite(t *testing.T) {
	suite.Run(t, new(ProverTestSuite))
}

func (ts *ProverTestSuite) SetupTest() {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  9999,
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  ibcHandlerAddress,
	})
	ts.Require().NoError(err)
	codec := core.MakeCodec()

	err = chain.Init("", 0, codec, false)
	ts.Require().NoError(err)
	// call SetRelayInfo
	err = chain.SetRelayInfo(&core.PathEnd{
		ClientID:     "mock-client-0",
		ConnectionID: "connection-0",
		ChannelID:    "channel-0",
		PortID:       "transfer",
		Order:        "UNORDERED",
	}, nil, nil)
	ts.Require().NoError(err)

	config := ProverConfig{
		TrustingPeriod: 100 * time.Second,
		MaxClockDrift:  1 * time.Millisecond,
		Debug:          true,
	}
	ts.chain = &mockChain{
		Chain:        NewChain(chain),
		latestHeight: 31297221,
		chainID:      9999,
	}
	ts.prover = NewProver(ts.chain, &config).(*Prover)
}

func (ts *ProverTestSuite) TestQueryLatestFinalizedHeader() {
	header, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().NoError(err)
	ts.Require().NoError(header.ValidateBasic())
	ts.Require().Len(header.(*Header).Headers, 3)
	h := header.(*Header)
	target, err := h.Target()
	ts.Require().NoError(err)
	ts.Require().Equal(target.Number.Int64(), int64(31297219))
	ts.Require().Len(h.PreviousValidators, 21)
	ts.Require().Len(h.CurrentValidators, 21)
	ts.Require().NotEqual(common.BytesToHash(crypto.Keccak256(h.CurrentValidators...)), common.BytesToHash(crypto.Keccak256(h.PreviousValidators...)))
}

func (ts *ProverTestSuite) TestSetupHeadersForUpdate() {
	type dstMock struct {
		Chain
		core.Prover
	}
	dst := dstMock{
		Chain:  ts.prover.chain,
		Prover: ts.prover,
	}
	latest := ts.chain.latestHeight
	defer func() {
		ts.chain.latestHeight = latest
	}()
	ts.chain.latestHeight = 31297000

	header, err := ts.prover.GetLatestFinalizedHeaderByLatestHeight(latest)
	ts.Require().NoError(err)
	setupDone, err := ts.prover.SetupHeadersForUpdate(&dst, header)
	ts.Require().NoError(err)
	ts.Require().Len(setupDone, 2)
	first := setupDone[0].(*Header)
	ts.Require().Len(first.Headers, 17) // Vote is not present in 14blocks(31297201~31297214)
	ts.Require().Equal(int(first.GetHeight().GetRevisionHeight()), 31297200)
	second := setupDone[1].(*Header)
	ts.Require().Len(second.Headers, 3)
	ts.Require().Equal(int(second.GetHeight().GetRevisionHeight()), 31297219)
}

func (ts *ProverTestSuite) TestCreateMsgCreateClient() {

	finalizedHeader, err := ts.prover.GetLatestFinalizedHeader()
	ts.Require().NoError(err)
	target, err := finalizedHeader.(*Header).Target()
	ts.Require().NoError(err)
	stateRoot, err := ts.prover.GetStorageRoot(target)
	ts.Require().NoError(err)
	previousEpoch := GetPreviousEpoch(finalizedHeader.GetHeight().GetRevisionHeight())
	previousValidatorSet, err := ts.prover.QueryValidatorSet(previousEpoch)
	ts.Require().NoError(err)
	currentEpoch := GetCurrentEpoch(finalizedHeader.GetHeight().GetRevisionHeight())
	currentValidatorSet, err := ts.prover.QueryValidatorSet(currentEpoch)
	ts.Require().NoError(err)
	msg, err := ts.prover.CreateMsgCreateClient("", finalizedHeader, types.AccAddress{})
	ts.Require().NoError(err)
	ts.Require().Equal(msg.ClientState.TypeUrl, "/ibc.lightclients.parlia.v1.ClientState")
	var cs ClientState
	ts.Require().NoError(proto.Unmarshal(msg.ClientState.Value, &cs))
	ts.Require().Equal(cs.ChainId, uint64(9999))
	ts.Require().Equal(cs.TrustingPeriod, 100*time.Second)
	ts.Require().Equal(cs.MaxClockDrift, 1*time.Millisecond)
	ts.Require().False(cs.Frozen)
	ts.Require().Equal(common.Bytes2Hex(cs.IbcStoreAddress), ibcHandlerAddress)
	var commitment [32]byte
	ts.Require().Equal(common.Bytes2Hex(cs.IbcCommitmentsSlot), common.Bytes2Hex(commitment[:]))
	ts.Require().Equal(cs.GetLatestHeight(), finalizedHeader.GetHeight())

	var consState ConsensusState
	ts.Require().NoError(proto.Unmarshal(msg.ConsensusState.Value, &consState))
	ts.Require().Equal(consState.CurrentValidatorsHash, crypto.Keccak256(currentValidatorSet...))
	ts.Require().Equal(consState.PreviousValidatorsHash, crypto.Keccak256(previousValidatorSet...))
	ts.Require().Equal(consState.Timestamp, target.Time)
	ts.Require().Equal(common.BytesToHash(consState.StateRoot), stateRoot)
}

func (ts *ProverTestSuite) TestQueryClientStateWithProof() {
	ctx := core.NewQueryContext(context.TODO(), clienttypes.NewHeight(0, 21400))
	cs, err := ts.prover.chain.QueryClientState(ctx)
	ts.Require().NoError(err)

	bzCs, err := ts.prover.chain.Codec().Marshal(cs)
	ts.Require().NoError(err)

	proof, proofHeight, err := ts.prover.ProveState(ctx, host.FullClientStatePath(ts.prover.chain.Path().ClientID), bzCs)
	ts.Require().NoError(err)

	ts.Require().Equal(proofHeight.GetRevisionNumber(), uint64(0))
	ts.Require().Equal(proofHeight.GetRevisionHeight(), uint64(21400))

	// storage_key is 0x0c0dd47e5867d48cad725de0d09f9549bd564c1d143f6c1f451b26ccd981eeae
	ts.Require().Equal(common.Bytes2Hex(proof), "f853f8518080a0143145e818eeff83817419a6632ea193fd1acaa4f791eb17282f623f38117f568080808080808080a016cbf6e0ba10512eb618d99a1e34025adb7e6f31d335bda7fb20c8bb95fb5b978080808080")
}

func (ts *ProverTestSuite) TestConnection() {
	res := &conntypes.QueryConnectionResponse{
		Connection: &conntypes.ConnectionEnd{
			ClientId: "99-parlia-0",
			Versions: []*conntypes.Version{
				{
					Identifier: "1",
					Features:   []string{"ORDER_ORDERED", "ORDER_UNORDERED"},
				},
			},
			State: conntypes.OPEN,
			Counterparty: conntypes.Counterparty{
				ClientId:     "99-parlia-0",
				ConnectionId: "connection-0",
				Prefix: types3.MerklePrefix{
					KeyPrefix: []byte("ibc"),
				},
			},
			DelayPeriod: 0,
		},
		Proof: []byte{249, 2, 108, 249, 1, 177, 160, 243, 2, 132, 113, 118, 63, 160, 241, 161, 149, 174, 195, 18, 210, 53, 140, 244, 55, 106, 61, 135, 92, 126, 3, 174, 227, 145, 76, 246, 158, 163, 237, 128, 160, 127, 209, 245, 74, 140, 45, 22, 54, 65, 152, 69, 181, 239, 59, 177, 124, 160, 102, 90, 184, 251, 217, 5, 60, 213, 213, 82, 239, 90, 170, 6, 2, 160, 41, 212, 235, 101, 41, 88, 83, 242, 202, 249, 194, 236, 70, 87, 205, 86, 210, 185, 20, 24, 165, 108, 78, 217, 227, 185, 171, 69, 147, 24, 214, 229, 160, 145, 96, 113, 245, 236, 179, 190, 225, 105, 241, 251, 65, 3, 235, 190, 98, 50, 95, 13, 58, 158, 126, 255, 126, 200, 182, 162, 184, 82, 48, 67, 136, 128, 160, 175, 124, 86, 245, 185, 249, 125, 146, 23, 9, 218, 185, 15, 109, 124, 33, 250, 59, 89, 96, 116, 82, 243, 65, 10, 193, 8, 40, 144, 139, 38, 64, 160, 224, 191, 86, 228, 105, 21, 42, 129, 130, 172, 228, 96, 248, 83, 25, 223, 99, 214, 201, 190, 202, 139, 42, 196, 142, 81, 92, 44, 50, 172, 251, 42, 160, 67, 76, 154, 154, 112, 58, 176, 167, 174, 126, 79, 134, 194, 208, 154, 245, 161, 106, 236, 125, 64, 136, 202, 72, 61, 70, 170, 12, 109, 132, 68, 213, 160, 170, 218, 158, 181, 234, 137, 42, 205, 212, 206, 113, 31, 185, 40, 158, 248, 185, 203, 175, 103, 31, 6, 150, 105, 26, 169, 115, 42, 94, 238, 154, 22, 160, 209, 8, 0, 140, 126, 171, 172, 12, 93, 82, 67, 64, 234, 3, 152, 165, 245, 137, 166, 131, 218, 2, 177, 29, 84, 166, 186, 8, 42, 245, 54, 145, 160, 183, 120, 101, 29, 90, 126, 76, 66, 215, 15, 21, 193, 218, 17, 65, 15, 9, 145, 242, 3, 203, 163, 150, 91, 77, 134, 86, 62, 207, 117, 71, 143, 160, 90, 98, 164, 183, 88, 177, 161, 231, 114, 25, 237, 70, 112, 69, 253, 90, 125, 202, 100, 255, 155, 200, 174, 225, 111, 199, 221, 194, 180, 124, 109, 50, 160, 187, 51, 102, 98, 64, 251, 30, 166, 130, 29, 10, 59, 50, 19, 246, 48, 184, 197, 144, 98, 5, 83, 71, 101, 160, 145, 11, 13, 122, 129, 16, 210, 128, 160, 67, 199, 95, 200, 128, 34, 48, 39, 12, 122, 115, 104, 117, 172, 182, 198, 69, 116, 151, 124, 143, 65, 129, 117, 79, 249, 190, 133, 168, 70, 52, 10, 128, 248, 113, 128, 160, 14, 147, 143, 255, 173, 177, 239, 236, 164, 203, 229, 21, 75, 174, 164, 236, 137, 188, 190, 203, 85, 8, 192, 11, 104, 183, 162, 207, 238, 101, 38, 67, 128, 128, 128, 160, 123, 103, 52, 98, 145, 109, 110, 134, 48, 20, 137, 241, 181, 253, 251, 6, 99, 206, 99, 49, 92, 213, 63, 76, 18, 22, 72, 175, 130, 0, 232, 129, 128, 128, 128, 128, 128, 128, 128, 160, 86, 19, 50, 171, 19, 198, 195, 87, 230, 246, 175, 58, 22, 123, 6, 172, 13, 14, 227, 136, 240, 15, 9, 31, 226, 129, 35, 102, 111, 56, 184, 89, 128, 128, 128, 248, 67, 160, 32, 63, 196, 45, 223, 108, 27, 91, 178, 24, 206, 36, 225, 76, 64, 175, 158, 14, 177, 39, 165, 215, 96, 80, 211, 125, 115, 105, 226, 252, 74, 71, 161, 160, 204, 217, 3, 153, 193, 37, 239, 8, 122, 231, 131, 230, 12, 221, 239, 188, 60, 214, 40, 96, 20, 190, 116, 18, 211, 1, 38, 31, 98, 207, 103, 255},
	}
	path := host.ConnectionPath("connection-0")
	commitment, err := messageToCommitment(res.Connection)
	ts.Require().NoError(err)
	storageRoot := common.BytesToHash([]byte{82, 151, 170, 160, 133, 205, 75, 144, 49, 43, 13, 172, 81, 2, 52, 123, 17, 51, 253, 55, 100, 124, 234, 205, 131, 149, 248, 211, 22, 210, 2, 68})
	ts.Require().NoError(verifyMembership(storageRoot, res.Proof, path, commitment))
}

func (ts *ProverTestSuite) TestConnectionStateProofAsLCPCommitment() {
	proofInit := common.Hex2Bytes("f8ccb872f8708369626398636f6e6e656374696f6e732f636f6e6e656374696f6e2d30a022ab576a7df38bb4860ffbc65f30d5a66536fb2d8ec3d5d7d4ab9a3ead0e4312900000000000000000000000000000013da0ee0b5f32ae2bff0d82149ea22b02e350fbbe467a514ba80bbadd89007df1d167949c406cd64ce7fce93eb9d7baf1288c41c921521db84153509dc20ccebff5b9b436cf108737c6bdc24782569dff36a376353407cbe19b2a8fbaf045755a52c674612a1274da7363cd84ff9f5272abcd0ae4e7043f3a3b00")
	var rawValueProof [][]byte
	if err := rlp.DecodeBytes(proofInit, &rawValueProof); err != nil {
		panic(err)
	}
	// same validation as LCPCommitment.sol#parseStateCommitmentAndProof
	ts.Require().Len(rawValueProof, 3)
	commitmentBytes := rawValueProof[0]
	signer := common.BytesToAddress(rawValueProof[1])
	signature := rawValueProof[2]
	ts.Require().Len(signature, 65)
	ts.Require().Equal(signer.String(), "0x9C406cD64Ce7fce93Eb9d7bAf1288C41C921521D")

	var rawCommitmentProof [][]byte
	if err := rlp.DecodeBytes(commitmentBytes, &rawCommitmentProof); err != nil {
		panic(err)
	}
	ts.Require().Len(rawCommitmentProof, 5)

	// assert commitment
	commitmentPrefix := rawCommitmentProof[0]
	commitmentPath := rawCommitmentProof[1]
	commitmentValue := common.BytesToHash(rawCommitmentProof[2])
	commitmentHeight := new(big.Int).SetBytes(rawCommitmentProof[3]).Uint64()
	commitmentStateId := common.BytesToHash(rawCommitmentProof[4])
	ts.Require().Equal(string(commitmentPrefix), "ibc")
	ts.Require().Equal(string(commitmentPath), "connections/connection-0")
	ts.Require().Equal(commitmentValue.String(), "0x22ab576a7df38bb4860ffbc65f30d5a66536fb2d8ec3d5d7d4ab9a3ead0e4312")
	ts.Require().Equal(commitmentHeight, uint64(317))
	ts.Require().Equal(commitmentStateId.String(), "0xee0b5f32ae2bff0d82149ea22b02e350fbbe467a514ba80bbadd89007df1d167")
}
