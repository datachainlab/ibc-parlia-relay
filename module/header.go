package module

import (
	"context"
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/core"
)

const epochBlockPeriod = 200
const extraVanity = 32
const extraSeal = 65
const validatorBytesLength = 20

// Parlia TODO client_type
const Parlia string = "99-parlia"

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

type HeaderI interface {
	core.HeaderI
	Target() *types.Header
	Account(path common.Address) (*types.StateAccount, error)
	ValidatorSet() ([][]byte, error)
}

type defaultHeader struct {
	*Header
	revisionNumber uint64

	// cache
	decodedTargetHeader *types.Header
	decodedAccountProof [][]byte
}

func NewHeader(revisionNumber uint64, header *Header) (HeaderI, error) {
	decodedHeaders, err := header.decodeEthHeaders()
	if err != nil {
		return nil, err
	}
	if len(decodedHeaders) == 0 {
		return nil, fmt.Errorf("invalid header length")
	}
	decodedAccountProof, err := header.decodeAccountProof()
	if err != nil {
		return nil, err
	}
	decodedTargetHeader := decodedHeaders[0]

	return &defaultHeader{
		revisionNumber:      revisionNumber,
		Header:              header,
		decodedTargetHeader: decodedTargetHeader,
		decodedAccountProof: decodedAccountProof,
	}, nil
}

func (h *defaultHeader) Target() *types.Header {
	return h.decodedTargetHeader
}

func (h *defaultHeader) ValidatorSet() ([][]byte, error) {
	extra := h.decodedTargetHeader.Extra
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

func (h *defaultHeader) Account(path common.Address) (*types.StateAccount, error) {
	rlpAccount, err := verifyProof(
		h.decodedTargetHeader.Root,
		crypto.Keccak256Hash(path.Bytes()).Bytes(),
		h.decodedAccountProof,
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

func (*defaultHeader) ClientType() string {
	return Parlia
}

func (h *defaultHeader) GetHeight() exported.Height {
	return clienttypes.NewHeight(h.revisionNumber, h.decodedTargetHeader.Number.Uint64())
}

func (h *defaultHeader) ValidateBasic() error {
	if h.Header == nil || h.decodedTargetHeader == nil || h.decodedAccountProof == nil {
		return fmt.Errorf("invalid header")
	}
	return nil
}

type HeaderReader interface {
	QueryETHHeaders(height uint64) ([]*ETHHeader, error)
}

type headerReader struct {
	blockByNumber func(ctx context.Context, number uint64) (*types.Header, error)
}

func NewHeaderReader(blockByNumber func(ctx context.Context, number uint64) (*types.Header, error)) HeaderReader {
	//TODO cache
	return &headerReader{
		blockByNumber: blockByNumber,
	}
}

// QueryETHHeaders returns the header corresponding to the height
func (pr *headerReader) QueryETHHeaders(height uint64) ([]*ETHHeader, error) {
	epochCount := height / epochBlockPeriod
	if epochCount > 0 {
		previousEpochHeight := (epochCount - 1) * epochBlockPeriod
		previousEpochBlock, err := pr.blockByNumber(context.TODO(), previousEpochHeight)
		if err != nil {
			return nil, err
		}
		threshold := pr.requiredCountToFinalize(previousEpochBlock)
		if height%epochBlockPeriod < uint64(threshold) {
			// before checkpoint
			return pr.getETHHeaders(height, threshold)
		}
	}
	// genesis count or after checkpoint
	lastEpochNumber := epochCount * epochBlockPeriod
	currentEpochBlock, err := pr.blockByNumber(context.TODO(), uint64(lastEpochNumber))
	if err != nil {
		return nil, err
	}
	return pr.getETHHeaders(height, pr.requiredCountToFinalize(currentEpochBlock))
}

func (pr *headerReader) requiredCountToFinalize(header *types.Header) int {
	validators := len(header.Extra[extraVanity:len(header.Extra)-extraSeal]) / validatorBytesLength
	if validators%2 == 1 {
		return validators/2 + 1
	} else {
		return validators / 2
	}
}

func (pr *headerReader) getETHHeaders(start uint64, requiredCountToFinalize int) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := 0; i < requiredCountToFinalize; i++ {
		block, err := pr.blockByNumber(context.TODO(), uint64(i)+start)
		if err != nil {
			return nil, err
		}
		header, err := pr.newETHHeader(block)
		if err != nil {
			return nil, err
		}
		ethHeaders = append(ethHeaders, header)
	}
	return ethHeaders, nil
}

func (pr *headerReader) newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}
