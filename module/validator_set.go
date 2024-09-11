package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Validators [][]byte

func (v Validators) Checkpoint(turnLength uint8) uint64 {
	return uint64(len(v)/2+1) * uint64(turnLength)
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
	required := v.threshold()
	return count >= required
}

func (v Validators) threshold() int {
	return len(v) - ceilDiv(len(v)*2, 3) + 1
}

func ceilDiv(x, y int) int {
	if y == 0 {
		return 0
	}
	return (x + y - 1) / y
}

func queryValidatorSetAndTurnLength(fn getHeaderFn, epochBlockNumber uint64) (Validators, uint8, error) {
	header, err := fn(context.TODO(), epochBlockNumber)
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
