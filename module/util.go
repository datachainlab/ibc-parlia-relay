package module

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
)

func getPreviousEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	if epochCount == 0 {
		return 0
	}
	return (epochCount - 1) * constant.BlocksPerEpoch
}

func isEpoch(v uint64) bool {
	return v%constant.BlocksPerEpoch == 0
}

func getCurrentEpoch(v uint64) uint64 {
	return toEpoch(v)
}

func toHeight(height exported.Height) clienttypes.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
}

func toEpoch(v uint64) uint64 {
	epochCount := v / constant.BlocksPerEpoch
	return epochCount * constant.BlocksPerEpoch
}

func minUint64(x uint64, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}
