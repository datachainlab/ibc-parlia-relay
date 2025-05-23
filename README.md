# ibc-parlia-relay 

![CI](https://github.com/datachainlab/ibc-parlia-relay/workflows/CI/badge.svg?branch=main)

## Supported Versions
- [yui-relayer v0.5.11](https://github.com/hyperledger-labs/yui-relayer/releases/tag/v0.5.11)
- [ethereum-ibc-relay-chain v0.3.17](https://github.com/datachainlab/ethereum-ibc-relay-chain/releases/tag/v0.3.7)
- [parlia-elc v0.3.10](https://github.com/datachainlab/parlia-elc/releases/tag/v0.3.10)

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

1. Set HF height as soon as possible
As soon as the HF height is determined, please modify the timestamp in the ForkSpec to the height as soon as possible.
HF height is calculated from timestamp, but the further away from the HF, the longer it takes to calculate.

2. Limitation of the CreateClient
When the latest HF height is not set it is impossible to create client if the latest finalize header is after latest HF timestamp