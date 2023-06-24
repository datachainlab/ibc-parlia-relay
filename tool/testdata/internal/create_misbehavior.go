package internal

import (
	"fmt"
	"github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/spf13/cobra"
	"log"
)

const (
	hdwMnemonic = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath     = "m/44'/60'/0'/0/0"
	IbcAddress  = "0x702E40245797c5a2108A566b3CE2Bf14Bc6aF841"
)

func CreateMisbehavior() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehavior",
		Short: "Create testdata for misbehavior. ",
	}
	cmd.AddCommand(localNet())
	return cmd
}

// Launch local net before execute.
//   - make chain
//     -> Change NUMS_OF_VALIDATOR to 1 in docker-compose.yml
//     -> Use e2e/chains/bsc/validators/keystore/bsc-validator1 for each chain's validator
//   - make contracts
func localNet() *cobra.Command {
	return &cobra.Command{
		Use:   "local",
		Short: "create misbehavior testdata with local net",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID := int64(9999)
			height := uint64(200)
			header1, err := getHeader(chainID, 8645, height)
			if err != nil {
				log.Panic(err)
			}
			header2, err := getHeader(chainID, 8545, height)
			if err != nil {
				log.Panic(err)
			}
			clientId := "xx-parlia-1"
			misbehavior := module.Misbehaviour{
				ClientId: clientId,
				Header_1: header1,
				Header_2: header2,
			}

			// print hex for lcp-parlia test
			pack, _ := types.PackMisbehaviour(&misbehavior)
			marshal, _ := pack.Marshal()
			log.Println(common.Bytes2Hex(marshal))
			return nil
		},
	}
}

func getHeader(chainID int64, port int64, latestBlockNumber uint64) (*module.Header, error) {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  chainID,
		RpcAddr:     fmt.Sprintf("http://localhost:%d", port),
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  IbcAddress,
	})
	if err != nil {
		return nil, err
	}

	config := module.ProverConfig{
		Debug: true,
	}
	prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

	// Get Finalized header
	latestHeight := types.NewHeight(0, latestBlockNumber)
	latest := latestHeight.GetRevisionHeight()
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
	if err != nil {
		return nil, err
	}
	header := iHeader.(*module.Header)
	target, err := header.Target()
	if err != nil {
		return nil, err
	}

	// Setup finalized header
	updating, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header)
	if err != nil {
		return nil, err
	}
	return updating[0].(*module.Header), nil
}
