package module

import "github.com/ethereum/go-ethereum/core/types"

// Facade for tool modules

func GetPreviousEpoch(v uint64) uint64 {
	return getPreviousEpoch(v)
}

func GetCurrentEpoch(v uint64) uint64 {
	return getCurrentEpoch(v)
}

func QueryFinalizedHeader(fn getHeaderFn, height uint64, limitHeight uint64) ([]*ETHHeader, error) {
	return queryFinalizedHeader(fn, height, limitHeight)
}

func QueryValidatorSet(fn getHeaderFn, height uint64) (Validators, error) {
	return queryValidatorSet(fn, height)
}

func ExtractValidatorSet(h *types.Header) (Validators, error) {
	return extractValidatorSet(h)
}
