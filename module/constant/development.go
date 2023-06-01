//go:build dev

package constant

import (
	"log"
	"strconv"
)

// can change by ldflags
var blocksPerEpoch string = "200"

var BlocksPerEpoch uint64 = 200

func init() {
	iBlocksPerEpoch, err := strconv.Atoi(blocksPerEpoch)
	if err != nil {
		panic(err)
	}
	BlocksPerEpoch = uint64(iBlocksPerEpoch)
	log.Printf("blocks per epoch = %d", BlocksPerEpoch)
}
