package module

import (
	"bytes"
	"fmt"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gogo/protobuf/proto"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
)

func (pr *Prover) getAccountProof(height int64) ([]byte, common.Hash, error) {
	stateProof, err := pr.chain.GetProof(
		pr.chain.IBCAddress(),
		nil,
		big.NewInt(height),
	)
	if err != nil {
		return nil, [32]byte{}, err
	}
	hash := stateProof.StorageHash[:]
	v := make([]string, len(hash))
	for i, e := range hash {
		v[i] = strconv.Itoa(int(e))
	}

	log.Printf("storageRoot = %s", strings.Join(v, ","))
	return stateProof.AccountProofRLP, common.BytesToHash(stateProof.StorageHash[:]), nil
}

func (pr *Prover) getStateCommitmentProof(path []byte, height exported.Height) ([]byte, error) {
	// calculate slot for commitment
	slot := crypto.Keccak256Hash(append(
		crypto.Keccak256Hash(path).Bytes(),
		common.Hash{}.Bytes()...,
	))
	marshaledSlot, err := slot.MarshalText()
	if err != nil {
		return nil, err
	}

	// call eth_getProof
	stateProof, err := pr.chain.GetProof(
		pr.chain.IBCAddress(),
		[][]byte{marshaledSlot},
		big.NewInt(int64(height.GetRevisionHeight())),
	)
	if err != nil {
		return nil, err
	}
	hash := stateProof.StorageHash[:]
	v := make([]string, len(hash))
	for i, e := range hash {
		v[i] = strconv.Itoa(int(e))
	}
	vv := make([]string, len(stateProof.StorageProofRLP[0]))
	for i, e := range stateProof.StorageProofRLP[0] {
		vv[i] = strconv.Itoa(int(e))
	}
	//	log.Printf("path = %s, slot = %s, storageRoot = %s, proof = %s", path, marshaledSlot, strings.Join(v, ","), strings.Join(vv, ","))
	return stateProof.StorageProofRLP[0], nil
}

func (pr *Prover) getStateCommitmentProofWithRoot(path []byte, height exported.Height) (common.Hash, []byte, error) {
	// calculate slot for commitment
	slot := crypto.Keccak256Hash(append(
		crypto.Keccak256Hash(path).Bytes(),
		common.Hash{}.Bytes()...,
	))
	marshaledSlot, err := slot.MarshalText()
	if err != nil {
		return common.Hash{}, nil, err
	}

	// call eth_getProof
	stateProof, err := pr.chain.GetProof(
		pr.chain.IBCAddress(),
		[][]byte{marshaledSlot},
		big.NewInt(int64(height.GetRevisionHeight())),
	)
	if err != nil {
		return common.Hash{}, nil, err
	}
	hash := stateProof.StorageHash[:]
	v := make([]string, len(hash))
	for i, e := range hash {
		v[i] = strconv.Itoa(int(e))
	}
	vv := make([]string, len(stateProof.StorageProofRLP[0]))
	for i, e := range stateProof.StorageProofRLP[0] {
		vv[i] = strconv.Itoa(int(e))
	}
	log.Printf("path = %s, slot = %s, storageRoot = %s, proof = %s", path, marshaledSlot, strings.Join(v, ","), strings.Join(vv, ","))
	return common.BytesToHash(hash), stateProof.StorageProofRLP[0], nil
}

type proofList struct {
	list  [][]byte
	index int
}

func (p *proofList) Has([]byte) (bool, error) {
	panic("not implemented")
}

func (p *proofList) Get([]byte) ([]byte, error) {
	if p.index >= len(p.list) {
		return nil, fmt.Errorf("out of index")
	}
	v := p.list[p.index]
	p.index += 1
	return v, nil
}

func verifyProof(rootHash common.Hash, key []byte, proof [][]byte) ([]byte, error) {
	return trie.VerifyProof(rootHash, key, &proofList{list: proof, index: 0})
}

func verifyMembership(root common.Hash, bzValueProof []byte, path string, commitment []byte) error {
	var rawValueProof [][][]byte
	if err := rlp.DecodeBytes(bzValueProof, &rawValueProof); err != nil {
		return fmt.Errorf("rlp.DecodeBytes(bzValueProof, ...) failed: %v", err)
	}
	var valueProof [][]byte
	for _, raw := range rawValueProof {
		if bz, err := rlp.EncodeToBytes(raw); err != nil {
			return fmt.Errorf("rlp.EncodeToBytes(raw) failed: %v", err)
		} else {
			valueProof = append(valueProof, bz)
		}
	}

	key := crypto.Keccak256(crypto.Keccak256(append(crypto.Keccak256([]byte(path)), common.Hash{}.Bytes()...)))

	recoveredCommitment, err := verifyProof(root, key, valueProof)
	if err != nil {
		return fmt.Errorf("verifyProof failed: %v", err)
	}

	rlpCommitment, err := rlp.EncodeToBytes(commitment)
	if err != nil {
		return fmt.Errorf("rlp.EncodeToBytes(commitment) failed: %v", err)
	}
	if !bytes.Equal(recoveredCommitment, rlpCommitment) {
		return fmt.Errorf("value unmatch: %v(length=%d) != %v(length=%d)",
			recoveredCommitment, len(recoveredCommitment),
			rlpCommitment, len(rlpCommitment),
		)
	}
	return nil
}

func messageToCommitment(msg proto.Message) ([]byte, error) {
	marshaled, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(marshaled), nil
}
