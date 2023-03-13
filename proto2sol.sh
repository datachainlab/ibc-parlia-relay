SOLPB_DIR=../solidity-protobuf
for file in $(find ./proto/ibc/lightclients -name '*.proto')
do
  echo "Generating "$file
  protoc -I$(pwd)/proto -I${SOLPB_DIR}/protobuf-solidity/src/protoc/include --plugin=protoc-gen-sol=${SOLPB_DIR}/protobuf-solidity/src/protoc/plugin/gen_sol.py --"sol_out=use_runtime=@hyperledger-labs/yui-ibc-solidity/contracts/proto/ProtoBufRuntime.sol&solc_version=0.8.9:$(pwd)/test/contracts/contracts" $(pwd)/$file
done