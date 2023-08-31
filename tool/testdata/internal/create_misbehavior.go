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
	"github.com/spf13/cobra"
	"log"
)

const (
	hdwMnemonic           = "math razor capable expose worth grape metal sunset metal sudden usage scheme"
	hdwPath               = "m/44'/60'/0'/0/0"
	IbcAddress            = "0x702E40245797c5a2108A566b3CE2Bf14Bc6aF841"
	LocalNetValidatorSize = 3
	MainNetValidatorSize  = 21
	MainNetIbcAddress     = "0x151f3951FA218cac426edFe078fA9e5C6dceA500"
)

func CreateMisbehavior() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "misbehavior",
		Short: "Create testdata for misbehavior. ",
	}
	cmd.AddCommand(misbehaviorSuccessCmd())
	cmd.AddCommand(misbehaviorErrorCmd())
	return cmd
}

func misbehaviorSuccessCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "success",
		Short: "create misbehavior testdata for success",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID := int64(9999)
			targetHeight, header1, err := getLocalHeader(chainID, 8645, 0)
			if err != nil {
				log.Panic(err)
			}
			_, header2, err := getLocalHeader(chainID, 8545, targetHeight)
			if err != nil {
				log.Panic(err)
			}
			clientId := "xx-parlia-1"
			misbehavior := module.Misbehaviour{
				ClientId: clientId,
				Header_1: header1,
				Header_2: header2,
			}

			// print hex for lcp-parlia test
			pack, _ := types.PackClientMessage(&misbehavior)
			marshal, _ := pack.Marshal()
			log.Println("misbehavior", common.Bytes2Hex(marshal))
			log.Println("trustedHeight", header1.TrustedHeight)
			log.Println("targetValidatorHash", common.Bytes2Hex(crypto.Keccak256(header1.TargetValidators...)))

			epochCount := header1.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
			if header1.GetHeight().GetRevisionHeight()%constant.BlocksPerEpoch >= (LocalNetValidatorSize/2 + 1) {
				log.Println("targetValidatorEpoch", epochCount*constant.BlocksPerEpoch)
			} else {
				log.Println("targetValidatorEpoch", (epochCount-1)*constant.BlocksPerEpoch)
			}
			return nil
		},
	}
}

func misbehaviorErrorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "error",
		Short: "create misbehavior testdata for error",
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := ethereum.NewChain(ethereum.ChainConfig{
				EthChainId:  56,
				RpcAddr:     "https://bsc-mainnet-rpc.allthatnode.com",
				HdwMnemonic: hdwMnemonic,
				HdwPath:     hdwPath,
				IbcAddress:  MainNetIbcAddress,
			})
			if err != nil {
				return err
			}

			config := module.ProverConfig{
				Debug: true,
			}
			prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

			latestHeight, err := chain.LatestHeight()
			if err != nil {
				return err
			}
			latest := latestHeight.GetRevisionHeight()
			println(latest)
			header, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
			target, err := header.(*module.Header).DecodedTarget()
			updating, err := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header.(*module.Header))
			if err != nil {
				return err
			}

			// Exactly same block
			misbehavior := module.Misbehaviour{
				ClientId: "xx-parlia-1",
				Header_1: updating[0].(*module.Header),
				Header_2: updating[0].(*module.Header),
			}
			pack, _ := types.PackClientMessage(&misbehavior)
			marshal, _ := pack.Marshal()
			log.Println("Exactly same block: misbehavior", common.Bytes2Hex(marshal))
			log.Println("Exactly same block: height", target.Number.Int64())

			// Invalid block
			header2, _ := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
			updating2, _ := prover.SetupHeadersForUpdateByLatestHeight(types.NewHeight(header2.GetHeight().GetRevisionNumber(), target.Number.Uint64()-1), header2.(*module.Header))
			target2, _ := updating2[0].(*module.Header).DecodedTarget()
			target2.Root = common.Hash{}
			rlpTarget, err := rlp.EncodeToBytes(target2)
			updating2[0].(*module.Header).Target = &module.ETHHeader{Header: rlpTarget}
			misbehavior2 := module.Misbehaviour{
				ClientId: "xx-parlia-1",
				Header_1: updating[0].(*module.Header),
				Header_2: updating2[0].(*module.Header),
			}
			pack, _ = types.PackClientMessage(&misbehavior2)
			marshal, _ = pack.Marshal()
			log.Println("Invalid block: misbehavior", common.Bytes2Hex(marshal))
			log.Println("Invalid block: height", header.GetHeight())
			log.Println("Invalid block: target_validator_hash", common.Bytes2Hex(crypto.Keccak256(header.(*module.Header).TargetValidators...)))
			log.Println("Invalid block: trusted_height", updating[0].(*module.Header).TrustedHeight)
			epochCount := header.GetHeight().GetRevisionHeight() / constant.BlocksPerEpoch
			if header.GetHeight().GetRevisionHeight()%constant.BlocksPerEpoch >= (MainNetValidatorSize/2 + 1) {
				log.Println("Invalid block: targetValidatorEpoch", epochCount*constant.BlocksPerEpoch)
			} else {
				log.Println("Invalid block: targetValidatorEpoch", (epochCount-1)*constant.BlocksPerEpoch)
			}

			return nil
		},
	}
}

func getLocalHeader(chainID int64, port int64, targetHeight uint64) (uint64, *module.Header, error) {
	chain, err := ethereum.NewChain(ethereum.ChainConfig{
		EthChainId:  chainID,
		RpcAddr:     fmt.Sprintf("http://localhost:%d", port),
		HdwMnemonic: hdwMnemonic,
		HdwPath:     hdwPath,
		IbcAddress:  IbcAddress,
	})
	if err != nil {
		return targetHeight, nil, err
	}
	if targetHeight == 0 {
		latest, err := chain.LatestHeight()
		if err != nil {
			return targetHeight, nil, err
		}
		targetHeight = latest.GetRevisionHeight()
	}
	config := module.ProverConfig{
		Debug: true,
	}
	prover := module.NewProver(module.NewChain(chain), &config).(*module.Prover)

	// Get Finalized header
	latestHeight := types.NewHeight(0, targetHeight)
	latest := latestHeight.GetRevisionHeight()
	iHeader, err := prover.GetLatestFinalizedHeaderByLatestHeight(latest)
	if err != nil {
		return latest, nil, err
	}
	header := iHeader.(*module.Header)
	target, err := header.DecodedTarget()
	if err != nil {
		return latest, nil, err
	}
	trustedHeight := types.NewHeight(0, target.Number.Uint64()-5)
	header.TrustedHeight = &trustedHeight
	return latest, header, nil
}
