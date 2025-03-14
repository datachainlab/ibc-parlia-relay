package internal

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cometbft/cometbft/libs/json"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

type historyModule struct {
}

func (m *historyModule) mainnet() *cobra.Command {
	var num uint64
	cmd := &cobra.Command{
		Use: "mainnet",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Printf("num = %d\n", num)
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight(cmd.Context())
			if err != nil {
				return err
			}

			createdEpoch, err := m.outputMsgClient(cmd.Context(), prover, latest.GetRevisionHeight()-num, "create_mainnet.json")
			if err != nil {
				return err
			}

			return m.outputMsgUpdate(cmd.Context(), prover, createdEpoch, latest.GetRevisionHeight(), num, "update_mainnet.json")
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
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight(cmd.Context())
			if err != nil {
				return err
			}

			createdEpoch, err := m.outputMsgClient(cmd.Context(), prover, latest.GetRevisionHeight()-num, "create_testnet.json")
			if err != nil {
				return err
			}

			return m.outputMsgUpdate(cmd.Context(), prover, createdEpoch, latest.GetRevisionHeight(), num, "update_testnet.json")
		},
	}
	cmd.Flags().Uint64Var(&num, "num", 240, "--num")
	return cmd
}

func (m *historyModule) outputMsgUpdate(ctx context.Context, prover *module.Prover, createdEpoch, latest, num uint64, path string) error {
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
		header, err := prover.GetLatestFinalizedHeaderByLatestHeight(ctx, targetLatest)
		if err != nil {
			log.Println(err)
			continue
		}
		blocks, err := prover.SetupHeadersForUpdateByLatestHeight(ctx, lastFinalized, header.(*module.Header))
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

func (m *historyModule) outputMsgClient(ctx context.Context, prover *module.Prover, firstNumber uint64, path string) (uint64, error) {
	firstHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(ctx, firstNumber)
	if err != nil {
		return 0, err
	}
	type createData struct {
		ClientState    string `json:"clientState"`
		ConsensusState string `json:"consensusState"`
	}
	cs, consState, err := prover.CreateInitialLightClientState(ctx, types.NewHeight(0, firstNumber))
	if err != nil {
		return 0, err
	}

	protoClientState, err := codectypes.NewAnyWithValue(cs.(proto.Message))
	if err != nil {
		return 0, err
	}
	protoConsState, err := codectypes.NewAnyWithValue(consState.(proto.Message))
	if err != nil {
		return 0, err
	}

	anyClientState, err := protoClientState.Marshal()
	if err != nil {
		return 0, err
	}
	anyConsState, err := protoConsState.Marshal()
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
	return firstHeader.GetHeight().GetRevisionHeight(), os.WriteFile(path, serialized, 0666)
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
