package module

import (
	"context"
	"math/big"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/core"
)

type Chain interface {
	core.Chain
	Header(ctx context.Context, height uint64) (*types.Header, error)
	IBCAddress() common.Address
	CanonicalChainID(ctx context.Context) (uint64, error)
	GetProof(ctx context.Context, address common.Address, storageKeys [][]byte, blockNumber *big.Int) (*client.StateProof, error)
}

type ethChain struct {
	core.Chain
	ibcAddress common.Address
	client     *client.ETHClient
}

func NewChain(chain core.Chain, ibcAddress common.Address, client *client.ETHClient) Chain {
	return &ethChain{Chain: chain, ibcAddress: ibcAddress, client: client}
}

func (c *ethChain) Header(ctx context.Context, height uint64) (*types.Header, error) {
	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(height)))
	if err != nil {
		return nil, err
	}
	return block.Header(), nil
}

func (c *ethChain) IBCAddress() common.Address {
	return c.ibcAddress
}

func (c *ethChain) CanonicalChainID(ctx context.Context) (uint64, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return chainID.Uint64(), nil
}

func (c *ethChain) GetProof(ctx context.Context, address common.Address, storageKeys [][]byte, blockNumber *big.Int) (*client.StateProof, error) {
	return c.client.GetProof(ctx, address, storageKeys, blockNumber)
}
