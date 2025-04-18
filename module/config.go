package module

import (
	"fmt"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/coreutil"
	"github.com/hyperledger-labs/yui-relayer/otelcore"
)

var _ core.ProverConfig = (*ProverConfig)(nil)

func (c *ProverConfig) Build(chain core.Chain) (core.Prover, error) {
	chain_, err := coreutil.UnwrapChain[*ethereum.Chain](chain)
	if err != nil {
		return nil, err
	}
	// Use chain, not chain_, for the case where the chain is wrapped by another struct that implements core.Chain (e.g. tracing bridge)
	return otelcore.NewProver(NewProver(NewChain(chain, chain_.Config().IBCAddress(), chain_.Client()), c), chain.ChainID(), tracer), nil
}

func (c *ProverConfig) Validate() error {
	if GetForkParameters(Network(c.Network)) == nil {
		return fmt.Errorf("unknown network: %s", c.Network)
	}
	return nil
}
