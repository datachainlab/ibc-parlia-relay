package module

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperledger-labs/yui-relayer/log"
)

type getHeaderFn func(context.Context, uint64) (*types.Header, error)

func queryLatestFinalizedHeader(ctx context.Context, getHeader getHeaderFn, latestBlockNumber uint64, forkSpecs []*ForkSpec) (uint64, []*ETHHeader, error) {
	logger := log.GetLogger()
	for i := latestBlockNumber; i > 0; i-- {
		header, err := getHeader(ctx, i)
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

		logger.DebugContext(ctx, "Try to seek verifying headers to finalize", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)

		headers, err := queryFinalizedHeader(ctx, getHeader, probablyFinalized, latestBlockNumber, forkSpecs)
		if err != nil {
			return 0, nil, err
		}
		if headers != nil {
			return probablyFinalized, headers, nil
		}
		logger.DebugContext(ctx, "Failed to seek verifying headers to finalize. So seek previous finalized header.", "probablyFinalized", probablyFinalized, "latest", latestBlockNumber)
	}
	return 0, nil, fmt.Errorf("no finalized header found: %d", latestBlockNumber)
}

func queryFinalizedHeader(ctx context.Context, fn getHeaderFn, height uint64, limitHeight uint64, forkSpecs []*ForkSpec) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := height; i+2 <= limitHeight; i++ {
		targetBlock, targetETHHeader, _, err := queryETHHeader(ctx, fn, i)
		if err != nil {
			return nil, err
		}
		currentForkSpec, _, err := FindTargetForkSpec(forkSpecs, targetBlock.Number.Uint64(), MilliTimestamp(targetBlock))
		if err != nil {
			return nil, err
		}

		childBlock, childETHHeader, childVote, err := queryETHHeader(ctx, fn, i+1)
		if err != nil {
			return nil, err
		}
		var descendants []*ETHHeader
		for j := uint64(0); j < uint64(currentForkSpec.KAncestorGenerationDepth); j++ {
			descendantIndex := i + 2 + j
			if descendantIndex > limitHeight {
				log.GetLogger().DebugContext(ctx, "Insufficient verifying headers to finalize. no valid descendant found for target.", "height", height, "limit", limitHeight, "target", targetBlock.Number)
				return nil, nil
			}

			_, descendantHeader, descendantVote, err := queryETHHeader(ctx, fn, descendantIndex)
			if err != nil {
				return nil, err
			}

			// Ensure valida vote relation
			if childVote == nil || descendantVote == nil ||
				descendantVote.Data.SourceNumber != targetBlock.Number.Uint64() ||
				descendantVote.Data.SourceNumber != childVote.Data.TargetNumber ||
				descendantVote.Data.TargetNumber != childBlock.Number.Uint64() {
				// Append to verify header sequence
				descendants = append(ethHeaders, descendantHeader)
				continue
			}
			return append(append(ethHeaders, targetETHHeader, childETHHeader), descendants...), nil
		}
		// Append to verify header sequence
		ethHeaders = append(ethHeaders, targetETHHeader)
	}
	log.GetLogger().DebugContext(ctx, "Insufficient verifying headers to finalize", "height", height, "limit", limitHeight)
	return nil, nil
}

func queryETHHeader(ctx context.Context, fn getHeaderFn, height uint64) (*types.Header, *ETHHeader, *VoteAttestation, error) {
	block, err := fn(ctx, height)
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
