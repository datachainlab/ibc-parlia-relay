package internal

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func CreateCreateClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create testdata for Create client. ",
	}
	cmd.AddCommand(CreateClientSuccessCmd())
	return cmd
}

func CreateClientSuccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "success",
		Short: "create CreateClient testdata for success",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, chain, err := createMainnetProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			var commitmentsSlot [32]byte
			// create initial client state
			latestHeight := types.NewHeight(latest.GetRevisionNumber(), latest.GetRevisionHeight())
			clientState := module.ClientState{
				TrustingPeriod:     100 * time.Second,
				MaxClockDrift:      1 * time.Millisecond,
				ChainId:            56,
				LatestHeight:       &latestHeight,
				Frozen:             false,
				IbcStoreAddress:    common.HexToAddress(MainNetIbcAddress).Bytes(),
				IbcCommitmentsSlot: commitmentsSlot[:],
			}
			anyClientState, err := codectypes.NewAnyWithValue(&clientState)
			if err != nil {
				return err
			}
			csb, err := anyClientState.Marshal()
			if err != nil {
				return err
			}
			log.Println("clientState", common.Bytes2Hex(csb))
			log.Println("height", latestHeight)
			return nil
		},
	}
	return cmd
}
