package module

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

func fromRlp(hex string) *types.Header {
	var h types.Header
	err := rlp.DecodeBytes(common.Hex2Bytes(hex), &h)
	if err != nil {
		log.Fatal(err)
	}
	return &h
}

func previousEpochHeader() *types.Header {
	return fromRlp("f90391a0844dee9abff97d261ae0049fe38246ac10aba49f2b8618f28f7c2d19e62eccf9a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a0ecf1aa30fa754576ac4abc3cf2a61d1babd41c7e5515855efd857b2d3f37866ba00f0ea7d212c4aaca329b03f5e9ed9c69d3641eb5e03a4edb69b61e6f9d8d51efa0c3372a1f332fc4245e1a9fdcb62580fc6dae741087a8029560f19216dd3d58b9b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000281c88402625a00826c7484669a6a91b90173d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6d9a13701eafb76870cb220843b8c6476824bfa158c66a3f3d2fba1d440da8edc79b59ed9a3a43db62bd7659f7d4e25073f9241dba560600b23e26c30d48ea0395eeeb4ede04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a567015cbc63c3d778cef5e8dbfdaf1fd8f758a764f2667aad2d3775954e4ac23e726226b66f0a94631bd0b6d937b22955d73eed65a31a6f535662f51cc7547143f6f201a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
}

func epochHeader() *types.Header {
	return fromRlp("f90442a04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5da01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948fdaaa7e6631e438625ca25c857a3727ea28e565a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa0015ebe4a5d6cd56f0bf97db1d21746f59ab5cbecf216e34753920d815403ada2a03cd1ebc99cd975182c58de47be968c97658cff4c465e20654185f408a851403cb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201908402625a008229a884669a6ce9b90223d98301040b846765746889676f312e32312e3132856c696e7578000020155b72048fdaaa7e6631e438625ca25c857a3727ea28e565b94a73be71b4a5703b4d0b36e4f65d52615b668b385efc047f7f385ace378981fa3750a0bc16ca6f8217be599bcfa274a7876ea32e7a748c697d01345145485561305b248cd0ede772633b8baea9958f9b602db36d78934d948244a13c2d66e998f987783276e9aee6facbff50b0d63574406b51b2e42bc54d19116d2348ac83461e2e0915d508ad921ebe99c27c8fdbd30aecdbe86f95aee2e06995f83ebeb327924669629f193ffd3257315c79ed5a4867ec53b502b5e6e04db2de85453e0936b441c339a26d10cfa71b50a50dd5edefbafd33740101d074b6d58b56a787a7644ddfff77d0c00f9e62cc58c931e671afc564f3f6e255cc6fc8a56701f8ae0fb8608b6dc552b410a6fa44fa31643850bcb314f1d4edb32c0c79ee3efef5397691f3685d80057d77510a00e77a39e8b2497419053c3b81a8901e85590a20a0a2dad529c82f6c175ec3ebca8a9112415aa94718af673c16c0e90e327e27709666e499f84882018ea0709f88597f05218c198818991cf5598c9280db30d5bfe899da9b7a8c963bff6c82018fa04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d8012518315e9c22a4a648f4d26efcf57f877a26498de6d53fe7a267e8d5ef01482009817fc9de90ca8008ef1f420aa606ddc0c56a975bace3906601fd5cde657d600a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
}

func epochHeaderPlus1() *types.Header {
	return fromRlp("f9032ea0e256fac4dd62cc71eaefd8d6c24ae5209c0e48f5c0b62bcced06dfa838c2ad31a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794b2e42bc54d19116d2348ac83461e2e0915d508ada08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201918402625a008084669a6cecb90111d98301040b846765746889676f312e32312e3132856c696e7578000020155b72f8ae0fb8609338bf42b6ef715e9c887e1b285e706355c2a993cd227497b447f8aad4b7fa44d18cd895862e1a2b961b78656d620f9c015e777cf9bcb6c50e1db2783818bd91f647f6879f8bd199f266f1166f9241f00f955fb5210e7e89e7678680900d1cc1f84882018fa04ec3c90370deeeab62de72108470bccac75d1abe118a778f01afa7a99c976a5d820190a0e256fac4dd62cc71eaefd8d6c24ae5209c0e48f5c0b62bcced06dfa838c2ad3180d6c559fceda3795918a2fdd34cb4f77608c8095b67091aadcef5d5c4efe09fc0477a1977c0037f19244240648637c3c5ec9fe22f1db4899a022575bec96cfccf01a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
}

func epochHeaderPlus2() *types.Header {
	return fromRlp("f9032ea0669cf9fc3b3d480d7a8de1cc051e1d95fa24be8870aa93bb5bfc03f1d592db46a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794d9a13701eafb76870cb220843b8c6476824bfa15a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201928402625a008084669a6cefb90111d98301040b846765746889676f312e32312e3132856c696e7578000020155b72f8ae0fb8608d465cf004274c6caaddd041116c93f12cba458536d23efc6148134af357a685c7b84b1d3de59eb61b2756e7b99a99de00a6158245e16d23dd10602ef719836ce41862c567745727a145df4bde737415e8ffbec5056d4653b02fe1ddc8f21069f848820190a0e256fac4dd62cc71eaefd8d6c24ae5209c0e48f5c0b62bcced06dfa838c2ad31820191a0669cf9fc3b3d480d7a8de1cc051e1d95fa24be8870aa93bb5bfc03f1d592db468083bec325b33996ad2fc35e146d7e8fd215177b1423378d3b07cb73293d5384733488a981b64ae03cbcf44af281cfbd57431addbd8432042c49fb4c78d5573f0d01a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
}

func epochHeaderPlus3() *types.Header {
	return fromRlp("f9032ea0705444384ac2acf605c4ca8353e5070f37cc0976597dfdfc74413e278725eee7a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794e04db2de85453e0936b441c339a26d10cfa71b50a08d30abd786d85a8a10ba441afafdf853b7fd2769351f6600402b88a1ac2d4d7aa056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028201938402625a008084669a6cf2b90111d98301040b846765746889676f312e32312e3132856c696e7578000020155b72f8ae0fb860a6e88879bc7f8ff4ee90ba0776b414b97ffcb0a2bc4f9484075ea121e00b7ce27e21c0fbf76f34746b232e76a37cfc220011efbd7863237e75eb2830748736e9092847a283401bf6793e94d00619764884d012609cf5bace403e93b66f60c3bcf848820191a0669cf9fc3b3d480d7a8de1cc051e1d95fa24be8870aa93bb5bfc03f1d592db46820192a0705444384ac2acf605c4ca8353e5070f37cc0976597dfdfc74413e278725eee780f5171f7fc98ba3118543b55677128c397b07085a29dcd71c989c7f0eb602308342cf64470f85ebe31ebdda4244e7992b82ff2026ac461d0f2997332ccb99255201a0000000000000000000000000000000000000000000000000000000000000000088000000000000000080a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080")
}

func headerByHeight(height int64) *types.Header {
	switch height {
	case 200:
		return previousEpochHeader()
	case 400:
		return epochHeader()
	case 401:
		return epochHeaderPlus1()
	case 402:
		return epochHeaderPlus2()
	case 403:
		return epochHeaderPlus3()
	}
	return nil
}
