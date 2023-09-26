package internal

import (
	"fmt"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
	"log"
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
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			return m.printHeader(prover, latest.GetRevisionHeight())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "31894084",
		Short: "for 31894084 block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, _, err := createProver()
			if err != nil {
				return err
			}
			return m.printHeader(prover, 31894084)
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
			return m.printHeader(prover, epochCount*constant.BlocksPerEpoch+10)
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
			log.Println("header", common.Bytes2Hex(marshal))
			log.Println("height", header.GetHeight().GetRevisionHeight())
			log.Println("trustedHeight", header.(*module.Header).TrustedHeight.GetRevisionHeight())
			epochCount := header.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
			log.Println("currentEpochHeight", epochCount*constant.BlocksPerEpoch)
			log.Println("currentValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.(*module.Header).CurrentValidators...)))
			log.Println("previousValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.(*module.Header).PreviousValidators...)))
			return nil
		},
	}
}

func (m *updateClientModule) printHeader(prover *module.Prover, height uint64) error {
	log.Println("printHeader latest=", height)
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(height)
	if err != nil {
		return err
	}
	if err = iHeader.ValidateBasic(); err != nil {
		return err
	}
	header := iHeader.(*module.Header)
	target, err := header.Target()
	if err != nil {
		return err
	}

	account, err := header.Account(common.HexToAddress(mainAndTestNetIbcAddress))
	if err != nil {
		return err
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
	log.Println("header", common.Bytes2Hex(marshal))
	log.Println("stateRoot", account.Root)
	log.Println("height", header.GetHeight().GetRevisionHeight())
	log.Println("trustedHeight", header.TrustedHeight.GetRevisionHeight())
	epochCount := header.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
	log.Println("currentEpochHeight", epochCount*constant.BlocksPerEpoch)

	// validators hash
	log.Println("currentValidatorSize", len(header.CurrentValidators))
	log.Println("currentValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.CurrentValidators...)))
	log.Println("previousValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.PreviousValidators...)))
	if target.Number.Uint64()%constant.BlocksPerEpoch == 0 {
		newValidators, err := module.ExtractValidatorSet(target)
		if err != nil {
			return err
		}
		if len(newValidators) != mainNetValidatorSize {
			return fmt.Errorf("invalid validator size for test")
		}
		log.Println("newValidatorHash", common.Bytes2Hex(crypto.Keccak256(newValidators...)))
	}
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
