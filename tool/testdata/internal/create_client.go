package internal

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
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
			cs, consState, err := prover.CreateInitialLightClientState(nil)
			if err != nil {
				return err
			}

			protoClientState, err := codectypes.NewAnyWithValue(cs.(proto.Message))
			if err != nil {
				return err
			}
			protoConsState, err := codectypes.NewAnyWithValue(consState.(proto.Message))
			if err != nil {
				return err
			}
			anyClientState, err := protoClientState.Marshal()
			if err != nil {
				return err
			}
			anyConsState, err := protoConsState.Marshal()
			if err != nil {
				return err
			}
			currentValidatorSet, currentTurnTerm, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetCurrentEpoch(cs.GetLatestHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			previousValidatorSet, previousTurnTerm, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetPreviousEpoch(cs.GetLatestHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			log.Println("clientState", common.Bytes2Hex(anyClientState))
			log.Println("consensusState", common.Bytes2Hex(anyConsState))
			log.Println("height", cs.GetLatestHeight().GetRevisionHeight())
			log.Println("time", consState.GetTimestamp())
			log.Println("currentEpochHash", module.MakeEpochHash(currentValidatorSet, currentTurnTerm))
			log.Println("previousEpochHash", module.MakeEpochHash(previousValidatorSet, previousTurnTerm))
			log.Println("storageRoot", consState.(*module.ConsensusState).StateRoot)

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
