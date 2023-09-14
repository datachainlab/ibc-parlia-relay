package module

import (
	"bytes"
	"fmt"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	BLSPublicKeyLength  = 48
	BLSSignatureLength  = 96
	validatorNumberSize = 1
)

type BLSPublicKey [BLSPublicKeyLength]byte
type BLSSignature [BLSSignatureLength]byte
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

	var attestationBytes []byte
	if header.Number.Uint64()%constant.BlocksPerEpoch != 0 {
		attestationBytes = header.Extra[extraVanity : len(header.Extra)-extraSeal]
	} else {
		num := int(header.Extra[extraVanity])
		if len(header.Extra) <= extraVanity+extraSeal+validatorNumberSize+num*validatorBytesLength {
			return nil, nil
		}
		start := extraVanity + validatorNumberSize + num*validatorBytesLength
		end := len(header.Extra) - extraSeal
		attestationBytes = header.Extra[start:end]
	}

	var attestation VoteAttestation
	if err := rlp.Decode(bytes.NewReader(attestationBytes), &attestation); err != nil {
		return nil, fmt.Errorf("block %d has vote attestation info, decode err: %s", header.Number.Uint64(), err)
	}
	return &attestation, nil
}