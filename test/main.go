package main

import (
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"log"

	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/hyperledger-labs/yui-relayer/cmd"
	mock "github.com/hyperledger-labs/yui-relayer/provers/mock/module"
)

func main() {
	if err := cmd.Execute(
		mock.Module{},
		ethereum.Module{},
		module.Module{},
	); err != nil {
		log.Fatal(err)
	}
}
