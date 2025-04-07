package module

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	blsPublicKeyLength  = 48
	blsSignatureLength  = 96
	validatorNumberSize = 1
	turnLengthLength    = 1
)

type BLSPublicKey [blsPublicKeyLength]byte
type BLSSignature [blsSignatureLength]byte
type ValidatorsBitSet uint64

type VoteAttestation struct {
	VoteAddressSet uint64
	AggSignature   BLSSignature
	Data           *VoteData
	Extra          []byte
}

type VoteData struct {
	SourceNumber uint64
	SourceHash   common.Hash
	TargetNumber uint64
	TargetHash   common.Hash
}

func getVoteAttestationFromHeader(header *types.Header) (*VoteAttestation, error) {
	if len(header.Extra) <= extraVanity+extraSeal {
		return nil, nil
	}

	isEpoch := true
	if _, _, err := extractValidatorSetAndTurnLength(header); err != nil {
		isEpoch = false
	}

	var attestationBytes []byte
	if !isEpoch {
		attestationBytes = header.Extra[extraVanity : len(header.Extra)-extraSeal]
	} else {
		num := int(header.Extra[extraVanity])
		start := extraVanity + validatorNumberSize + num*validatorBytesLength + turnLengthLength
		end := len(header.Extra) - extraSeal
		if end <= start {
			return nil, nil
		}
		attestationBytes = header.Extra[start:end]
	}

	var attestation VoteAttestation
	if err := rlp.Decode(bytes.NewReader(attestationBytes), &attestation); err != nil {
		return nil, fmt.Errorf("block %d has vote attestation info, decode err: %s", header.Number.Uint64(), err)
	}
	return &attestation, nil
}
