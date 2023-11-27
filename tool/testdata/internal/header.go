package internal

import (
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/ethereum/go-ethereum/common"
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
	log.Println("printHeader latest=", height)
	headers, err := module.GetFinalizedHeader(chain.Header, height, height+10)
	if err != nil {
		return err
	}
	header := module.Header{
		Headers: headers,
		TrustedHeight: &types.Height{
			RevisionNumber: 0,
			RevisionHeight: 0,
		},
		PreviousValidators: [][]byte{common.Address{}.Bytes()},
		CurrentValidators:  [][]byte{common.Address{}.Bytes()},
	}

	pack, err := types.PackClientMessage(&header)
	if err != nil {
		return err
	}
	marshal, err := pack.Marshal()

	log.Println("header", common.Bytes2Hex(marshal))
	log.Println(header.GetHeight(), len(header.Headers))
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
