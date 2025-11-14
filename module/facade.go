package module

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// Facade for tool modules
func QueryValidatorSetAndTurnLength(ctx context.Context, fn getHeaderFn, height uint64) (Validators, uint8, error) {
	return queryValidatorSetAndTurnLength(ctx, fn, height)
}

func ExtractValidatorSetAndTurnLength(h *types.Header) (Validators, uint8, error) {
	return extractValidatorSetAndTurnLength(h)
}

func MakeEpochHash(validators Validators, turnLength uint8) []byte {
	return makeEpochHash(validators, turnLength)
}
