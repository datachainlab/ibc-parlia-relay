#!/bin/bash

rm -rf lcp-bridge && git clone https://github.com/datachainlab/lcp-bridge
cd lcp-bridge
make build
make network
cd contracts
npm i
npx truffle migrate --network bsc_local
npx truffle migrate --network eth_local
rm -rf $HOME/.yui-relayer
go run main.go config init
go run main.go chains add-dir config/demo/
go run main.go paths add ibc0 ibc1 ibc01 --file=config/path.json

## setup
go run main.go tx clients ibc01