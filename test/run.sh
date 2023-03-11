#!/bin/bash

rm -rf lcp-bridge && git clone https://github.com/datachainlab/lcp-bridge
cd lcp-bridge

## ETH chain
SOLIDITY_PATH=$GOPATH/pkg/mod/github.com/hyperledger-labs/yui-ibc-solidity@v0.2.5-0.20221130073947-3315e5fa0f5b
cp -R $SOLIDITY_PATH/chains/geth ./development/chains/geth
chmod 755 ./development/chains/geth/run.sh
docker build -t eth_local ./development/chains/geth
docker run --rm --name eth-local -d -p 8645:8545 -p 8646:8546 eth_local

## BSC chain
make build
make -C development/chains/bsc network-all

## deploy contract
cd contracts
cp ../../config/truffle-config.js .
cp ../../config/4_initialize_parlia.js ./migrations/
rm -rf build
rm -rf contracts
rm -rf migrations
cp -R $SOLIDITY_PATH/contracts contracts
cp -R $SOLIDITY_PATH/migrations migrations
npm i
npx truffle build
npx truffle migrate --network eth_local
npx truffle migrate --network bsc_local

## setup relayer
cd ../../
rm -rf $HOME/.yui-relayer
go run main.go config init
go run main.go chains add-dir config/demo/
go run main.go paths add ibc0 ibc1 ibc01 --file=config/path.json
jq <<< `cat $HOME/.yui-relayer/config/config.yaml`

## run cli
go run main.go tx clients ibc01