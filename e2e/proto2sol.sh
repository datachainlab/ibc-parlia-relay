SOLPB_DIR=../../solidity-protobuf
for file in $(find ../proto/ibc -name '*.proto')
do
  echo "Generating "$file
  protoc -I$(pwd)/contracts/proto -I$(pwd)/../proto -I${SOLPB_DIR}/protobuf-solidity/src/protoc/include --plugin=protoc-gen-sol=${SOLPB_DIR}/protobuf-solidity/src/protoc/plugin/gen_sol.py --"sol_out=ignore_protos=gogoproto/gogo.proto&use_runtime=@hyperledger-labs/yui-ibc-solidity/contracts/proto/ProtoBufRuntime.sol&solc_version=0.8.9:$(pwd)/contracts/contracts" $(pwd)/$file
done
echo "add import duration"
SOL_PARLIA=contracts/contracts/ibc/lightclients/parlia/v1/parlia.sol
sed -i '' -e "s/client.sol\"\;/client.sol\"\;\\nimport \".\/duration.sol\"\;/" $SOL_PARLIA
