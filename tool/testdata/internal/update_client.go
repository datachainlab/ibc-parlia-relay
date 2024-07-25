package internal

import (
	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
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
	var num uint64
	var diff int64
	specified := &cobra.Command{
		Use: "specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return errors.WithStack(err)
			}
			currentEpoch := module.GetCurrentEpoch(num)
			previousEpoch := module.GetPreviousEpoch(num)
			validator, turnLength, err := module.QueryValidatorSetAndTurnLength(chain.Header, previousEpoch)
			if err != nil {
				return errors.WithStack(err)
			}
			checkpoint := currentEpoch + validator.Checkpoint(turnLength)
			target, err := prover.GetLatestFinalizedHeaderByLatestHeight(uint64(int64(num) + 2 + diff))
			if err != nil {
				return errors.WithStack(err)
			}
			log.Println("checkpoint", checkpoint, "turnLength", turnLength, "target", target.GetHeight())
			headers, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(0, previousEpoch), target.(*module.Header))
			if err != nil {
				return errors.WithStack(err)
			}
			for _, header := range headers {
				pack, err := types.PackClientMessage(header)
				if err != nil {
					return errors.WithStack(err)
				}
				marshal, err := pack.Marshal()
				if err != nil {
					return err
				}
				log.Println(common.Bytes2Hex(marshal))
			}
			return nil
		},
	}
	specified.Flags().Uint64Var(&num, "num", num, "--num")
	specified.Flags().Int64Var(&diff, "diff", diff, "--diff")
	cmd.AddCommand(specified)
	return cmd
}

func (m *updateClientModule) error() *cobra.Command {
	return &cobra.Command{
		Use:   "error",
		Short: "create updateClient testdata for error",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver()
			if err != nil {
				return errors.WithStack(err)
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return errors.WithStack(err)
			}
			epoch := module.GetCurrentEpoch(latest.GetRevisionHeight())
			header, err := prover.GetLatestFinalizedHeaderByLatestHeight(epoch + 2)
			if err != nil {
				return errors.WithStack(err)
			}
			updating, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(0, header.GetHeight().GetRevisionNumber()-constant.BlocksPerEpoch), header.(*module.Header))
			if err != nil {
				return errors.WithStack(err)
			}

			// non neighboring epoch
			newTrustedHeight := types.NewHeight(0, header.GetHeight().GetRevisionHeight()-2*constant.BlocksPerEpoch)
			updating[0].(*module.Header).TrustedHeight = &newTrustedHeight
			pack, err := types.PackClientMessage(updating[0])
			if err != nil {
				return errors.WithStack(err)
			}
			marshal, err := pack.Marshal()
			if err != nil {
				return errors.WithStack(err)
			}
			log.Println("header", common.Bytes2Hex(marshal))
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
	currentValidatorSetOfTrustedHeight, currentTurnLengthOfTrustedHeight, err := module.QueryValidatorSetAndTurnLength(chain.Header, module.GetCurrentEpoch(trustedHeight))
	if err != nil {
		return err
	}
	previousValidatorSetOfTrustedHeight, previousTurnLengthOfTrustedHeight, err := module.QueryValidatorSetAndTurnLength(chain.Header, module.GetPreviousEpoch(trustedHeight))
	if err != nil {
		return err
	}
	log.Println("header", common.Bytes2Hex(marshal))
	log.Println("stateRoot", account.Root)
	log.Println("height", header.GetHeight().GetRevisionHeight())
	log.Println("trustedHeight", trustedHeight)
	log.Println("currentEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(currentValidatorSetOfTrustedHeight, currentTurnLengthOfTrustedHeight)))
	log.Println("previousEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(previousValidatorSetOfTrustedHeight, previousTurnLengthOfTrustedHeight)))
	if target.Number.Uint64()%constant.BlocksPerEpoch == 0 {
		newValidators, newTurnLength, err := module.ExtractValidatorSetAndTurnLength(target)
		if err != nil {
			return err
		}
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(newValidators, newTurnLength)))
	} else {
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.CurrentValidators, uint8(header.CurrentTurnLength))))
	}
	log.Println("newPreviousEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.PreviousValidators, uint8(header.PreviousTurnLength))))
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
