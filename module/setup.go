package module

import (
	"context"
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
)

type queryVerifiableNeighboringEpochHeaderFn = func(uint64, uint64) (core.Header, error)

func setupHeadersForUpdate(
	queryVerifiableNeighboringEpochHeader queryVerifiableNeighboringEpochHeaderFn,
	getHeader getHeaderFn,
	clientStateLatestHeight exported.Height,
	latestFinalizedHeader *Header,
	latestHeight exported.Height,
) ([]core.Header, error) {
	logger := log.GetLogger()
	logger.Debug("setupHeadersForUpdate start", "target", latestFinalizedHeader.GetHeight().GetRevisionHeight())
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()
	firstUnsavedEpoch := toEpoch(savedLatestHeight) + constant.BlocksPerEpoch
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()
	if latestFinalizedHeight < firstUnsavedEpoch {
		return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
	}

	trustedEpochHeight := toEpoch(savedLatestHeight)

	// Append insufficient epoch blocks
	for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
		verifiableEpoch, err := setupNeighboringEpochHeader(getHeader, queryVerifiableNeighboringEpochHeader, epochHeight, trustedEpochHeight, latestHeight)
		if err != nil {
			return nil, err
		}
		if verifiableEpoch == nil {
			logger.Error("[FastFinalityError]", fmt.Errorf("insufficient vote attestation: epochHeight=%d, trustedEpochHeight=%d", epochHeight, trustedEpochHeight))
			return withTrustedHeight(targetHeaders, clientStateLatestHeight), nil
		}
		targetHeaders = append(targetHeaders, verifiableEpoch)
		trustedEpochHeight = epochHeight
		logger.Debug("setup epoch header", "height", epochHeight)
	}
	return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
}

func setupNeighboringEpochHeader(
	getHeader getHeaderFn,
	queryVerifiableHeader queryVerifiableNeighboringEpochHeaderFn,
	epochHeight uint64,
	trustedEpochHeight uint64,
	latestHeight exported.Height,
) (core.Header, error) {
	// neighboring epoch needs block before checkpoint
	currentValidatorSet, currentTurnLength, err := queryValidatorSetAndTurnLength(context.TODO(), getHeader, epochHeight)
	if err != nil {
		return nil, fmt.Errorf("setupNeighboringEpochHeader: failed to get current validator set: epochHeight=%d : %+v", epochHeight, err)
	}
	// ex) trusted(prevSaved = 200), epochHeight = 400 must be finalized by min(610,latest)
	nextCheckpoint := currentValidatorSet.Checkpoint(currentTurnLength) + (epochHeight + constant.BlocksPerEpoch)
	limit := minUint64(nextCheckpoint-1, latestHeight.GetRevisionHeight())
	return queryVerifiableHeader(epochHeight, limit)
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
