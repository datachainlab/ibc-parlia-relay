package module

import (
	"context"
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
)

const skip = 100

type queryVerifiableNeighboringEpochHeaderFn = func(uint64, uint64) (core.Header, error)

func shouldSubmitBoundaryTimestampHeader(
	getHeader getHeaderFn,
	trustedBlockNumber uint64,
	latestFinalizedBlockNumber uint64,
	forkSpecs []*ForkSpec) (*uint64, uint64, error) {

	trustedBlock, err := getHeader(context.Background(), trustedBlockNumber)
	if err != nil {
		return nil, 0, err
	}
	latestFinalizedBlock, err := getHeader(context.Background(), latestFinalizedBlockNumber)
	if err != nil {
		return nil, 0, err
	}

	latestForkSpec := forkSpecs[len(forkSpecs)-1]
	latestCondition := latestForkSpec.GetHeightOrTimestamp()
	if x, ok := latestCondition.(*ForkSpec_Timestamp); ok {
		if MilliTimestamp(trustedBlock) < x.Timestamp && x.Timestamp < MilliTimestamp(latestFinalizedBlock) {
			// TODO 対象のheightを出す
			return &x.Timestamp, 0, nil
		}
	}
	return nil, 0, nil
}

func setupHeadersForUpdate(
	queryVerifiableNeighboringEpochHeader queryVerifiableNeighboringEpochHeaderFn,
	getHeader getHeaderFn,
	clientStateLatestHeight exported.Height,
	latestFinalizedHeader *Header,
	latestHeight exported.Height,
	forkSpecs []*ForkSpec,
) ([]core.Header, error) {
	logger := log.GetLogger()
	logger.Debug("setupHeadersForUpdate start", "target", latestFinalizedHeader.GetHeight().GetRevisionHeight())
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()

	trustedBlock, err := getHeader(context.Background(), savedLatestHeight)
	if err != nil {
		return nil, err
	}

	trustedCurrentForkSpec, trustedPreviousForkSpec, err := findTargetForkSpec(forkSpecs, savedLatestHeight, MilliTimestamp(trustedBlock))
	if err != nil {
		return nil, err
	}
	trustedBoundaryHeight, err := getBoundaryHeight(getHeader, savedLatestHeight, *trustedCurrentForkSpec)
	if err != nil {
		return nil, err
	}
	trustedBoundaryEpochs, err := trustedBoundaryHeight.getBoundaryEpochs(*trustedCurrentForkSpec, *trustedPreviousForkSpec)
	if err != nil {
		return nil, err
	}

	trustedEpochHeight := trustedBoundaryEpochs.CurrentEpochBlockNumber(savedLatestHeight)
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()

	// If the condition is timestamp. we must submit the header with the timestamp
	nextForkBoundaryTs, nextForkBoundaryHeightMinus1, err := shouldSubmitBoundaryTimestampHeader(getHeader, savedLatestHeight, latestFinalizedHeader.GetHeight().GetRevisionHeight(), forkSpecs)
	if err != nil {
		return nil, err
	}
	logger.Info("Must set boundary timestamp", "ts", nextForkBoundaryTs, "nextForkBoundaryHeightMinus1", nextForkBoundaryHeightMinus1)

	firstUnsaved := trustedEpochHeight + skip

	var submittingHeights []uint64
	if latestFinalizedHeight < firstUnsaved {
		if nextForkBoundaryTs != nil && nextForkBoundaryHeightMinus1 < latestFinalizedHeight {
			submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
		}
	} else {
		for epochCandidate := firstUnsaved; epochCandidate < latestFinalizedHeight; epochCandidate += skip {
			if nextForkBoundaryTs != nil && nextForkBoundaryHeightMinus1 < epochCandidate {
				submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1, epochCandidate)
			} else {
				submittingHeights = append(submittingHeights, epochCandidate)
			}
		}
		if nextForkBoundaryTs != nil &&
			submittingHeights[len(submittingHeights)-1] < nextForkBoundaryHeightMinus1 &&
			nextForkBoundaryHeightMinus1 < latestFinalizedHeight {
			submittingHeights = append(submittingHeights, nextForkBoundaryHeightMinus1)
		}
	}
	logger.Debug("submitting heights", "heights", submittingHeights)

	trustedHeight := clientStateLatestHeight.GetRevisionHeight()
	for _, submittingHeight := range submittingHeights {
		verifiableHeader, err := setupIntermediateHeader(queryVerifiableNeighboringEpochHeader, submittingHeight, latestHeight)
		if err != nil {
			return nil, err
		}
		if verifiableHeader == nil {
			logger.Error("[FastFinalityError]", fmt.Errorf("insufficient vote attestation: submittingHeight=%d, trusted=%d", submittingHeight, trustedHeight))
			return withTrustedHeight(targetHeaders, clientStateLatestHeight), nil
		}
		targetHeaders = append(targetHeaders, verifiableHeader)
		trustedHeight = submittingHeight
		logger.Debug("setup epoch header", "trusted", trustedHeight, "height", submittingHeight)
	}
	return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
}

func setupIntermediateHeader(
	queryVerifiableHeader queryVerifiableNeighboringEpochHeaderFn,
	submittingHeight uint64,
	latestHeight exported.Height,
) (core.Header, error) {
	return queryVerifiableHeader(submittingHeight, minUint64(submittingHeight+skip, latestHeight.GetRevisionHeight()))
}

func withTrustedHeight(targetHeaders []core.Header, clientStateLatestHeight exported.Height) []core.Header {
	logger := log.GetLogger()
	for i, h := range targetHeaders {
		var trustedHeight clienttypes.Height
		if i == 0 {
			trustedHeight = toHeight(clientStateLatestHeight)
		} else {
			trustedHeight = toHeight(targetHeaders[i-1].GetHeight())
		}
		h.(*Header).TrustedHeight = &trustedHeight

		logger.Debug("setupHeadersForUpdate end", "target", h.GetHeight(), "trusted", trustedHeight, "headerSize", len(h.(*Header).Headers))
	}
	return targetHeaders
}
