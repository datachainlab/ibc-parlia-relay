package internal

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-hd-signer/pkg/hd"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"
)

func createRPCAddr() (string, error) {
	rpcAddr, ok := viper.Get("BSC_RPC_ADDR").(string)
	if !ok {
		return "", fmt.Errorf("BSC_RPC_ADDR is required")
	}
	return rpcAddr, nil
}

func CreateSignerConfig() *types.Any {
	signerConfig := &hd.SignerConfig{
		Mnemonic: hdwMnemonic,
		Path:     hdwPath,
	}
	anySignerConfig, err := types.NewAnyWithValue(signerConfig)
	if err != nil {
		panic(err)
	}
	return anySignerConfig
}

func createProver() (*module.Prover, module.Chain, error) {
	rpcAddr, err := createRPCAddr()
	if err != nil {
		return nil, nil, err
	}
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId: 9999,
		RpcAddr:    rpcAddr,
		Signer:     CreateSignerConfig(),
		IbcAddress: os.Getenv("BSC_IBC_ADDR"),
	})
	if err != nil {
		return nil, nil, err
	}

	config := module.ProverConfig{
		TrustingPeriod: 86400 * time.Second,
		MaxClockDrift:  1 * time.Millisecond,
	}
	ec := module.NewChain(chain)
	return module.NewProver(ec, &config).(*module.Prover), ec, nil
}
