#!/bin/bash

#rm -rf lcp-bridge && git clone https://github.com/datachainlab/lcp-bridge
#cd lcp-bridge
#git checkout 1e8ce53d582a87108485598874fec7b256d6df9a

## ETH chain
#SOLIDITY_PATH=$GOPATH/pkg/mod/github.com/hyperledger-labs/yui-ibc-solidity@v0.2.5-0.20230320070810-64b95470cabe
#cp -R $SOLIDITY_PATH/chains/geth ./development/chains/geth
#chmod 755 ./development/chains/geth/run.sh
#docker build -t eth_local ./development/chains/geth
#docker run --rm --name eth-local -d -p 8645:8545 -p 8646:8546 eth_local

## BSC chain
cd chains
make build
make network-all

## deploy contract
cd ../contracts
npm i
npx truffle migrate --network bsc_local2 --reset
npx truffle migrate --network bsc_local --reset

## setup relayer
cd ../
rm -rf $HOME/.yui-relayer
go run main.go config init
go run main.go chains add-dir config/demo/
go run main.go paths add ibc0 ibc1 ibc01 --file=config/path.json
jq <<< `cat $HOME/.yui-relayer/config/config.yaml`

## run cli
go run main.go tx clients ibc01
go run main.go tx update-clients ibc01
go run main.go tx connection ibc01
go run main.go tx channel ibc01

## start relayer
go run main.go service start ibc01

## test
cd contracts
npx truffle exec apps/0-init.js --network bsc_local2
