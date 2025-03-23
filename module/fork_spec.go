package module

import (
	"context"
	"fmt"
	"github.com/hyperledger-labs/yui-relayer/log"
	"math"
	"os"
	"strconv"
)

type Network string

const (
	Localnet Network = "localnet"
	Testnet  Network = "testnet"
	Mainnet  Network = "mainnet"
)

var localLorentzHF isForkSpec_HeightOrTimestamp = &ForkSpec_Height{Height: 1}

func init() {
	localLorentzHFTimestamp := os.Getenv("LOCAL_LORENTZ_HF_TIMESTAMP")
	if localLorentzHFTimestamp != "" {
		result, err := strconv.Atoi(localLorentzHFTimestamp)
		if err != nil {
			panic(err)
		}
		localLorentzHF = &ForkSpec_Timestamp{Timestamp: uint64(result)}
	}
}

func GetForkParameters(network Network) []*ForkSpec {
	switch network {
	case Localnet:
		return []*ForkSpec{
			// Pascal HF
			{
				// Must Set Milli timestamp
				HeightOrTimestamp:         &ForkSpec_Height{Height: 0},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			// Lorentz HF
			{
				// Must Set Milli timestamp
				HeightOrTimestamp:         localLorentzHF,
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	case Testnet:
		return []*ForkSpec{
			{
				// https://forum.bnbchain.org/t/bnb-chain-upgrades-testnet/934
				HeightOrTimestamp:         &ForkSpec_Height{Height: 48576786},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: math.MaxUint64},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	case Mainnet:
		return []*ForkSpec{
			{
				// https://bscscan.com/block/47618307
				// https://github.com/bnb-chain/bsc/releases/tag/v1.5.7
				HeightOrTimestamp:         &ForkSpec_Height{Height: 47618307},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: math.MaxUint64},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	}
	return nil
}

type BoundaryEpochs struct {
	PreviousForkSpec ForkSpec
	CurrentForkSpec  ForkSpec
	BoundaryHeight   uint64
	PrevLast         uint64
	CurrentFirst     uint64
	Intermediates    []uint64
}

type BoundaryHeight struct {
	Height          uint64
	CurrentForkSpec ForkSpec
}

func (b BoundaryHeight) GetBoundaryEpochs(prevForkSpec ForkSpec) (*BoundaryEpochs, error) {
	boundaryHeight := b.Height
	prevLast := boundaryHeight - (boundaryHeight % prevForkSpec.EpochLength)
	index := uint64(0)
	currentFirst := uint64(0)
	for {
		candidate := boundaryHeight + index
		if candidate%b.CurrentForkSpec.EpochLength == 0 {
			currentFirst = candidate
			break
		}
		index++
	}
	intermediates := make([]uint64, 0)
	// starts 0, 200, 400...epoch_length
	if prevLast == 0 {
		const defaultEpochLength = 200
		for mid := prevLast + defaultEpochLength; mid < prevForkSpec.EpochLength; mid += defaultEpochLength {
			intermediates = append(intermediates, mid)
		}
	}
	for mid := prevLast + prevForkSpec.EpochLength; mid < currentFirst; mid += prevForkSpec.EpochLength {
		intermediates = append(intermediates, mid)
	}

	return &BoundaryEpochs{
		PreviousForkSpec: prevForkSpec,
		CurrentForkSpec:  b.CurrentForkSpec,
		BoundaryHeight:   boundaryHeight,
		PrevLast:         prevLast,
		CurrentFirst:     currentFirst,
		Intermediates:    intermediates,
	}, nil
}

func (be BoundaryEpochs) CurrentEpochBlockNumber(number uint64) uint64 {
	if number >= be.CurrentFirst {
		return number - (number % be.CurrentForkSpec.EpochLength)
	}

	if len(be.Intermediates) > 0 {
		for i := len(be.Intermediates) - 1; i >= 0; i-- {
			if number >= be.Intermediates[i] {
				return be.Intermediates[i]
			}
		}
	}
	return number - (number % be.PreviousForkSpec.EpochLength)
}

func (be BoundaryEpochs) PreviousEpochBlockNumber(currentEpochBlockNumber uint64) uint64 {
	if currentEpochBlockNumber == 0 {
		return 0
	}
	if currentEpochBlockNumber <= be.PrevLast {
		return currentEpochBlockNumber - be.PreviousForkSpec.EpochLength
	}

	for i, mid := range be.Intermediates {
		if currentEpochBlockNumber == mid {
			if i == 0 {
				return be.PrevLast
			}
			return be.Intermediates[i-1]
		}
	}

	if currentEpochBlockNumber == be.CurrentFirst {
		if len(be.Intermediates) == 0 {
			return be.PrevLast
		}
		return be.Intermediates[len(be.Intermediates)-1]
	}

	return currentEpochBlockNumber - be.CurrentForkSpec.EpochLength
}

func FindTargetForkSpec(forkSpecs []*ForkSpec, height uint64, timestamp uint64) (*ForkSpec, *ForkSpec, error) {
	reversed := make([]*ForkSpec, len(forkSpecs))
	for i, spec := range forkSpecs {
		reversed[len(forkSpecs)-i-1] = spec
	}

	getPrev := func(current *ForkSpec, i int) *ForkSpec {
		if i == len(reversed)-1 {
			return current
		}
		return reversed[i+1]
	}

	for i, spec := range reversed {
		if x, ok := spec.GetHeightOrTimestamp().(*ForkSpec_Height); ok {
			if x.Height <= height {
				return spec, getPrev(spec, i), nil
			}
		} else {
			if spec.GetTimestamp() <= timestamp {
				return spec, getPrev(spec, i), nil
			}
		}
	}
	return nil, nil, fmt.Errorf("no fork spec found height=%d, timestmp=%d", height, timestamp)
}

var boundaryHeightCache = make(map[uint64]uint64)

func GetBoundaryHeight(headerFn getHeaderFn, currentHeight uint64, currentForkSpec ForkSpec) (*BoundaryHeight, error) {
	logger := log.GetLogger()
	boundaryHeight := uint64(0)
	if condition, ok := currentForkSpec.GetHeightOrTimestamp().(*ForkSpec_Height); ok {
		boundaryHeight = condition.Height
	} else {
		ts := currentForkSpec.GetTimestamp()
		if v, ok := boundaryHeightCache[ts]; ok {
			boundaryHeight = v
		} else {
			logger.Debug("seek fork height", "currentHeight", currentHeight, "ts", ts)
			for i := int64(currentHeight); i >= 0; i-- {
				h, err := headerFn(context.Background(), uint64(i))
				if err != nil {
					return nil, err
				}
				if MilliTimestamp(h) == ts {
					boundaryHeight = h.Number.Uint64()
					logger.Debug("seek fork height found", "currentHeight", currentHeight, "ts", ts, "boundaryHeight", boundaryHeight)
					boundaryHeightCache[ts] = boundaryHeight
					break
				} else if MilliTimestamp(h) < ts {
					boundaryHeight = h.Number.Uint64() + 1
					logger.Debug("seek fork height found", "currentHeight", currentHeight, "ts", ts, "boundaryHeight", boundaryHeight)
					boundaryHeightCache[ts] = boundaryHeight
					break
				}
			}
		}
	}
	return &BoundaryHeight{
		Height:          boundaryHeight,
		CurrentForkSpec: currentForkSpec,
	}, nil
}
