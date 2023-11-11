package internal

import (
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"log"
)

type createClientModule struct {
}

func (m *createClientModule) createClientSuccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "success",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return err
			}
			header, err := prover.GetLatestFinalizedHeader()
			message, err := prover.CreateMsgCreateClient("", header, common.HexToAddress(mainAndTestNetIbcAddress).Bytes())
			if err != nil {
				return err
			}
			clientState, err := message.ClientState.Marshal()
			if err != nil {
				return err
			}
			consState, err := message.ConsensusState.Marshal()
			if err != nil {
				return err
			}
			currentValidatorSet, err := module.QueryValidatorSet(chain.Header, module.GetCurrentEpoch(header.GetHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			previousValidatorSet, err := module.QueryValidatorSet(chain.Header, module.GetPreviousEpoch(header.GetHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			target, err := header.(*module.Header).Target()
			if err != nil {
				return err
			}
			storageRoot, err := prover.GetStorageRoot(target)
			log.Println("clientState", common.Bytes2Hex(clientState))
			log.Println("consensusState", common.Bytes2Hex(consState))
			log.Println("height", target.Number)
			log.Println("time", target.Time)
			log.Println("currentValidatorSet", common.BytesToHash(crypto.Keccak256(currentValidatorSet...)))
			log.Println("previousValidatorSet", common.Bytes2Hex(crypto.Keccak256(previousValidatorSet...)))
			log.Println("storageRoot", storageRoot)

			return nil
		},
	}
	return cmd
}

func CreateCreateClient() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
	}
	m := createClientModule{}
	cmd.AddCommand(m.createClientSuccessCmd())
	return cmd
}
