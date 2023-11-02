package membership

import (
	"context"
	"fmt"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	types3 "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/spf13/cobra"
	"log"
	"math/big"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"
	ibcAddress  = "0x702E40245797c5a2108A566b3CE2Bf14Bc6aF841"
)

type verifyMembershipModule struct {
}

func (m *verifyMembershipModule) latest() *cobra.Command {
	return &cobra.Command{
		Use: "latest",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID := uint64(9999)
			connection := conntypes.ConnectionEnd{
				ClientId: "xx-parlia-0",
				Versions: []*conntypes.Version{{
					Identifier: "1",
					Features:   []string{"ORDER_ORDERED", "ORDER_UNORDERED"},
				}},
				State: conntypes.OPEN,
				Counterparty: conntypes.Counterparty{
					ClientId:     "xx-parlia-0",
					ConnectionId: "connection-0",
					Prefix: types3.MerklePrefix{
						KeyPrefix: []byte("ibc"),
					},
				},
				DelayPeriod: 0,
			}

			commitment, err := connection.Marshal()
			if err != nil {
				return err
			}

			path := host.ConnectionPath("connection-0")
			stateRoot, proof, proofHeight, err := m.proveState(chainID, 8645, path, commitment)
			if err != nil {
				log.Panic(err)
			}
			log.Println("proofHeight", proofHeight)
			log.Println("proof", common.Bytes2Hex(proof))
			log.Println("stateRoot", stateRoot)
			log.Println("value", common.Bytes2Hex(commitment))
			return nil
		},
	}
}

func (m *verifyMembershipModule) proveState(chainID uint64, port int64, path string, value []byte) (common.Hash, []byte, types.Height, error) {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  chainID,
		RpcAddr:     fmt.Sprintf("http://localhost:%d", port),
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  ibcAddress,
	})
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}
	latest, err := chain.LatestHeight()
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}
	config := module.ProverConfig{}
	prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

	ctx := core.NewQueryContext(context.Background(), latest)

	header, err := chain.Client().HeaderByNumber(ctx.Context(), big.NewInt(int64(latest.GetRevisionHeight())))
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}

	stateRoot, err := prover.GetStorageRoot(header)
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}

	proof, proofHeight, err := prover.ProveState(ctx, path, value)
	return stateRoot, proof, proofHeight, err
}

func CreateVerifyMembership() *cobra.Command {
	cmd := &cobra.Command{
		Use: "verify_membership",
	}
	m := verifyMembershipModule{}
	cmd.AddCommand(m.latest())
	return cmd
}
