package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Validators [][]byte

func (v Validators) CheckpointValue() uint64 {
	return uint64(len(v)/2 + 1)
}

func (v Validators) Checkpoint(epoch uint64) uint64 {
	return epoch + v.CheckpointValue()
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

func QueryValidatorSet(fn getHeaderFn, epochBlockNumber uint64) (Validators, error) {
	header, err := fn(context.TODO(), epochBlockNumber)
	if err != nil {
		return nil, err
	}
	return ExtractValidatorSet(header)
}

func ExtractValidatorSet(h *types.Header) (Validators, error) {
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
