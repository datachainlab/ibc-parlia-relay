package module

import (
	"bytes"
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/hyperledger-labs/yui-relayer/core"
	"math/big"

	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

func (pr *Prover) getAccountProof(height uint64) ([]byte, common.Hash, error) {
	stateProof, err := pr.chain.GetProof(
		pr.chain.IBCAddress(),
		nil,
		big.NewInt(0).SetUint64(height),
	)
	if err != nil {
		return nil, common.Hash{}, fmt.Errorf("failed to get account proof %+v", err)
	}
	return stateProof.AccountProofRLP, common.BytesToHash(stateProof.StorageHash[:]), nil
}

func (pr *Prover) getStateCommitmentProof(path []byte, height exported.Height) ([]byte, error) {
	// calculate slot for commitment
	storageKey := crypto.Keccak256Hash(append(
		crypto.Keccak256Hash(path).Bytes(),
		IBCCommitmentsSlot.Bytes()...,
	))
	storageKeyHex, err := storageKey.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal slot: height = %d, %+v", height.GetRevisionHeight(), err)
	}

	// call eth_getProof
	stateProof, err := pr.chain.GetProof(
		pr.chain.IBCAddress(),
		[][]byte{storageKeyHex},
		big.NewInt(int64(height.GetRevisionHeight())),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get state commitment proof : address = %s, height = %d, slot = %v, %+v",
			pr.chain.IBCAddress(), height.GetRevisionHeight(), storageKeyHex, err)
	}
	return stateProof.StorageProofRLP[0], nil
}

func (pr *Prover) GetStorageRoot(header *types.Header) (common.Hash, error) {
	rlpAccountProof, _, err := pr.getAccountProof(header.Number.Uint64())
	if err != nil {
		return common.Hash{}, err
	}
	stateAccount, err := verifyAccount(header, rlpAccountProof, pr.chain.IBCAddress())
	if err != nil {
		return common.Hash{}, err
	}
	return stateAccount.Root, nil
}

// ProveHostConsensusState returns an existence proof of the consensus state at `height`
// This proof would be ignored in ibc-go, but it is required to `getSelfConsensusState` of ibc-solidity.
func (pr *Prover) ProveHostConsensusState(ctx core.QueryContext, height exported.Height, consensusState exported.ConsensusState) (proof []byte, err error) {
	return clienttypes.MarshalConsensusState(pr.chain.Codec(), consensusState)
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

func decodeAccountProof(encodedAccountProof []byte) ([][]byte, error) {
	var decodedProof [][][]byte
	if err := rlp.DecodeBytes(encodedAccountProof, &decodedProof); err != nil {
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

func verifyAccount(target *types.Header, accountProof []byte, path common.Address) (*types.StateAccount, error) {
	decodedAccountProof, err := decodeAccountProof(accountProof)
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

func withValidators(headerFn getHeaderFn, height uint64, ethHeaders []*ETHHeader) (core.Header, error) {

	header := &Header{
		Headers: ethHeaders,
	}

	// Get validator set for verify headers
	previousEpoch := getPreviousEpoch(height)
	var previousTurnLength uint8
	var err error
	header.PreviousValidators, previousTurnLength, err = queryValidatorSetAndTurnLength(headerFn, previousEpoch)
	header.PreviousTurnLength = uint32(previousTurnLength)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in previous epoch : number = %d : %+v", previousEpoch, err)
	}
	currentEpoch := getCurrentEpoch(height)
	var currentTurnLength uint8
	header.CurrentValidators, currentTurnLength, err = queryValidatorSetAndTurnLength(headerFn, currentEpoch)
	header.CurrentTurnLength = uint32(currentTurnLength)
	if err != nil {
		return nil, fmt.Errorf("ValidatorSet was not found in current epoch : number= %d : %+v", currentEpoch, err)
	}

	return header, nil
}
