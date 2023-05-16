//go:build dev

package constant

import (
	"log"
	"strconv"
)

// can change by ldflags
var blocksPerEpoch string = "200"
var lubanFork string = "29295050"

var BlocksPerEpoch uint64 = 200
var LubanFork uint64 = 29295050

func init() {
	iBlocksPerEpoch, err := strconv.Atoi(blocksPerEpoch)
	if err != nil {
		panic(err)
	}
	iLubanFork, err := strconv.Atoi(lubanFork)
	if err != nil {
		panic(err)
	}
	BlocksPerEpoch = uint64(iBlocksPerEpoch)
	LubanFork = uint64(iLubanFork)
	log.Printf("blocks per epoch = %d, lubanFork=%d\n", BlocksPerEpoch, LubanFork)
}
