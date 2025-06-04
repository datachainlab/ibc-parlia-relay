package module

import (
	"context"
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
)

// Maximum header interval to be submitted to LCP
const skip = 100

type queryVerifiableNeighboringEpochHeaderFn = func(context.Context, uint64, uint64) (core.Header, error)

func shouldSubmitBoundaryTimestampHeader(
	ctx context.Context,
	getHeader getHeaderFn,
	trustedBlockNumber uint64,
	latestFinalizedBlockNumber uint64,
	forkSpecs []*ForkSpec) (*uint64, uint64, error) {

	trustedBlock, err := getHeader(ctx, trustedBlockNumber)
	if err != nil {
		return nil, 0, err
	}
	latestFinalizedBlock, err := getHeader(ctx, latestFinalizedBlockNumber)
	if err != nil {
		return nil, 0, err
	}

	latestForkSpec := forkSpecs[len(forkSpecs)-1]
	latestCondition := latestForkSpec.GetHeightOrTimestamp()
	if x, ok := latestCondition.(*ForkSpec_Timestamp); ok {
		if MilliTimestamp(trustedBlock) < x.Timestamp && x.Timestamp < MilliTimestamp(latestFinalizedBlock) {
			boundaryHeight, err := GetBoundaryHeight(ctx, getHeader, latestFinalizedBlock.Number.Uint64(), *latestForkSpec)
			if err != nil {
				return nil, 0, err
			}
			// Must be right before boundary height
			if boundaryHeight.Height == 0 {
				return nil, 0, fmt.Errorf("boundary height not found")
			}

			nextForkBoundaryHeightMinus1 := uint64(boundaryHeight.Height) - 1
			log.GetLogger().InfoContext(ctx, "ForkSpec height required", "ts", x.Timestamp, "height", boundaryHeight, "nextForkBoundaryHeightMinus1", nextForkBoundaryHeightMinus1)
			return &x.Timestamp, nextForkBoundaryHeightMinus1, nil
		}
	}
	return nil, 0, nil
}

func setupHeadersForUpdate(
	ctx context.Context,
	queryVerifiableNeighboringEpochHeader queryVerifiableNeighboringEpochHeaderFn,
	getHeader getHeaderFn,
	clientStateLatestHeight exported.Height,
	latestFinalizedHeader *Header,
	latestHeight exported.Height,
	forkSpecs []*ForkSpec,
) ([]core.Header, error) {
	logger := log.GetLogger()
	logger.DebugContext(ctx, "setupHeadersForUpdate start", "target", latestFinalizedHeader.GetHeight().GetRevisionHeight())
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()

	trustedBlock, err := getHeader(ctx, savedLatestHeight)
	if err != nil {
		return nil, err
	}

	trustedCurrentForkSpec, trustedPreviousForkSpec, err := FindTargetForkSpec(forkSpecs, savedLatestHeight, MilliTimestamp(trustedBlock))
	if err != nil {
		return nil, err
	}
	trustedBoundaryHeight, err := GetBoundaryHeight(ctx, getHeader, savedLatestHeight, *trustedCurrentForkSpec)
	if err != nil {
		return nil, err
	}
	trustedBoundaryEpochs, err := trustedBoundaryHeight.GetBoundaryEpochs(trustedPreviousForkSpec)
	if err != nil {
		return nil, err
	}

	trustedEpochHeight := trustedBoundaryEpochs.CurrentEpochBlockNumber(savedLatestHeight)
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()

	// If the condition is timestamp. we must submit the header with the timestamp
	nextForkBoundaryTs, nextForkBoundaryHeightMinus1, err := shouldSubmitBoundaryTimestampHeader(ctx, getHeader, savedLatestHeight, latestFinalizedHeader.GetHeight().GetRevisionHeight(), forkSpecs)
	if err != nil {
		return nil, err
	}

	firstUnsaved := trustedEpochHeight + skip
	for firstUnsaved <= savedLatestHeight {
		firstUnsaved += skip
	}

	submittingHeights := makeSubmittingHeights(latestFinalizedHeight, savedLatestHeight, firstUnsaved, nextForkBoundaryTs, nextForkBoundaryHeightMinus1)
	logger.DebugContext(ctx, "submitting heights", "heights", submittingHeights, "trusted height", savedLatestHeight, "trusted epoch", trustedEpochHeight, "first unsaved", firstUnsaved)

	trustedHeight := clientStateLatestHeight.GetRevisionHeight()
	for _, submittingHeight := range submittingHeights {
		verifiableHeader, err := setupIntermediateHeader(ctx, queryVerifiableNeighboringEpochHeader, submittingHeight, latestHeight)
		if err != nil {
			return nil, err
		}
		if verifiableHeader == nil {
			logger.ErrorContext(ctx, "[FastFinalityError]", fmt.Errorf("insufficient vote attestation: submittingHeight=%d, trusted=%d", submittingHeight, trustedHeight))
			return withTrustedHeight(ctx, targetHeaders, clientStateLatestHeight), nil
		}
		targetHeaders = append(targetHeaders, verifiableHeader)
		trustedHeight = submittingHeight
		logger.DebugContext(ctx, "setup epoch header", "trusted", trustedHeight, "height", submittingHeight)
	}
	return withTrustedHeight(ctx, append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
}

// Get verifiable headers. This method must be executed at block intervals that do not miss any epochs.
func setupIntermediateHeader(
	ctx context.Context,
	queryVerifiableHeader queryVerifiableNeighboringEpochHeaderFn,
	submittingHeight uint64,
	latestHeight exported.Height,
) (core.Header, error) {
	return queryVerifiableHeader(ctx, submittingHeight, minUint64(submittingHeight+skip, latestHeight.GetRevisionHeight()))
}

func withTrustedHeight(ctx context.Context, targetHeaders []core.Header, clientStateLatestHeight exported.Height) []core.Header {
	logger := log.GetLogger()
	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = toHeight(clientStateLatestHeight)
		} else {
			trustedHeight = toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight

		logger.DebugContext(ctx, "setupHeadersForUpdate end", "target", h.GetHeight(), "trusted", trustedHeight, "headerSize", len(h.(*Header).Headers))
	}
	return targetHeaders
}

func makeSubmittingHeights(latestFinalizedHeight uint64, savedLatestHeight uint64, firstUnsaved uint64, nextForkBoundaryTs *uint64, nextForkBoundaryHeightMinus1 uint64) []uint64 {
	var submittingHeights []uint64
	if latestFinalizedHeight < firstUnsaved {
		if nextForkBoundaryTs != nil && savedLatestHeight < nextForkBoundaryHeightMinus1 && nextForkBoundaryHeightMinus1 < latestFinalizedHeight {
			submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
		}
	} else {
		var temp []uint64
		for epochCandidate := firstUnsaved; epochCandidate < latestFinalizedHeight; epochCandidate += skip {
			temp = append(temp, epochCandidate)
		}
		if nextForkBoundaryTs != nil {
			for i, epochCandidate := range temp {
				if i > 0 {
					if temp[i-1] < nextForkBoundaryHeightMinus1 && nextForkBoundaryHeightMinus1 < epochCandidate {
						submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
					}
				} else if i == 0 {
					if savedLatestHeight < nextForkBoundaryHeightMinus1 && nextForkBoundaryHeightMinus1 < epochCandidate {
						submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
					}
				}
				submittingHeights = append(submittingHeights, epochCandidate)
			}
			if submittingHeights[len(submittingHeights)-1] < nextForkBoundaryHeightMinus1 && nextForkBoundaryHeightMinus1 < latestFinalizedHeight {
				submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
			}
		} else {
			submittingHeights = temp
		}
	}
	return submittingHeights
}
