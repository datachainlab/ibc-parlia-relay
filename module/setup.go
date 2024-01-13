package module

import (
	"fmt"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ibc-parlia-relay/module/constant"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
)

type queryVerifyingNeighboringEpochHeaderFn = func(uint64, uint64) (core.Header, error)
type queryVerifyingNonNeighboringEpochHeaderFn = func(uint64, uint64, uint64) (core.Header, error)

func setupHeadersForUpdate(
	queryVerifyingNeighboringEpochHeader queryVerifyingNeighboringEpochHeaderFn,
	queryVerifyingNonNeighboringEpochHeader queryVerifyingNonNeighboringEpochHeaderFn,
	getHeader getHeaderFn,
	clientStateLatestHeight exported.Height,
	latestFinalizedHeader *Header,
	latestHeight exported.Height,
) ([]core.Header, error) {
	targetHeaders := make([]core.Header, 0)

	// Needless to update already saved state
	if clientStateLatestHeight.GetRevisionHeight() == latestFinalizedHeader.GetHeight().GetRevisionHeight() {
		return targetHeaders, nil
	}
	savedLatestHeight := clientStateLatestHeight.GetRevisionHeight()
	firstUnsavedEpoch := toEpoch(savedLatestHeight) + constant.BlocksPerEpoch
	latestFinalizedHeight := latestFinalizedHeader.GetHeight().GetRevisionHeight()
	if latestFinalizedHeight <= firstUnsavedEpoch {
		return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
	}

	// Append insufficient epoch blocks
	trustedEpochHeight := toEpoch(savedLatestHeight)
	for epochHeight := firstUnsavedEpoch; epochHeight < latestFinalizedHeight; epochHeight += constant.BlocksPerEpoch {
		if epochHeight == trustedEpochHeight+constant.BlocksPerEpoch {
			verifiableEpoch, err := setupNeighboringEpochHeader(getHeader, queryVerifyingNeighboringEpochHeader, epochHeight, latestHeight)
			if err != nil {
				return nil, err
			}
			if verifiableEpoch == nil {
				// not found -> non-neighboring epoch
				continue
			}
			targetHeaders = append(targetHeaders, verifiableEpoch)
		} else {
			verifiableEpoch, err := setupNonNeighboringEpochHeader(getHeader, queryVerifyingNonNeighboringEpochHeader, epochHeight, trustedEpochHeight, latestHeight)
			if err != nil {
				return nil, err
			}
			if verifiableEpoch == nil {
				// not found -> next non-neighboring epoch
				continue
			}
			targetHeaders = append(targetHeaders, verifiableEpoch)
		}
		trustedEpochHeight = epochHeight
	}

	if !isEpoch(latestFinalizedHeight) {
		if trustedEpochHeight < toEpoch(latestFinalizedHeight) {
			// ex) trusted = 200, latest 401, not append latest because it can not be verified
			return withTrustedHeight(append(targetHeaders), clientStateLatestHeight), nil
		}
		return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
	}

	if trustedEpochHeight+constant.BlocksPerEpoch == latestFinalizedHeight {
		// neighboring epoch : ex) prevSavedEpoch = 200, latest 400
		return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
	}

	//Refresh latest finalized header
	latestVerifiableHeader, err := setupNonNeighboringEpochHeader(getHeader, queryVerifyingNonNeighboringEpochHeader, latestFinalizedHeight, trustedEpochHeight, latestHeight)
	if err != nil {
		return nil, err
	}
	if latestVerifiableHeader == nil {
		// No finalized header after checkpoint. latest can not be verified.
		return withTrustedHeight(targetHeaders, clientStateLatestHeight), nil
	}
	return withTrustedHeight(append(targetHeaders, latestFinalizedHeader), clientStateLatestHeight), nil
}

func setupNeighboringEpochHeader(
	getHeader getHeaderFn,
	queryVerifyingHeader queryVerifyingNeighboringEpochHeaderFn,
	epochHeight uint64,
	latestHeight exported.Height,
) (core.Header, error) {
	// ex) trusted(prevSaved = 200), epochHeight = 400 must be finalized by min(610,latest)
	// neighboring epoch needs block before checkpoint
	currentValidatorSet, err := queryValidatorSet(getHeader, epochHeight)
	if err != nil {
		return nil, fmt.Errorf("setupHeadersForUpdate failed to get checkpoint : height=%d : %+v", epochHeight, err)
	}
	nextCheckpoint := currentValidatorSet.Checkpoint(epochHeight + constant.BlocksPerEpoch)
	limit := minUint64(nextCheckpoint-1, latestHeight.GetRevisionHeight())
	return queryVerifyingHeader(epochHeight, limit)
}

func setupNonNeighboringEpochHeader(
	getHeader getHeaderFn,
	queryVerifyingHeader queryVerifyingNonNeighboringEpochHeaderFn,
	epochHeight uint64,
	trustedEpoch uint64,
	latestHeight exported.Height,
) (core.Header, error) {
	// ex) trusted(prevSaved = 200), epochHeight = 600 must be finalized from 611 〜 min(latest,810)
	currentValidatorSet, err := queryValidatorSet(getHeader, epochHeight)
	if err != nil {
		return nil, fmt.Errorf("setupNonNeighboringEpochHeader failed to get curent validator : trustedEpoch=%d : %+v", trustedEpoch, err)
	}
	nextCheckpoint := currentValidatorSet.Checkpoint(epochHeight + constant.BlocksPerEpoch)
	limit := minUint64(nextCheckpoint-1, latestHeight.GetRevisionHeight())
	trustedValidatorSet, err := queryValidatorSet(getHeader, trustedEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted validator set: trustedEpoch=%d : %+v", trustedEpoch, err)
	}
	if !trustedValidatorSet.Contains(currentValidatorSet) {
		return nil, fmt.Errorf("currentValidatorSet must contain 1/3 trusted validator set : epoch=%d, trustedEpoch=%d", epochHeight, trustedEpoch)
	}

	// Headers after checkpoint are required to verify
	previousValidatorSet, err := queryValidatorSet(getHeader, epochHeight-constant.BlocksPerEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous validator set: height=%d : %+v", epochHeight, err)
	}
	checkpoint := previousValidatorSet.Checkpoint(epochHeight)
	h, err := queryVerifyingHeader(epochHeight, limit, checkpoint)
	if h != nil {
		h.(*Header).TrustedCurrentValidators = trustedValidatorSet
	}
	return h, err
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

		logger.Debug("setupHeadersForUpdate", "target", h.GetHeight(), "trusted", trustedHeight, "headerSize", len(h.(*Header).Headers))
	}
	return targetHeaders
}
