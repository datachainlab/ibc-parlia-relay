//go:build dev

package constant

import (
	"log"
	"strconv"
)

// can change by ldflags
var epochBlockPeriod string = "200"

var EpochBlockPeriod uint64 = 200

func init() {
	v, err := strconv.Atoi(epochBlockPeriod)
	if err != nil {
		panic(err)
	}
	EpochBlockPeriod = uint64(v)
	log.Printf("blocks per epoch = %d\n", EpochBlockPeriod)
}
