package module

import (
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tendermint/libs/json"
	"log"
	"strconv"
	"strings"
)

const epochBlockPeriod = 200
const extraVanity = 32
const extraSeal = 65
const validatorBytesLength = 20

// Parlia TODO client_type
const Parlia string = "99-parlia"

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
	if _, err := h.decodeAccountProof(); err != nil {
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

func (h *Header) decodeAccountProof() ([][]byte, error) {
	var decodedProof [][][]byte
	if err := rlp.DecodeBytes(h.AccountProof, &decodedProof); err != nil {
		return nil, err
	}
	var accountProof [][]byte
	for i := range decodedProof {
		b, err := rlp.EncodeToBytes(decodedProof[i])
		if err != nil {
			return nil, err
		}
		accountProof = append(accountProof, b)
	}
	return accountProof, nil
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

func (h *Header) Account(path common.Address) (*types.StateAccount, error) {
	target, err := h.Target()
	if err != nil {
		return nil, err
	}
	decodedAccountProof, err := h.decodeAccountProof()
	if err != nil {
		return nil, err
	}
	rlpAccount, err := verifyProof(
		target.Root,
		crypto.Keccak256Hash(path.Bytes()).Bytes(),
		decodedAccountProof,
	)
	if err != nil {
		return nil, err
	}
	var account types.StateAccount
	if err = rlp.DecodeBytes(rlpAccount, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func (h *Header) ToPrettyString() string {
	contents := make([]string, 0)
	toString := func(data []byte) string {
		v := make([]string, len(data))
		for i, e := range data {
			v[i] = strconv.Itoa(int(e))
		}
		return strings.Join(v, ",")
	}

	for i, e := range h.Headers {
		contents = append(contents, fmt.Sprintf("raw_header[%d] = [%s]\n", i, toString(e.Header)))
	}
	accountProof, err := h.decodeAccountProof()
	if err == nil {
		for i, e := range accountProof {
			contents = append(contents, fmt.Sprintf("decoded_account_proof[%d] = [%s]\n", i, toString(e)))
		}
	}
	headers, err := h.decodeEthHeaders()
	if err == nil {
		decodedHeaders, err := json.MarshalIndent(headers, "", " ")
		if err == nil {
			contents = append(contents, fmt.Sprintf("decoded_header = %s\n", decodedHeaders))
		}
	}
	anyHeader, err := clienttypes.PackHeader(h)
	if err == nil {
		contents = append(contents, fmt.Sprintf("protobuf type url = %s\n", anyHeader.TypeUrl))
		contents = append(contents, fmt.Sprintf("protobuf value = [%s]\n", toString(anyHeader.Value)))
		msg, err := anyHeader.Marshal()
		if err == nil {
			contents = append(contents, fmt.Sprintf("protobuf marshal = [%s]\n", toString(msg)))
		}
	}
	return strings.Join(contents, "")
}

func extractValidatorSet(h *types.Header) ([][]byte, error) {
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, fmt.Errorf("invalid extra length")
	}
	var validatorSet [][]byte
	validators := extra[extraVanity : len(extra)-extraSeal]
	validatorCount := len(validators) / validatorBytesLength
	for i := 0; i < validatorCount; i++ {
		start := validatorBytesLength * i
		validatorSet = append(validatorSet, validators[start:start+validatorBytesLength])
	}
	return validatorSet, nil
}
