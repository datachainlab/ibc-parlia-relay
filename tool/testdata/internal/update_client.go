package internal

import (
	"context"
	"log"

	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			latest, err := chain.LatestHeight(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			return m.printHeader(cmd.Context(), prover, chain, latest.GetRevisionHeight())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "epoch",
		Short: "for epoch block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight(cmd.Context())
			if err != nil {
				return err
			}
			be, err := m.GetBoundaryEpoch(chain, latest.GetRevisionHeight())
			if err != nil {
				return err
			}
			epoch := be.CurrentEpochBlockNumber(latest.GetRevisionHeight())
			return m.printHeader(cmd.Context(), prover, chain, epoch+2)
		},
	})
	var num uint64
	var diff int64
	specified := &cobra.Command{
		Use: "specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			be, err := m.GetBoundaryEpoch(chain, num)
			if err != nil {
				return err
			}
			currentEpoch := be.CurrentEpochBlockNumber(num)
			previousEpoch := be.PreviousEpochBlockNumber(currentEpoch)
			validator, turnLength, err := module.QueryValidatorSetAndTurnLength(cmd.Context(), chain.Header, previousEpoch)
			if err != nil {
				return errors.WithStack(err)
			}
			checkpoint := currentEpoch + validator.Checkpoint(turnLength)
			target, err := prover.GetLatestFinalizedHeaderByLatestHeight(cmd.Context(), uint64(int64(num)+2+diff))
			if err != nil {
				return errors.WithStack(err)
			}
			log.Println("checkpoint", checkpoint, "turnLength", turnLength, "target", target.GetHeight())
			headers, err := prover.SetupHeadersForUpdateByLatestHeight(cmd.Context(), types.NewHeight(0, previousEpoch), target.(*module.Header))
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
			prover, chain, err := createProver(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			latest, err := chain.LatestHeight(cmd.Context())
			if err != nil {
				return errors.WithStack(err)
			}
			be, err := m.GetBoundaryEpoch(chain, latest.GetRevisionHeight())
			if err != nil {
				return errors.WithStack(err)
			}
			epoch := be.CurrentEpochBlockNumber(latest.GetRevisionHeight())
			prevEpoch := be.PreviousEpochBlockNumber(epoch)
			header, err := prover.GetLatestFinalizedHeaderByLatestHeight(cmd.Context(), epoch+2)
			if err != nil {
				return errors.WithStack(err)
			}
			updating, err := prover.SetupHeadersForUpdateByLatestHeight(cmd.Context(), types.NewHeight(0, prevEpoch), header.(*module.Header))
			if err != nil {
				return errors.WithStack(err)
			}

			// non neighboring epoch
			prevPrevEpoch := be.PreviousEpochBlockNumber(prevEpoch)
			newTrustedHeight := types.NewHeight(0, prevPrevEpoch)
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

func (m *updateClientModule) printHeader(ctx context.Context, prover *module.Prover, chain module.Chain, height uint64) error {
	log.Println("printHeader latest=", height)
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(ctx, height)
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

	// setup
	updating, err := prover.SetupHeadersForUpdateByLatestHeight(ctx, types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header)
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
	trustedBe, err := m.GetBoundaryEpoch(chain, trustedHeight)
	if err != nil {
		return err
	}
	currentEpoch := trustedBe.CurrentEpochBlockNumber(trustedHeight)
	currentValidatorSetOfTrustedHeight, currentTurnLengthOfTrustedHeight, err := module.QueryValidatorSetAndTurnLength(ctx, chain.Header, currentEpoch)
	if err != nil {
		return err
	}
	previousEpoch := trustedBe.PreviousEpochBlockNumber(currentEpoch)
	previousValidatorSetOfTrustedHeight, previousTurnLengthOfTrustedHeight, err := module.QueryValidatorSetAndTurnLength(ctx, chain.Header, previousEpoch)
	if err != nil {
		return err
	}
	log.Println("header", common.Bytes2Hex(marshal))
	log.Println("height", header.GetHeight().GetRevisionHeight())
	log.Println("trustedHeight", trustedHeight)
	log.Println("currentEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(currentValidatorSetOfTrustedHeight, currentTurnLengthOfTrustedHeight)))
	log.Println("previousEpochHashOfTrustedHeight", common.Bytes2Hex(module.MakeEpochHash(previousValidatorSetOfTrustedHeight, previousTurnLengthOfTrustedHeight)))
	newValidators, newTurnLength, err := module.ExtractValidatorSetAndTurnLength(target)
	if err != nil {
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.CurrentValidators, uint8(header.CurrentTurnLength))))
	} else {
		log.Println("newCurrentEpochHash", common.Bytes2Hex(module.MakeEpochHash(newValidators, newTurnLength)))
	}
	log.Println("newPreviousEpochHash", common.Bytes2Hex(module.MakeEpochHash(header.PreviousValidators, uint8(header.PreviousTurnLength))))
	return nil
}

func (m *updateClientModule) GetBoundaryEpoch(chain module.Chain, height uint64) (*module.BoundaryEpochs, error) {
	header, err := chain.Header(context.Background(), height)
	if err != nil {
		return nil, err
	}
	forkSpec, prev, err := module.FindTargetForkSpec(module.GetForkParameters(module.Localnet), header.Number.Uint64(), module.MilliTimestamp(header))
	if err != nil {
		return nil, err
	}
	bh, err := module.GetBoundaryHeight(context.Background(), chain.Header, header.Number.Uint64(), *forkSpec)
	if err != nil {
		return nil, err
	}
	return bh.GetBoundaryEpochs(prev)
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
