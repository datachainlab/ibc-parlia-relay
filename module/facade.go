package module

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// Facade for tool modules

func GetPreviousEpoch(v uint64) uint64 {
	return getPreviousEpoch(v)
}

func GetCurrentEpoch(v uint64) uint64 {
	return getCurrentEpoch(v)
}

func QueryFinalizedHeader(fn getHeaderFn, height uint64, limitHeight uint64) ([]*ETHHeader, error) {
	return queryFinalizedHeader(context.TODO(), fn, height, limitHeight)
}

func QueryValidatorSetAndTurnLength(fn getHeaderFn, height uint64) (Validators, uint8, error) {
	return queryValidatorSetAndTurnLength(context.TODO(), fn, height)
}

func ExtractValidatorSetAndTurnLength(h *types.Header) (Validators, uint8, error) {
	return extractValidatorSetAndTurnLength(h)
}

func MakeEpochHash(validators Validators, turnLength uint8) []byte {
	return makeEpochHash(validators, turnLength)
}
