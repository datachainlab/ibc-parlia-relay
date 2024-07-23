package internal

import (
	"context"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
)

type headerModule struct {
}

func (m *headerModule) success() *cobra.Command {
	cmd := &cobra.Command{
		Use: "success",
	}
	cmd.AddCommand(&cobra.Command{
		Use: "latest",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, chain, err := createProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			return m.printHeader(chain, latest.GetRevisionHeight())
		},
	})

	var num uint64
	specified := &cobra.Command{
		Use: "specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, chain, err := createProver()
			if err != nil {
				return err
			}
			return m.printHeader(chain, num)
		},
	}
	specified.Flags().Uint64Var(&num, "num", num, "--num")
	cmd.AddCommand(specified)
	return cmd
}

func (m *headerModule) printHeader(chain module.Chain, height uint64) error {
	log.Println("printHeader height=", height)
	header, err := chain.Header(context.Background(), height)
	if err != nil {
		return errors.WithStack(err)
	}
	if height%constant.BlocksPerEpoch == 0 {
		vals, turnLength, err := module.ExtractValidatorSetAndTurnLength(header)
		if err != nil {
			return errors.WithStack(err)
		}
		log.Println("validators = ")
		for _, val := range vals {
			log.Println(common.Bytes2Hex(val))
		}
		log.Println("turnLength = ", turnLength)
	}

	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Println("ETHHeader.header=", common.Bytes2Hex(rlpHeader))
	return nil
}

func CreateHeader() *cobra.Command {
	cmd := &cobra.Command{
		Use: "header",
	}
	m := headerModule{}
	cmd.AddCommand(m.success())
	return cmd
}
