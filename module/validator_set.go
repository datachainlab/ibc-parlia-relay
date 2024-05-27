package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Validators [][]byte

func (v Validators) Checkpoint(turnTerm uint8) uint64 {
	return uint64(len(v)/2*int(turnTerm) + 1)
}

func (v Validators) Contains(other Validators) bool {
	count := 0
	for _, x := range other {
		for _, y := range v {
			if common.Bytes2Hex(x) == common.Bytes2Hex(y) {
				count++
				break
			}
		}
	}
	required := ceilDiv(len(v), 3)
	return count >= required
}

func ceilDiv(x, y int) int {
	if y == 0 {
		return 0
	}
	return (x + y - 1) / y
}

func queryValidatorSetAndTurnTerm(fn getHeaderFn, epochBlockNumber uint64) (Validators, uint8, error) {
	header, err := fn(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, 1, err
	}
	return extractValidatorSetAndTurnTerm(header)
}

func extractValidatorSetAndTurnTerm(h *types.Header) (Validators, uint8, error) {
	const turnTermLength = 1
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, 1, fmt.Errorf("invalid extra length : %d", h.Number.Uint64())
	}
	num := int(extra[extraVanity])
	if num == 0 || len(extra) <= extraVanity+extraSeal+num*validatorBytesLength+turnTermLength {
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
	turnTerm := extra[end]
	return validatorSet, turnTerm, nil
}
