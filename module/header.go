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
	if _, err := h.DecodedChild(); err != nil {
		return err
	}
	if _, err := h.DecodedGrandChild(); err != nil {
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

func (h *Header) DecodedChild() (*types.Header, error) {
	var ethHeader types.Header
	if err := rlp.DecodeBytes(h.Child.Header, &ethHeader); err != nil {
		return nil, err
	}
	return &ethHeader, nil
}

func (h *Header) DecodedGrandChild() (*types.Header, error) {
	var ethHeader types.Header
	if err := rlp.DecodeBytes(h.GrandChild.Header, &ethHeader); err != nil {
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

func ExtractValidatorSet(h *types.Header) ([][]byte, error) {
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, fmt.Errorf("invalid extra length : %d", h.Number.Uint64())
	}
	num := int(extra[extraVanity])
	if num == 0 || len(extra) <= extraVanity+extraSeal+num*validatorBytesLength {
		return nil, fmt.Errorf("invalid validator bytes length: %d", h.Number.Uint64())
	}
	start := extraVanity + validatorNumberSize
	end := start + num*validatorBytesLength
	validators := extra[start:end]

	var validatorSet [][]byte
	for i := 0; i < num; i++ {
		s := validatorBytesLength * i
		e := s + validatorBytesLength
		validatorWithBLS := validators[s:e]
		validatorSet = append(validatorSet, validatorWithBLS)
	}

	return validatorSet, nil
}
