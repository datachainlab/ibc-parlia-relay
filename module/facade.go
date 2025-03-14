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

func QueryFinalizedHeader(ctx context.Context, fn getHeaderFn, height uint64, limitHeight uint64) ([]*ETHHeader, error) {
	return queryFinalizedHeader(ctx, fn, height, limitHeight)
}

func QueryValidatorSetAndTurnLength(ctx context.Context, fn getHeaderFn, height uint64) (Validators, uint8, error) {
	return queryValidatorSetAndTurnLength(ctx, fn, height)
}

func ExtractValidatorSetAndTurnLength(h *types.Header) (Validators, uint8, error) {
	return extractValidatorSetAndTurnLength(h)
}

func MakeEpochHash(validators Validators, turnLength uint8) []byte {
	return makeEpochHash(validators, turnLength)
}
