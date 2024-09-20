package main

import (
	"log"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-hd-signer/pkg/hd"
	"github.com/hyperledger-labs/yui-relayer/cmd"

	"github.com/datachainlab/ibc-parlia-relay/module"
)

func main() {
	if err := cmd.Execute(
		ethereum.Module{},
		module.Module{},
		hd.Module{},
	); err != nil {
		log.Fatal(err)
	}
}
