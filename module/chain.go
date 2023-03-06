package module

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/client"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
	"math/big"
)

type ChainI interface {
	core.ChainI
	LatestHeight(ctx context.Context) (uint64, error)
	LatestLightHeight(ctx context.Context) (uint64, error)
	Header(ctx context.Context, height uint64) (*types.Header, error)
	IBCHandlerAddress() common.Address
	CanonicalChainID(ctx context.Context) (uint64, error)
	GetStateProof(address common.Address, storageKeys [][]byte, blockNumber *big.Int) (*client.StateProof, error)
}

type ethChain struct {
	*ethereum.Chain
}

func NewChain(chain *ethereum.Chain) ChainI {
	return &ethChain{Chain: chain}
}

func (c *ethChain) LatestHeight(ctx context.Context) (uint64, error) {
	return c.Client().BlockNumber(ctx)
}

func (c *ethChain) LatestLightHeight(ctx context.Context) (uint64, error) {
	//TODO from LCP
	return c.Client().BlockNumber(ctx)
}

func (c *ethChain) Header(ctx context.Context, height uint64) (*types.Header, error) {
	block, err := c.Client().BlockByNumber(ctx, big.NewInt(int64(height)))
	if err != nil {
		return nil, err
	}
	return block.Header(), nil
}

func (c *ethChain) IBCHandlerAddress() common.Address {
	return c.Config().IBCHandlerAddress()
}

func (c *ethChain) CanonicalChainID(ctx context.Context) (uint64, error) {
	chainID, err := c.Client().ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return chainID.Uint64(), nil
}

func (c *ethChain) GetStateProof(address common.Address, storageKeys [][]byte, blockNumber *big.Int) (*client.StateProof, error) {
	return c.Client().GetStateProof(address, storageKeys, blockNumber)
}
