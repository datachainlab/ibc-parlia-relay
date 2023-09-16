package internal

import (
	"github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

type historyModule struct {
}

func (m *historyModule) mainnet() *cobra.Command {
	var num uint64
	cmd := &cobra.Command{
		Use:   "mainnet",
		Short: "create many data testdata",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Printf("num = %d\n", num)
			prover, chain, err := createMainnetProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}

			createdEpoch, err := m.outputMsgClient(prover, latest.GetRevisionHeight()-num)
			if err != nil {
				return err
			}

			return m.outputMsgUpdate(prover, createdEpoch, latest.GetRevisionHeight(), num)
		},
	}
	cmd.Flags().Uint64Var(&num, "num", 240, "--num")
	return cmd
}

func (m *historyModule) outputMsgUpdate(prover *module.Prover, createdEpoch, latest uint64, num uint64) error {
	type updatingData struct {
		Header string `json:"header"`
	}

	type updates struct {
		Data []updatingData `json:"data"`
	}

	data := updates{}
	for i := num; i > 0; i-- {
		targetLatest := latest - i
		header, err := prover.GetLatestFinalizedHeaderByLatestHeight(targetLatest)
		if err != nil {
			return err
		}
		target, err := header.(*module.Header).DecodedTarget()
		if err != nil {
			return err
		}
		clientStateLatestHeight := types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1)
		if i == num {
			clientStateLatestHeight = types.NewHeight(header.GetHeight().GetRevisionNumber(), createdEpoch)
		}

		blocks, _ := prover.SetupHeadersForUpdateByLatestHeight(clientStateLatestHeight, header.(*module.Header))

		for _, block := range blocks {
			log.Printf("finalzied=%d\n", block.GetHeight())
			pack, err := types.PackClientMessage(block)
			if err != nil {
				return err
			}
			marshal, err := pack.Marshal()
			if err != nil {
				return err
			}
			data.Data = append(data.Data, updatingData{Header: common.Bytes2Hex(marshal)})
		}
		time.Sleep(200 * time.Millisecond)
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile("update_mainnet.json", serialized, 777)
}

func (m *historyModule) outputMsgClient(prover *module.Prover, firstNumber uint64) (uint64, error) {
	firstHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(firstNumber)
	if err != nil {
		return 0, err
	}
	type createData struct {
		ClientState    string `json:"clientState"`
		ConsensusState string `json:"consensusState"`
	}
	msgCreateClient, err := prover.CreateMsgCreateClient("", firstHeader, nil)
	if err != nil {
		return 0, err
	}
	anyClientState, err := msgCreateClient.ClientState.Marshal()
	if err != nil {
		return 0, err
	}
	anyConsState, err := msgCreateClient.ConsensusState.Marshal()
	if err != nil {
		return 0, err
	}
	creating := createData{
		ClientState:    common.Bytes2Hex(anyClientState),
		ConsensusState: common.Bytes2Hex(anyConsState),
	}
	serialized, err := json.Marshal(creating)
	if err != nil {
		return 0, err
	}
	epochs := firstHeader.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
	return (epochs - 1) * constant.BlocksPerEpoch, os.WriteFile("create_mainnet.json", serialized, 777)
}

func CreateHistoryClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Create testdata for update client. ",
	}
	m := historyModule{}
	cmd.AddCommand(m.mainnet())
	return cmd
}
