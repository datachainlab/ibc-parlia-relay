package internal

import (
	"github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module"
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
		Use: "mainnet",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Printf("num = %d\n", num)
			prover, chain, err := createProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}

			createdEpoch, err := m.outputMsgClient(prover, latest.GetRevisionHeight()-num, "create_mainnet.json")
			if err != nil {
				return err
			}

			return m.outputMsgUpdate(prover, createdEpoch, latest.GetRevisionHeight(), num, "update_mainnet.json")
		},
	}
	cmd.Flags().Uint64Var(&num, "num", 240, "--num")
	return cmd
}

func (m *historyModule) testnet() *cobra.Command {
	var num uint64
	cmd := &cobra.Command{
		Use: "testnet",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Printf("num = %d\n", num)
			prover, chain, err := createProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}

			createdEpoch, err := m.outputMsgClient(prover, latest.GetRevisionHeight()-num, "create_testnet.json")
			if err != nil {
				return err
			}

			return m.outputMsgUpdate(prover, createdEpoch, latest.GetRevisionHeight(), num, "update_testnet.json")
		},
	}
	cmd.Flags().Uint64Var(&num, "num", 240, "--num")
	return cmd
}

func (m *historyModule) outputMsgUpdate(prover *module.Prover, createdEpoch, latest uint64, num uint64, path string) error {
	type updatingData struct {
		Header string `json:"header"`
	}

	type updates struct {
		Data []updatingData `json:"data"`
	}

	data := updates{}
	var lastFinalized exported.Height = types.NewHeight(0, createdEpoch)
	for i := num; i > 0; i-- {
		log.Println(num - i)
		targetLatest := latest - i
		header, err := prover.GetLatestFinalizedHeaderByLatestHeight(targetLatest)
		if err != nil {
			log.Println(err)
			continue
		}
		blocks, err := prover.SetupHeadersForUpdateByLatestHeight(lastFinalized, header.(*module.Header))
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("set up complete=%d\n", len(blocks))
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
		if len(blocks) > 0 {
			lastFinalized = blocks[0].GetHeight()
		}
		time.Sleep(500 * time.Millisecond)
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, serialized, 0666)
}

func (m *historyModule) outputMsgClient(prover *module.Prover, firstNumber uint64, path string) (uint64, error) {
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
	return firstNumber, os.WriteFile(path, serialized, 0666)
}

func CreateHistoryClient() *cobra.Command {
	cmd := &cobra.Command{
		Use: "history",
	}
	m := historyModule{}
	cmd.AddCommand(m.mainnet())
	cmd.AddCommand(m.testnet())
	return cmd
}
