package module

import (
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

const extraVanity = 32
const extraSeal = 65
const validatorBytesLengthBeforeLuban = 20
const validatorBytesLength = 68

// Parlia TODO client_type
const Parlia string = "xx-parlia"

func (*Header) ClientType() string {
	return Parlia
}

func (h *Header) GetHeight() exported.Height {
	target, err := h.Target()
	if err != nil {
		log.Panicf("invalid header: %v", h)
	}
	//TODO revision number
	return clienttypes.NewHeight(0, target.Number.Uint64())
}

func (h *Header) ValidateBasic() error {
	if _, err := h.Target(); err != nil {
		return err
	}
	if _, err := decodeAccountProof(h.GetAccountProof()); err != nil {
		return err
	}
	return nil
}

func (h *Header) decodeEthHeaders() ([]*types.Header, error) {
	ethHeaders := make([]*types.Header, len(h.Headers))
	for i, e := range h.Headers {
		var ethHeader types.Header
		if err := rlp.DecodeBytes(e.Header, &ethHeader); err != nil {
			return nil, err
		}
		ethHeaders[i] = &ethHeader
	}
	return ethHeaders, nil
}

func (h *Header) Target() (*types.Header, error) {
	decodedHeaders, err := h.decodeEthHeaders()
	if err != nil {
		return nil, err
	}
	if len(decodedHeaders) == 0 {
		return nil, fmt.Errorf("invalid header length")
	}
	return decodedHeaders[0], nil
}

func (h *Header) Account(path common.Address) (*types.StateAccount, error) {
	target, err := h.Target()
	if err != nil {
		return nil, err
	}
	return verifyAccount(target, h.AccountProof, path)
}

func extractValidatorSet(h *types.Header) ([][]byte, error) {
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, fmt.Errorf("invalid extra length : %d", h.Number.Uint64())
	}
	var validatorSet [][]byte
	validators := extra[extraVanity : len(extra)-extraSeal]
	if h.Number.Uint64() >= constant.LubanFork {
		validatorCount := int(validators[0])
		validatorsWithBLS := validators[1 : validatorCount*validatorBytesLength]
		for i := 0; i < validatorCount; i++ {
			start := validatorBytesLength * i
			validatorWithBLS := validatorsWithBLS[start : start+validatorBytesLength]
			validatorSet = append(validatorSet, validatorWithBLS[:validatorBytesLengthBeforeLuban])
		}
	} else {
		validatorCount := len(validators) / validatorBytesLengthBeforeLuban
		for i := 0; i < validatorCount; i++ {
			start := validatorBytesLengthBeforeLuban * i
			validatorSet = append(validatorSet, validators[start:start+validatorBytesLengthBeforeLuban])
		}
	}
	return validatorSet, nil
}
