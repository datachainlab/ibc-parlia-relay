package module

import (
	"fmt"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
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
	log.Printf("path = %s, slot = %s, storageRoot = %s, proof = %s", path, marshaledSlot, strings.Join(v, ","), strings.Join(vv, ","))
	return stateProof.StorageProofRLP[0], nil
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
