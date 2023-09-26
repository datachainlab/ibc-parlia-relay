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

	var num uint64
	specified := &cobra.Command{
		Use:   "specified",
		Short: "for specified block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, _, err := createProver()
			if err != nil {
				return err
			}
			return m.printHeader(prover, num)
		},
	}
	specified.Flags().Uint64Var(&num, "num", num, "--num")
	cmd.AddCommand(specified)
	return cmd
}

func (m *headerModule) printHeader(prover *module.Prover, height uint64) error {
	log.Println("printHeader latest=", height)
	headers, err := prover.QueryVerifyingEthHeaders(height)
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
		Use:   "header",
		Short: "Create testdata for update client. ",
	}
	m := headerModule{}
	cmd.AddCommand(m.success())
	return cmd
}
