# ibc-parlia-relay

![CI](https://github.com/datachainlab/ibc-parlia-relay/workflows/CI/badge.svg?branch=main)

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

## Change luban fork

* You can change block for luban fork
* This is only for local net.
```
go build -tags dev -ldflags="-X github.com/datachainlab/ibc-parlia-relay/module/constant.lubanFork=6" -o testrly .
```

## Development

Generate proto buf with protobuf definition of [lcp-parlia](https://github.com/datachainlab/lcp-parlia).

```
cd $GOPATH/github.com/datachainlab
git clone https://github.com/datachainlab/lcp-parlia
cd ibc-parlia-relay
make proto-gen
```
