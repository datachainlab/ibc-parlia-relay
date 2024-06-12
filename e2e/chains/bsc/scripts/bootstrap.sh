#!/usr/bin/env bash

ls -ltr /root/genesis
ls -ltr /root/genesis/scripts

workspace=$(
  cd $(dirname $0)
  pwd
)/..

function prepare() {
  if ! [[ -f /usr/local/bin/geth ]]; then
    echo "geth do not exist!"
    exit 1
  fi
  # run `make clean` to remove files
  cd ${workspace}/genesis
  rm -rf validators.conf
}

function init_validator() {
  node_id=$1

  # set validator address
  mkdir -p ${workspace}/storage/${node_id}/keystore
  cp ${workspace}/validators/keystore/${node_id} ${workspace}/storage/${node_id}/keystore/${node_id}
  validatorAddr=0x$(cat ${workspace}/storage/${node_id}/keystore/${node_id} | jq .address | sed 's/"//g')
  echo ${validatorAddr} >${workspace}/storage/${node_id}/address

  # create new BLS vote address
  expect ${workspace}/scripts/create_bls_key.sh ${workspace}/storage/${node_id}
  voteAddr=0x$(cat ${workspace}/storage/${node_id}/bls/keystore/*json| jq .pubkey | sed 's/"//g')
  echo $voteAddr

  echo "${validatorAddr},${validatorAddr},${validatorAddr},0x0000000010000000,${voteAddr}" >>${workspace}/genesis/scripts/validators.conf
}

function generate_genesis() {
  INIT_HOLDER_ADDRESSES=$(ls ${workspace}/init-holders | tr '\n' ',')
  INIT_HOLDER_ADDRESSES=${INIT_HOLDER_ADDRESSES/%,/}
  echo "blocks per epoch = ${BLOCKS_PER_EPOCH}"

  echo "replace genesis-template.template"
  sed "s/{{BLOCKS_PER_EPOCH}}/${BLOCKS_PER_EPOCH}/g" ${workspace}/genesis/genesis-template.template >${workspace}/genesis/genesis-template.json

  echo "replace init_holders.template"
  sed "s/{{INIT_HOLDER_ADDRESSES}}/${INIT_HOLDER_ADDRESSES}/g" ${workspace}/genesis/scripts/init_holders.template | sed "s/{{INIT_HOLDER_BALANCE}}/${INIT_HOLDER_BALANCE}/g" >${workspace}/genesis/scripts/init_holders.js

  echo "replace generate.py"
  sed "s/{{BLOCKS_PER_EPOCH}}/${BLOCKS_PER_EPOCH}/g" ${workspace}/genesis/scripts/generate.py >${workspace}/genesis/scripts/generate.py.out
  sed "s/{{BSC_CHAIN_ID}}/${BSC_CHAIN_ID}/g" ${workspace}/genesis/scripts/generate.py.out >${workspace}/genesis/scripts/generate.py

  echo "replace BSCValidatorSet"
  sed "s/{{INIT_NUM_OF_CABINETS}}/${INIT_NUM_OF_CABINETS}/g" ${workspace}/genesis/contracts/BSCValidatorSet.sol >${workspace}/genesis/contracts/BSCValidatorSet.sol.out
  mv ${workspace}/genesis/contracts/BSCValidatorSet.sol.out ${workspace}/genesis/contracts/BSCValidatorSet.sol

  node scripts/generate-validator.js
  chainIDHex=$(printf '%04x\n' ${BSC_CHAIN_ID})

  echo "start generate process"
  /root/.local/bin/poetry run python3 scripts/generate.py dev
  #node scripts/generate-genesis.js --chainId ${BSC_CHAIN_ID}
}

function init_genesis_data() {
  echo "start to initialize genesis data"
  node_type=$1
  node_id=$2
  geth --datadir ${workspace}/storage/${node_id} init --state.scheme hash --db.engine=leveldb ${workspace}/genesis/genesis.json
  cp ${workspace}/config/config-${node_type}.toml ${workspace}/storage/${node_id}/config.toml
  sed -i -e "s/{{NetworkId}}/${BSC_CHAIN_ID}/g" ${workspace}/storage/${node_id}/config.toml
  if [ "${node_id}" == "bsc-rpc" ]; then
    ls -ltr ${workspace}/storage/${node_id}/keystore
    cp ${workspace}/init-holders/* ${workspace}/storage/${node_id}/keystore
    cp ${workspace}/genesis/genesis.json ${workspace}/storage/${node_id}
    cp ${workspace}/config/bootstrap.key ${workspace}/storage/${node_id}/geth/nodekey
  fi
}

prepare

# First, generate config for each validator
for ((i = 1; i <= ${NUMS_OF_VALIDATOR}; i++)); do
  init_validator "bsc-validator${i}"
done

# Then, use validator configs to generate genesis file
generate_genesis

# Finally, use genesis file to init cluster data
init_genesis_data bsc-rpc bsc-rpc

for ((i = 1; i <= ${NUMS_OF_VALIDATOR}; i++)); do
  init_genesis_data validator "bsc-validator${i}"
done
