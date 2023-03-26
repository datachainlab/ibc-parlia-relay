package main

import (
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(
		ethereum.Module{},
		module.Module{},
	); err != nil {
		log.Fatal(err)
	}
}
