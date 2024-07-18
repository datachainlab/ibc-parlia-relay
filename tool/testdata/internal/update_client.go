package internal

import (
	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"os"
)

type updateClientModule struct {
}

func (m *updateClientModule) success() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "success",
		Short: "create updateClient testdata for success",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "latest",
		Short: "for latest block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return errors.WithStack(err)
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return errors.WithStack(err)
			}
			return m.printHeader(prover, chain, latest.GetRevisionHeight())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "epoch",
		Short: "for epoch block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			epochCount := latest.GetRevisionHeight() / constant.BlocksPerEpoch
			return m.printHeader(prover, chain, epochCount*constant.BlocksPerEpoch+2)
		},
	})
	return cmd
}

func (m *updateClientModule) error() *cobra.Command {
	return &cobra.Command{
		Use:   "error",
		Short: "create updateClient testdata for error",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			header, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest.GetRevisionHeight())
			if err != nil {
				return err
			}
			target, err := header.(*module.Header).Target()
			if err != nil {
				return err
			}
			updating, _ := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header.(*module.Header))
			target.Root = common.Hash{}
			rlpTarget, err := rlp.EncodeToBytes(target)
			updating[0].(*module.Header).Headers[0] = &module.ETHHeader{Header: rlpTarget}
			pack, err := types.PackClientMessage(updating[0])
			if err != nil {
				return err
			}
			marshal, err := pack.Marshal()
			if err != nil {
				return err
			}
			trustedHeight := header.(*module.Header).TrustedHeight.GetRevisionHeight()
			currentValidatorSetOfTrustedHeight, currentTurnTermOfTrustedHeight, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetCurrentEpoch(trustedHeight))
			if err != nil {
				return err
			}
			previousValidatorSetOfTrustedHeight, previousTurnTermOfTrustedHeight, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetPreviousEpoch(trustedHeight))
			if err != nil {
				return err
			}
			log.Println("header", common.Bytes2Hex(marshal))
			log.Println("height", header.GetHeight().GetRevisionHeight())
			log.Println("trustedHeight", trustedHeight)
			log.Println("currentEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(currentValidatorSetOfTrustedHeight, currentTurnTermOfTrustedHeight)))
			log.Println("previousEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(previousValidatorSetOfTrustedHeight, previousTurnTermOfTrustedHeight)))
			log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.(*module.Header).CurrentValidators, uint8(header.(*module.Header).CurrentTurnTerm))))
			log.Println("newPreviousEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.(*module.Header).PreviousValidators, uint8(header.(*module.Header).PreviousTurnTerm))))
			return nil
		},
	}
}

func (m *updateClientModule) printHeader(prover *module.Prover, chain module.Chain, height uint64) error {
	log.Println("printHeader latest=", height)
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(height)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = iHeader.ValidateBasic(); err != nil {
		return errors.WithStack(err)
	}
	header := iHeader.(*module.Header)
	target, err := header.Target()
	if err != nil {
		return errors.WithStack(err)
	}

	account, err := header.Account(common.HexToAddress(os.Getenv("BSC_IBC_ADDR")))
	if err != nil {
		return errors.WithStack(err)
	}

	// setup
	updating, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header)
	if err != nil {
		return err
	}

	// updating msg
	pack, err := types.PackClientMessage(updating[0])
	if err != nil {
		return err
	}
	marshal, err := pack.Marshal()
	if err != nil {
		return err
	}

	trustedHeight := updating[0].(*module.Header).TrustedHeight.GetRevisionHeight()
	currentValidatorSetOfTrustedHeight, currentTurnTermOfTrustedHeight, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetCurrentEpoch(trustedHeight))
	if err != nil {
		return err
	}
	previousValidatorSetOfTrustedHeight, previousTurnTermOfTrustedHeight, err := module.QueryValidatorSetAndTurnTerm(chain.Header, module.GetPreviousEpoch(trustedHeight))
	if err != nil {
		return err
	}
	log.Println("header", common.Bytes2Hex(marshal))
	log.Println("stateRoot", account.Root)
	log.Println("height", header.GetHeight().GetRevisionHeight())
	log.Println("trustedHeight", trustedHeight)
	log.Println("currentEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(currentValidatorSetOfTrustedHeight, currentTurnTermOfTrustedHeight)))
	log.Println("previousEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(previousValidatorSetOfTrustedHeight, previousTurnTermOfTrustedHeight)))
	if target.Number.Uint64()%constant.BlocksPerEpoch == 0 {
		newValidators, newTurnTerm, err := module.ExtractValidatorSetAndTurnTerm(target)
		if err != nil {
			return err
		}
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(newValidators, newTurnTerm)))
	} else {
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.CurrentValidators, uint8(header.CurrentTurnTerm))))
	}
	log.Println("newPreviousEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.PreviousValidators, uint8(header.PreviousTurnTerm))))
	return nil
}

func CreateUpdateClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Create testdata for update client. ",
	}
	m := updateClientModule{}
	cmd.AddCommand(m.success())
	cmd.AddCommand(m.error())
	return cmd
}
