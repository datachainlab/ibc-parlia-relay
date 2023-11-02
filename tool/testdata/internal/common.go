package internal

import (
	"fmt"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/spf13/viper"
	"time"
)

const (
	hdwMnemonic              = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath                  = "m/44'/60'/0'/0/0"
	ibcAddress               = "0x702E40245797c5a2108A566b3CE2Bf14Bc6aF841"
	localNetValidatorSize    = 3
	mainNetValidatorSize     = 21
	mainAndTestNetIbcAddress = "0x151f3951FA218cac426edFe078fA9e5C6dceA500"
)

func createRPCAddr() (string, error) {
	rpcAddr, ok := viper.Get("BSC_RPC_ADDR").(string)
	if !ok {
		return "", fmt.Errorf("BSC_RPC_ADDR is required")
	}
	return rpcAddr, nil
}

func createProver() (*module.Prover, core.Chain, error) {
	rpcAddr, err := createRPCAddr()
	if err != nil {
		return nil, nil, err
	}
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  56,
		RpcAddr:     rpcAddr,
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  mainAndTestNetIbcAddress,
	})
	if err != nil {
		return nil, chain, err
	}

	config := module.ProverConfig{
		TrustingPeriod: 86400 * time.Second,
		MaxClockDrift:  1 * time.Millisecond,
	}
	ec := module.NewChain(chain)
	return module.NewProver(ec, &config).(*module.Prover), chain, nil
}
