# ibc-parlia-relay 

![CI](https://github.com/datachainlab/ibc-parlia-relay/workflows/CI/badge.svg?branch=main)

## Supported Versions
- [yui-relayer v0.5.3](https://github.com/hyperledger-labs/yui-relayer/releases/tag/v0.5.3)
- [ethereum-ibc-relay-chain v0.3.4](https://github.com/datachainlab/ethereum-ibc-relay-chain/releases/tag/v0.3.4)

## Setup Relayer

Add this module to [yui-relayer](https://github.com/hyperledger-labs/yui-relayer) and activate it.

```go
package main

import (
	"log"
	"github.com/hyperledger-labs/yui-relayer/cmd"
	parlia "github.com/datachainlab/ibc-parlia-relay/module"
)

func main() {
	if err := cmd.Execute(
		// counterparty.Module{}, //counter party
		parlia.Module{}, // Parlia Prover Module 
    ); err != nil {
		log.Fatal(err)
	}
}
```

## Change blocks per epoch 

* You can change blocks per epoch by build arguments.
* This is only for local net.
```
go build -tags dev -ldflags="-X github.com/datachainlab/ibc-parlia-relay/module/constant.blocksPerEpoch=20" -o testrly .
```

## Development

Generate proto buf with protobuf definition of [parlia-elc](https://github.com/datachainlab/parlia-elc).

```
cd $GOPATH/github.com/datachainlab
git clone https://github.com/datachainlab/parlia-elc
cd ibc-parlia-relay
make proto-import
make proto-gen
```
