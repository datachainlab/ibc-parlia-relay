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

func uniqMap[T any, R comparable](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, 0, len(collection))
	seen := make(map[R]struct{}, len(collection))

	for i := range collection {
		r := iteratee(collection[i], i)
		if _, ok := seen[r]; !ok {
			result = append(result, r)
			seen[r] = struct{}{}
		}
	}
	return result
}
