package module

import (
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module/env"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tendermint/libs/json"
	"log"
	"strconv"
	"strings"
)

const extraVanity = 32
const extraSeal = 65
const validatorBytesLengthBeforeLuban = 20
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

	type Pretty struct {
		Raw             []string
		Header          []*types.Header
		AccountProof    []string
		ProtoBufTypeURL string
		ProtoBufValue   string
		ProtoBufMarshal string
	}

	prettyByteArray := func(data []byte) string {
		ret := make([]string, len(data))
		for i, e := range data {
			ret[i] = strconv.Itoa(int(e))
		}
		return fmt.Sprintf("vec![%s]", strings.Join(ret, ","))
	}

	pretty := &Pretty{}

	pretty.Raw = make([]string, len(h.Headers))
	for i, e := range h.Headers {
		pretty.Raw[i] = prettyByteArray(e.Header)
	}

	accountProof, err := h.decodeAccountProof()
	if err == nil {
		pretty.AccountProof = make([]string, len(accountProof))
		for i, e := range accountProof {
			pretty.AccountProof[i] = prettyByteArray(e)
		}
	}
	headers, err := h.decodeEthHeaders()
	if err == nil {
		pretty.Header = headers
	}
	anyHeader, err := clienttypes.PackHeader(h)
	if err == nil {
		pretty.ProtoBufTypeURL = anyHeader.TypeUrl
		pretty.ProtoBufValue = prettyByteArray(anyHeader.Value)
		if msg, mErr := anyHeader.Marshal(); mErr == nil {
			pretty.ProtoBufMarshal = prettyByteArray(msg)
		}
	}
	value, _ := json.MarshalIndent(pretty, "  ", "  ")
	return string(value)
}

func extractValidatorSet(h *types.Header) ([][]byte, error) {
	extra := h.Extra
	if len(extra) < extraVanity+extraSeal {
		return nil, fmt.Errorf("invalid extra length : %d", h.Number.Uint64())
	}
	var validatorSet [][]byte
	validators := extra[extraVanity : len(extra)-extraSeal]
	if h.Number.Uint64() >= env.LubanFork {
		validatorCount := int(validators[0])
		validatorsWithBLS := validators[1 : validatorCount*validatorBytesLength]
		for i := 0; i < validatorCount; i++ {
			start := validatorBytesLength * i
			validatorWithBLS := validatorsWithBLS[start : start+validatorBytesLength]
			validatorSet = append(validatorSet, validatorWithBLS[:validatorBytesLengthBeforeLuban])
		}
	} else {
		validatorCount := len(validators) / validatorBytesLengthBeforeLuban
		for i := 0; i < validatorCount; i++ {
			start := validatorBytesLengthBeforeLuban * i
			validatorSet = append(validatorSet, validators[start:start+validatorBytesLengthBeforeLuban])
		}
	}
	return validatorSet, nil
}
