package module

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/hyperledger-labs/yui-relayer/log"
	"os"
	"slices"
	"strconv"
)

type Network string

const (
	Localnet Network = "localnet"
	Testnet  Network = "testnet"
	Mainnet  Network = "mainnet"
)

var localLatestHF isForkSpec_HeightOrTimestamp = &ForkSpec_Height{Height: 1}

func init() {
	localLatestHFTimestamp := os.Getenv("LOCAL_LATEST_HF_TIMESTAMP")
	if localLatestHFTimestamp != "" {
		result, err := strconv.Atoi(localLatestHFTimestamp)
		if err != nil {
			panic(err)
		}
		localLatestHF = &ForkSpec_Timestamp{Timestamp: uint64(result)}
	}
}

const (
	indexPascalHF  = 0
	indexLorentzHF = 1
	indexMaxwellHF = 2
)

func getForkSpecParams() []*ForkSpec {
	return []*ForkSpec{
		// Pascal HF
		{
			AdditionalHeaderItemCount: 1,
			EpochLength:               200,
			MaxTurnLength:             9,
			GasLimitBoundDivider:      256,
			EnableHeaderMsec:          false,
		},
		// Lorentz HF
		{
			AdditionalHeaderItemCount: 1,
			EpochLength:               500,
			MaxTurnLength:             64,
			GasLimitBoundDivider:      1024,
			EnableHeaderMsec:          true,
		},
		// Maxwell HF
		{
			AdditionalHeaderItemCount: 1,
			EpochLength:               1000,
			MaxTurnLength:             64,
			GasLimitBoundDivider:      1024,
			EnableHeaderMsec:          true,
		},
	}
}

func GetForkParameters(network Network) []*ForkSpec {
	hardForks := getForkSpecParams()
	switch network {
	case Localnet:
		hardForks[indexPascalHF].HeightOrTimestamp = &ForkSpec_Height{Height: 0}
		hardForks[indexLorentzHF].HeightOrTimestamp = &ForkSpec_Height{Height: 1}
		hardForks[indexMaxwellHF].HeightOrTimestamp = localLatestHF
		return hardForks
	case Testnet:
		hardForks[indexPascalHF].HeightOrTimestamp = &ForkSpec_Height{Height: 48576786}
		hardForks[indexLorentzHF].HeightOrTimestamp = &ForkSpec_Height{Height: 49791365}
		//TODO Maxwell
		return hardForks
	case Mainnet:
		hardForks[indexPascalHF].HeightOrTimestamp = &ForkSpec_Height{Height: 47618307}
		// https://bscscan.com/block/48773576
		// https://github.com/bnb-chain/bsc/releases/tag/v1.5.10
		hardForks[indexLorentzHF].HeightOrTimestamp = &ForkSpec_Height{Height: 48773576}
		//TODO Maxwell
		return hardForks
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

func (b BoundaryHeight) GetBoundaryEpochs(prevForkSpecs []*ForkSpec) (*BoundaryEpochs, error) {
	if len(prevForkSpecs) == 0 {
		return nil, errors.New("EmptyPreviousForkSpecs")
	}
	prevForkSpec := prevForkSpecs[0]
	boundaryHeight := b.Height
	prevLast := boundaryHeight - (boundaryHeight % prevForkSpec.EpochLength)
	currentFirst := uint64(0)
	if boundaryHeight%b.CurrentForkSpec.EpochLength == 0 {
		currentFirst = boundaryHeight
	} else {
		currentFirst = boundaryHeight + (b.CurrentForkSpec.EpochLength - boundaryHeight%b.CurrentForkSpec.EpochLength)
	}

	intermediates := make([]uint64, 0)

	if prevLast == 0 {
		epochLengthList := uniqMap(prevForkSpecs, func(item *ForkSpec, index int) uint64 {
			return item.EpochLength
		})
		slices.Reverse(epochLengthList)
		for i := 0; i < len(epochLengthList)-1; i++ {
			start, end := epochLengthList[i], epochLengthList[i+1]
			value := start
			for value < end {
				intermediates = append(intermediates, value)
				value += start
			}
		}
	}

	mid := prevLast + prevForkSpec.EpochLength
	for mid < currentFirst {
		intermediates = append(intermediates, mid)
		mid += prevForkSpec.EpochLength
	}

	return &BoundaryEpochs{
		PreviousForkSpec: *prevForkSpec,
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

func FindTargetForkSpec(forkSpecs []*ForkSpec, height uint64, timestamp uint64) (*ForkSpec, []*ForkSpec, error) {
	reversed := make([]*ForkSpec, len(forkSpecs))
	for i, spec := range forkSpecs {
		reversed[len(forkSpecs)-i-1] = spec
	}

	getPreviousForkSpecs := func(current *ForkSpec, i int) []*ForkSpec {
		if i == len(reversed)-1 {
			return []*ForkSpec{current}
		}
		return reversed[i+1:]
	}

	for i, spec := range reversed {
		if x, ok := spec.GetHeightOrTimestamp().(*ForkSpec_Height); ok {
			if x.Height <= height {
				return spec, getPreviousForkSpecs(spec, i), nil
			}
		} else {
			if spec.GetTimestamp() <= timestamp {
				return spec, getPreviousForkSpecs(spec, i), nil
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
