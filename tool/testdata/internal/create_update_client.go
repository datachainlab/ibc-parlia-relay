package internal

import (
	"fmt"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ibc-parlia-relay/module"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

func CreateUpdateClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Create testdata for update client. ",
	}
	cmd.AddCommand(updateClientSuccessCmd())
	cmd.AddCommand(updateClientErrorCmd())
	return cmd
}

func updateClientSuccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "success",
		Short: "create updateClient testdata for success",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "latest",
		Short: "for latest block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createMainnetProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			return printMainnetHeader(prover, latest.GetRevisionHeight())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "epoch",
		Short: "for epoch block",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createMainnetProver()
			if err != nil {
				return err
			}
			latest, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			epochCount := latest.GetRevisionHeight() / constant.BlocksPerEpoch
			return printMainnetHeader(prover, epochCount*constant.BlocksPerEpoch+3)
		},
	})
	return cmd
}

func updateClientErrorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "error",
		Short: "create updateClient testdata for error",
		RunE: func(cmd *cobra.Command, args []string) error {
			prover, chain, err := createMainnetProver()
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
			target, err := header.(*module.Header).DecodedTarget()
			if err != nil {
				return err
			}
			updating, _ := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header.(*module.Header))
			target.Root = common.Hash{}
			rlpTarget, err := rlp.EncodeToBytes(target)
			updating[0].(*module.Header).Target = &module.ETHHeader{Header: rlpTarget}
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
			log.Println("targetValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.(*module.Header).TargetValidators...)))
			log.Println("previousTargetValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.(*module.Header).PreviousTargetValidators...)))
			return nil
		},
	}
}
func printMainnetHeader(prover *module.Prover, height uint64) error {
	log.Println("printMainnetHeader latest=", height)
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(height)
	if err != nil {
		return err
	}
	if err = iHeader.ValidateBasic(); err != nil {
		return err
	}
	header := iHeader.(*module.Header)
	target, err := header.DecodedTarget()
	if err != nil {
		return err
	}

	account, err := header.Account(common.HexToAddress(MainNetIbcAddress))
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
	log.Println("targetValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.TargetValidators...)))
	log.Println("previousTargetValidatorHash", common.Bytes2Hex(crypto.Keccak256(header.PreviousTargetValidators...)))
	if target.Number.Uint64()%constant.BlocksPerEpoch == 0 {
		newValidators, err := module.ExtractValidatorSet(target)
		if err != nil {
			return err
		}
		if len(newValidators) != MainNetValidatorSize {
			return fmt.Errorf("invalid validator size for test")
		}
		log.Println("newValidatorHash", common.Bytes2Hex(crypto.Keccak256(newValidators...)))
	}
	return nil
}

func createRPCAddr() (string, error) {
	rpcAddr, ok := viper.Get("BSC_MAINNET_RPC_ADDR").(string)
	if !ok {
		return "", fmt.Errorf("BSC_MAINNET_RPC_ADDR is required")
	}
	return rpcAddr, nil
}

func createMainnetProver() (*module.Prover, core.Chain, error) {
	rpcAddr, err := createRPCAddr()
	if err != nil {
		return nil, nil, err
	}
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  56,
		RpcAddr:     rpcAddr,
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  MainNetIbcAddress,
	})
	if err != nil {
		return nil, chain, err
	}

	config := module.ProverConfig{
		Debug: true,
	}
	ec := module.NewChain(chain)
	return module.NewProver(ec, &config).(*module.Prover), chain, nil
}
