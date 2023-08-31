package module

import (
	"fmt"
	"log"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
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
	target, err := h.DecodedTarget()
	if err != nil {
		log.Panicf("invalid header: %v", h)
	}
	//TODO revision number
	return clienttypes.NewHeight(0, target.Number.Uint64())
}

func (h *Header) ValidateBasic() error {
	if _, err := h.DecodedTarget(); err != nil {
		return err
	}
	if _, err := h.DecodedParent(); err != nil {
		return err
	}
	if _, err := decodeAccountProof(h.AccountProof); err != nil {
		return err
	}
	return nil
}

func (h *Header) DecodedTarget() (*types.Header, error) {
	var ethHeader types.Header
	if err := rlp.DecodeBytes(h.Target.Header, &ethHeader); err != nil {
		return nil, err
	}
	return &ethHeader, nil
}

func (h *Header) DecodedParent() (*types.Header, error) {
	var ethHeader types.Header
	if err := rlp.DecodeBytes(h.Parent.Header, &ethHeader); err != nil {
		return nil, err
	}
	return &ethHeader, nil
}

func (h *Header) Account(path common.Address) (*types.StateAccount, error) {
	target, err := h.DecodedTarget()
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

	validatorCount := int(validators[0])
	validatorsWithBLS := validators[1 : validatorCount*validatorBytesLength]
	for i := 0; i < validatorCount; i++ {
		start := validatorBytesLength * i
		validatorWithBLS := validatorsWithBLS[start : start+validatorBytesLength]
		validatorSet = append(validatorSet, validatorWithBLS[:validatorBytesLengthBeforeLuban])
	}

	return validatorSet, nil
}
