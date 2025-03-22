package module

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

func toHeight(height exported.Height) clienttypes.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight())
}

func minUint64(x uint64, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}
