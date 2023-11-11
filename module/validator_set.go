package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
)

func QueryValidatorSet(fn getHeaderFn, epochBlockNumber uint64) ([][]byte, error) {
	header, err := fn(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, err
	}
	return ExtractValidatorSet(header)
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
