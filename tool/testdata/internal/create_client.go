package internal

import (
	"context"
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
			cs, consState, err := prover.CreateInitialLightClientState(context.Background(), nil)
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
			currentValidatorSet, currentTurnLength, err := module.QueryValidatorSetAndTurnLength(chain.Header, module.GetCurrentEpoch(cs.GetLatestHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			previousValidatorSet, previousTurnLength, err := module.QueryValidatorSetAndTurnLength(chain.Header, module.GetPreviousEpoch(cs.GetLatestHeight().GetRevisionHeight()))
			if err != nil {
				return err
			}
			log.Println("clientState", common.Bytes2Hex(anyClientState))
			log.Println("consensusState", common.Bytes2Hex(anyConsState))
			log.Println("height", cs.GetLatestHeight().GetRevisionHeight())
			log.Println("time", consState.GetTimestamp())
			log.Println("currentEpochHash", common.BytesToHash(module.MakeEpochHash(currentValidatorSet, currentTurnLength)))
			log.Println("previousEpochHash", common.BytesToHash(module.MakeEpochHash(previousValidatorSet, previousTurnLength)))
			log.Println("storageRoot", common.BytesToHash(consState.(*module.ConsensusState).StateRoot))

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
