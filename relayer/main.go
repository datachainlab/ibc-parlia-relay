package main

import (
	"log"

	ethereumlc "github.com/datachainlab/ethereum-lcp/go/relay"
	"github.com/datachainlab/ibc-parlia-relay/module"
	lcp "github.com/datachainlab/lcp/go/relay"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/cmd"
)

func main() {
	if err := cmd.Execute(
		ethereum.Module{},
		ethereumlc.Module{},
		module.Module{},
		lcp.Module{},
	); err != nil {
		log.Fatal(err)
	}
}
