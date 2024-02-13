package module

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/log"
)

type getHeaderFn func(context.Context, uint64) (*types.Header, error)

func queryLatestFinalizedHeader(getHeader getHeaderFn, latestBlockNumber uint64) (uint64, []*ETHHeader, error) {
	logger := log.GetLogger()
	for i := latestBlockNumber; i > 0; i-- {
		header, err := getHeader(context.Background(), i)
		if err != nil {
			return 0, nil, err
		}
		vote, err := getVoteAttestationFromHeader(header)
		if err != nil {
			return 0, nil, err
		}
		if vote == nil {
			continue
		}
		probablyFinalized := vote.Data.SourceNumber

		logger.Debug("Try to seek verifying headers to finalize", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)

		headers, err := queryFinalizedHeader(getHeader, probablyFinalized, latestBlockNumber)
		if err != nil {
			return 0, nil, err
		}
		if headers != nil {
			return probablyFinalized, headers, nil
		}
		logger.Debug("Failed to seek verifying headers to finalize. So seek previous finalized header.", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)
	}
	return 0, nil, fmt.Errorf("no finalized header found: %d", latestBlockNumber)
}

func queryFinalizedHeader(fn getHeaderFn, height uint64, limitHeight uint64) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := height; i+2 <= limitHeight; i++ {
		targetBlock, targetETHHeader, _, err := queryETHHeader(fn, i)
		if err != nil {
			return nil, err
		}
		childBlock, childETHHeader, childVote, err := queryETHHeader(fn, i+1)
		if err != nil {
			return nil, err
		}
		_, grandChildETHHeader, grandChildVote, err := queryETHHeader(fn, i+2)
		if err != nil {
			return nil, err
		}

		if childVote == nil || grandChildVote == nil ||
			grandChildVote.Data.SourceNumber != targetBlock.Number.Uint64() ||
			grandChildVote.Data.SourceNumber != childVote.Data.TargetNumber ||
			grandChildVote.Data.TargetNumber != childBlock.Number.Uint64() {
			// Append to verify header sequence
			ethHeaders = append(ethHeaders, targetETHHeader)
			continue
		}
		return append(ethHeaders, targetETHHeader, childETHHeader, grandChildETHHeader), nil
	}
	log.GetLogger().Debug("Insufficient verifying headers to finalize", "height", height, "limit", limitHeight)
	return nil, nil
}

func queryFinalizedHeaderAfterCheckpoint(fn getHeaderFn, height uint64, limitHeight uint64, checkpoint uint64) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := height; i < checkpoint; i++ {
		_, h, _, err := queryETHHeader(fn, i)
		if err != nil {
			return nil, err
		}
		ethHeaders = append(ethHeaders, h)
	}
	afterCheckpoint, err := queryFinalizedHeader(fn, checkpoint, limitHeight)
	if err != nil {
		return nil, err
	}
	if afterCheckpoint == nil {
		return nil, nil
	}
	return append(ethHeaders, afterCheckpoint...), nil
}

func queryETHHeader(fn getHeaderFn, height uint64) (*types.Header, *ETHHeader, *VoteAttestation, error) {
	block, err := fn(context.TODO(), height)
	if err != nil {
		return nil, nil, nil, err
	}
	ethHeader, err := newETHHeader(block)
	if err != nil {
		return nil, nil, nil, err
	}
	vote, err := getVoteAttestationFromHeader(block)
	if err != nil {
		return nil, nil, nil, err
	}
	return block, ethHeader, vote, err
}

func newETHHeader(header *types.Header) (*ETHHeader, error) {
	rlpHeader, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}
	return &ETHHeader{Header: rlpHeader}, nil
}
