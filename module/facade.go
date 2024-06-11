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

func QueryValidatorSetAndTurnTerm(fn getHeaderFn, height uint64) (Validators, uint8, error) {
	return queryValidatorSetAndTurnTerm(fn, height)
}

func ExtractValidatorSetAndTurnTerm(h *types.Header) (Validators, uint8, error) {
	return extractValidatorSetAndTurnTerm(h)
}

func MakeEpochHash(validators Validators, turnTerm uint8) []byte {
	return makeEpochHash(validators, turnTerm)
}
