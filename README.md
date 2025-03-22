# ibc-parlia-relay 

![CI](https://github.com/datachainlab/ibc-parlia-relay/workflows/CI/badge.svg?branch=main)

## Supported Versions
- [yui-relayer v0.5.11](https://github.com/hyperledger-labs/yui-relayer/releases/tag/v0.5.11)
- [ethereum-ibc-relay-chain v0.3.16](https://github.com/datachainlab/ethereum-ibc-relay-chain/releases/tag/v0.3.6)

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

## Development

Generate proto buf with protobuf definition of [parlia-elc](https://github.com/datachainlab/parlia-elc).

```
cd $GOPATH/github.com/datachainlab
git clone https://github.com/datachainlab/parlia-elc
cd ibc-parlia-relay
make proto-import
make proto-gen
```

## About ForkSpec

As soon as the HF height is determined, please modify the timestamp in the ForkSpec to the height as soon as possible.
HF height is calculated from timestamp, but the further away from the HF, the longer it takes to calculate.
