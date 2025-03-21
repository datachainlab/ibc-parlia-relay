package membership

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	types3 "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/tool/testdata/internal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/spf13/cobra"
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
			stateRoot, proof, proofHeight, err := m.proveState(cmd.Context(), chainID, path, commitment)
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

func (m *verifyMembershipModule) proveState(ctx context.Context, chainID uint64, path string, value []byte) (common.Hash, []byte, types.Height, error) {
	rpcAddr := os.Getenv("BSC_RPC_ADDR")
	ibcAddress := os.Getenv("BSC_IBC_ADDR")
	log.Println(rpcAddr, ibcAddress)
	chain, err := ethereum.NewChain(ctx, ethereum.ChainConfig{
		EthChainId: chainID,
		RpcAddr:    rpcAddr,
		Signer:     internal.CreateSignerConfig(),
		IbcAddress: ibcAddress,
	})
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}
	latest, err := chain.LatestHeight(ctx)
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}
	config := module.ProverConfig{}
	prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

	queryCtx := core.NewQueryContext(ctx, latest)

	header, err := chain.Client().HeaderByNumber(ctx, big.NewInt(int64(latest.GetRevisionHeight())))
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}

	proof, proofHeight, err := prover.ProveState(queryCtx, path, value)
	storageRoot, err := prover.GetStorageRoot(ctx, header)
	if err != nil {
		return common.Hash{}, nil, types.Height{}, err
	}
	log.Println("storageRoot", storageRoot)

	return header.Root, proof, proofHeight, err
}

func CreateVerifyMembership() *cobra.Command {
	cmd := &cobra.Command{
		Use: "verify_membership",
	}
	m := verifyMembershipModule{}
	cmd.AddCommand(m.latest())
	return cmd
}
