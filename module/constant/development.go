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
	v, err := strconv.Atoi(blocksPerEpoch)
	if err != nil {
		panic(err)
	}
	BlocksPerEpoch = uint64(v)
	log.Printf("blocks per epoch = %d\n", BlocksPerEpoch)
}
