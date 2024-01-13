package module

import (
	"fmt"
	"log"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const extraVanity = 32
const extraSeal = 65
const validatorBytesLength = 68

// Parlia TODO client_type
const Parlia string = "xx-parlia"

func (*Header) ClientType() string {
	return Parlia
}

func (h *Header) GetHeight() exported.Height {
	target, err := h.Target()
	if err != nil {
		log.Panicf("invalid header: %v", h)
	}
	//TODO revision number
	return clienttypes.NewHeight(0, target.Number.Uint64())
}

func (h *Header) ValidateBasic() error {
	if _, err := h.Target(); err != nil {
		return err
	}
	if _, err := decodeAccountProof(h.AccountProof); err != nil {
		return err
	}
	return nil
}

func (h *Header) decodeEthHeaders() ([]*types.Header, error) {
	ethHeaders := make([]*types.Header, len(h.Headers))
	for i, e := range h.Headers {
		var ethHeader types.Header
		if err := rlp.DecodeBytes(e.Header, &ethHeader); err != nil {
			return nil, err
		}
		ethHeaders[i] = &ethHeader
	}
	return ethHeaders, nil
}

func (h *Header) Target() (*types.Header, error) {
	decodedHeaders, err := h.decodeEthHeaders()
	if err != nil {
		return nil, err
	}
	if len(decodedHeaders) == 0 {
		return nil, fmt.Errorf("invalid header length")
	}
	return decodedHeaders[0], nil
}

func (h *Header) Last() (*types.Header, error) {
	decodedHeaders, err := h.decodeEthHeaders()
	if err != nil {
		return nil, err
	}
	if len(decodedHeaders) == 0 {
		return nil, fmt.Errorf("invalid header length")
	}
	return decodedHeaders[len(decodedHeaders)-1], nil
}

func (h *Header) Account(path common.Address) (*types.StateAccount, error) {
	target, err := h.Target()
	if err != nil {
		return nil, err
	}
	return verifyAccount(target, h.AccountProof, path)
}
