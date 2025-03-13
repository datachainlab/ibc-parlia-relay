package module

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

type Validators [][]byte

func (v Validators) Checkpoint(turnLength uint8) uint64 {
	return uint64(len(v)/2+1) * uint64(turnLength)
}

func queryValidatorSetAndTurnLength(ctx context.Context, fn getHeaderFn, epochBlockNumber uint64) (Validators, uint8, error) {
	header, err := fn(ctx, epochBlockNumber)
	if err != nil {
		return nil, 1, err
	}
	return extractValidatorSetAndTurnLength(header)
}

func extractValidatorSetAndTurnLength(h *types.Header) (Validators, uint8, error) {
	const turnLengthLength = 1
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, 1, fmt.Errorf("invalid extra length : %d", h.Number.Uint64())
	}
	num := int(extra[extraVanity])
	if num == 0 || len(extra) <= extraVanity+extraSeal+num*validatorBytesLength+turnLengthLength {
		return nil, 1, fmt.Errorf("invalid validator bytes length: %d", h.Number.Uint64())
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
	turnLength := extra[end]
	return validatorSet, turnLength, nil
}
