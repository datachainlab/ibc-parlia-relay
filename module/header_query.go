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

// queryFinalizedHeader returns finalized header sequence
// ex)
// 302 -> target -> 301 -> target -> 300
// 302 -> source ------------------> 300
//
// 302 -> target -> 300 -> target -> 298
// 302 -> source ------------------> 298
func queryFinalizedHeader(ctx context.Context, fn getHeaderFn, height uint64, limitHeight uint64, forkSpecs []*ForkSpec) ([]*ETHHeader, error) {
	var ethHeaders []*ETHHeader
	for i := height; i+2 <= limitHeight; i++ {
		finalizedBlock, finalizedETHHeader, _, err := queryETHHeader(ctx, fn, i)
		if err != nil {
			return nil, err
		}
		ethHeaders = append(ethHeaders, finalizedETHHeader)

		currentForkSpec, _, err := FindTargetForkSpec(forkSpecs, finalizedBlock.Number.Uint64(), MilliTimestamp(finalizedBlock))
		if err != nil {
			return nil, err
		}

		// child: descendant whose vote.TargetNumber == finalizedBlock.Number
		var childList []*ETHHeader
		for j := i + 1; j+1 <= limitHeight; j++ {
			childBlock, childETHHeader, childVote, err := queryETHHeader(ctx, fn, i+1)
			if err != nil {
				return nil, err
			}
			childList = append(childList, childETHHeader)
			if childVote == nil {
				continue
			}
			if childVote.Data.TargetNumber != finalizedBlock.Number.Uint64() {
				continue
			}

			// grandChild: descendant whose vote.TargetNumber == mid.Number and vote.SourceNumber == finalizedBlock.Number
			var grandChildList []*ETHHeader
			for k := uint64(1); k <= uint64(currentForkSpec.KAncestorGenerationDepth); k++ {
				grandChildIndex := j + k
				if grandChildIndex > limitHeight {
					break
				}
				_, grandChildETHHeader, grandChildVote, err := queryETHHeader(ctx, fn, grandChildIndex)
				if err != nil {
					return nil, err
				}
				grandChildList = append(grandChildList, grandChildETHHeader)
				if grandChildVote == nil {
					continue
				}
				if grandChildVote.Data.SourceNumber == finalizedBlock.Number.Uint64() ||
					grandChildVote.Data.TargetNumber == childBlock.Number.Uint64() {
					// Found headers.
					// ELC Requires all sequential headers from starting header
					return append(append(ethHeaders, childList...), grandChildList...), nil
				}
			}
		}
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
