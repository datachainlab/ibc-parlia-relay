package module

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

func fromRlp(hex string) *types.Header {
	var h types.Header
	err := rlp.DecodeBytes(common.Hex2Bytes(hex), &h)
	if err != nil {
		log.Fatal(err)
	}
	return &h
}
