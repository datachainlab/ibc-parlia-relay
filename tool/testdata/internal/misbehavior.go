package internal

import (
	"fmt"
	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
	"log"
	"os"
)

type misbehaviorModule struct {
}

func (m *misbehaviorModule) success() *cobra.Command {
	return &cobra.Command{
		Use: "success",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID := uint64(9999)
			targetHeight, header1, err := m.getLocalHeader(chainID, 8645, 0, 1)
			if err != nil {
				log.Panic(err)
			}
			_, header2, err := m.getLocalHeader(chainID, 8545, targetHeight, 2)
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
			pack, _ := types.PackClientMessage(&misbehavior)
			marshal, _ := pack.Marshal()
			log.Println("misbehavior", common.Bytes2Hex(marshal))
			log.Println("h1_height", header1.GetHeight())
			log.Println("h1_trustedHeight", header1.TrustedHeight)
			log.Println("h2_height", header2.GetHeight())
			log.Println("h2_trustedHeight", header2.TrustedHeight)
			return nil
		},
	}
}

func (m *misbehaviorModule) error() *cobra.Command {
	return &cobra.Command{
		Use: "error",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return err
			}

			latestHeight, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			latest := latestHeight.GetRevisionHeight()
			println(latest)
			header, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
			target, err := header.(*module.Header).Target()
			updating, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header.(*module.Header))
			if err != nil {
				return err
			}

			// Exactly same block
			misbehavior := module.Misbehaviour{
				ClientId: "xx-parlia-1",
				Header_1: updating[0].(*module.Header),
				Header_2: updating[0].(*module.Header),
			}
			pack, _ := types.PackClientMessage(&misbehavior)
			marshal, _ := pack.Marshal()
			log.Println("Exactly same block: misbehavior1", common.Bytes2Hex(marshal[0:len(marshal)/2]))
			log.Println("Exactly same block: misbehavior2", common.Bytes2Hex(marshal[len(marshal)/2:]))
			log.Println("Exactly same block: height", target.Number.Int64())

			// Invalid block
			header2, _ := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
			updating2, _ := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header2.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header2.(*module.Header))
			target2, _ := updating2[0].(*module.Header).Target()
			target2.Root = common.Hash{}
			rlpTarget, err := rlp.EncodeToBytes(target2)
			updating2[0].(*module.Header).Headers[0] = &module.ETHHeader{Header: rlpTarget}
			misbehavior2 := module.Misbehaviour{
				ClientId: "xx-parlia-1",
				Header_1: updating[0].(*module.Header),
				Header_2: updating2[0].(*module.Header),
			}
			pack, _ = types.PackClientMessage(&misbehavior2)
			marshal, _ = pack.Marshal()
			log.Println("Invalid block: misbehavior1", common.Bytes2Hex(marshal[0:len(marshal)/2]))
			log.Println("Invalid block: misbehavior2", common.Bytes2Hex(marshal[len(marshal)/2:]))
			log.Println("Invalid block: height", header.GetHeight())
			log.Println("Invalid block: current_validator_hash", common.Bytes2Hex(module.MakeEpochHash(header.(*module.Header).CurrentValidators, uint8(header.(*module.Header).CurrentTurnLength))))
			log.Println("Invalid block: previous_validator_hash", common.Bytes2Hex(module.MakeEpochHash(header.(*module.Header).PreviousValidators, uint8(header.(*module.Header).PreviousTurnLength))))
			log.Println("Invalid block: trusted_height", updating[0].(*module.Header).TrustedHeight)
			epochCount := header.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
			log.Println("Invalid block: currentEpoch", epochCount*constant.BlocksPerEpoch)
			return nil
		},
	}
}

func (m *misbehaviorModule) getLocalHeader(chainID uint64, port int64, targetHeight uint64, trustedDiff uint64) (uint64, *module.Header, error) {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId: chainID,
		RpcAddr:    fmt.Sprintf("http://localhost:%d", port),
		Signer:     CreateSignerConfig(),
		IbcAddress: os.Getenv("BSC_IBC_ADDR"),
	})
	if err != nil {
		return targetHeight, nil, err
	}
	if targetHeight == 0 {
		latest, err := chain.LatestHeight()
		if err != nil {
			return targetHeight, nil, err
		}
		targetHeight = latest.GetRevisionHeight() - 1
	}
	config := module.ProverConfig{}
	prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

	// Get Finalized header
	latestHeight := types.NewHeight(0, targetHeight)
	latest := latestHeight.GetRevisionHeight()
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
	if err != nil {
		return latest, nil, err
	}
	header := iHeader.(*module.Header)
	target, err := header.Target()
	if err != nil {
		return latest, nil, err
	}
	trustedHeight := types.NewHeight(0, target.Number.Uint64()-trustedDiff)
	header.TrustedHeight = &trustedHeight
	return latest, header, nil
}

func CreateMisbehavior() *cobra.Command {
	cmd := &cobra.Command{
		Use: "misbehavior",
	}
	m := misbehaviorModule{}
	cmd.AddCommand(m.success())
	cmd.AddCommand(m.error())
	return cmd
}
