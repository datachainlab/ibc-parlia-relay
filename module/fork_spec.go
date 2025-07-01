package module

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"os"
	"slices"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/log"
)

type Network string

const (
	Localnet Network = "localnet"
	Testnet  Network = "testnet"
	Mainnet  Network = "mainnet"
)

var localLatestHF isForkSpec_HeightOrTimestamp = &ForkSpec_Height{Height: 2}

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
		hardForks[indexMaxwellHF].HeightOrTimestamp = &ForkSpec_Height{Height: 52552978}
		return hardForks
	case Mainnet:
		hardForks[indexPascalHF].HeightOrTimestamp = &ForkSpec_Height{Height: 47618307}
		// https://bscscan.com/block/48773576
		// https://github.com/bnb-chain/bsc/releases/tag/v1.5.10
		hardForks[indexLorentzHF].HeightOrTimestamp = &ForkSpec_Height{Height: 48773576}
		// https://github.com/bnb-chain/bsc/releases/tag/v1.5.16
		hardForks[indexMaxwellHF].HeightOrTimestamp = &ForkSpec_Timestamp{Timestamp: 1751250600 * 1000}
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

func GetBoundaryHeight(ctx context.Context, headerFn getHeaderFn, currentHeight uint64, currentForkSpec ForkSpec) (*BoundaryHeight, error) {
	var err error
	logger := log.GetLogger()
	boundaryHeight := uint64(0)
	if condition, ok := currentForkSpec.GetHeightOrTimestamp().(*ForkSpec_Height); ok {
		boundaryHeight = condition.Height
	} else {
		ts := currentForkSpec.GetTimestamp()
		if v, ok := boundaryHeightCache[ts]; ok {
			boundaryHeight = v
		} else {
			logger.DebugContext(ctx, "seek fork height", "currentHeight", currentHeight, "ts", ts)
			boundaryHeight, err = searchBoundaryHeight(ctx, currentHeight, ts, headerFn)
			if err != nil {
				return nil, err
			}
			boundaryHeightCache[ts] = boundaryHeight
		}
	}
	return &BoundaryHeight{
		Height:          boundaryHeight,
		CurrentForkSpec: currentForkSpec,
	}, nil
}

func searchBoundaryHeight(ctx context.Context, currentHeight uint64, targetTs uint64, headerFn getHeaderFn) (uint64, error) {
	// There are potentially many blocks between the boundary and the current
	// blocks. Also, finding the timestamp for a particular block is expensive
	// as it requires an RPC call to a node.
	//
	// Thus, this implementation aims to prune a large number of blocks from the
	// search space by estimating the distance between the boundary and the
	// current block (based on the average rate of block production) and jumping
	// directly to a candidate block at that distance. In case of a miss, all
	// blocks on one side of the candidate can be discarded, and a new attempt
	// can be made by re-estimating the new distance and jumping to a candidate
	// on the other side.
	//
	// Theoretical worst-case performance is O(N), but since the rate of block
	// production can be predicted with high accuracy, this implementation is
	// expected to be faster than binary search in practice.
	var (
		position       uint64        = currentHeight     // candidate block number currently under consideration
		low            uint64        = 0                 // inclusive lower bound of the current search range
		high           uint64        = currentHeight + 1 // exclusive upper bound of the current search range
		previousHeader *types.Header                     // header of the block seen in the previous iteration
	)

	// Loop invariant:
	//
	//     0 <= low <= position < high <= currentHeight + 1
	//     &&
	//     low <= result < high
	//
	// Bound function (decreases in each iteration, and is always >= 0):
	//
	//      high - low
	for low < high {
		currentHeader, err := headerFn(ctx, uint64(position))
		if err != nil {
			return 0, err
		}

		currentTs := MilliTimestamp(currentHeader)
		if currentTs == targetTs {
			return currentHeader.Number.Uint64(), nil
		}

		distance := estimateDistance(previousHeader, currentHeader, targetTs)

		if currentTs > targetTs {
			// Jump to a lower block.
			high = position

			// Since these are unsigned, position-distance might underflow.
			if low+distance > position {
				position = low
			} else {
				position = position - distance
			}
		} else {
			// Jump to a higher block.
			low = position + 1

			position = position + distance

			if position >= high {
				position = high - 1
			}
		}

		previousHeader = currentHeader
	}

	// If no block with an exact timestamp match was found, then we want the
	// earliest block that's _after_ the target timestamp.
	return low, nil
}

// estimateDistance returns the estimated number of blocks between the block indicated by currentHeader
// and the boundary block nearest to targetTs. It assumes that previousHeader either is nil, or refers to
// a different block than currentHeader.
func estimateDistance(previousHeader, currentHeader *types.Header, targetTs uint64) uint64 {
	if previousHeader == nil {
		return 1
	}

	var (
		timeDiffPrevCur   uint64 // milliseconds between the previous and current blocks
		timeDiffTargetCur uint64 // milliseconds between the current block and target timestamp
	)

	currentTs := MilliTimestamp(currentHeader)
	previousTs := MilliTimestamp(previousHeader)

	blockCountPrevCurBig := new(big.Int).Sub(previousHeader.Number, currentHeader.Number)
	blockCountPrevCurBig = blockCountPrevCurBig.Abs(blockCountPrevCurBig)
	blockCountPrevCur, _ := blockCountPrevCurBig.Float64()

	if currentTs > previousTs {
		timeDiffPrevCur = currentTs - previousTs
	} else {
		timeDiffPrevCur = previousTs - currentTs
	}

	if timeDiffPrevCur == 0 {
		// Found two different blocks with the same timestamp. The distance
		// should be at least 1 to avoid getting stuck in the current block.
		return 1
	}

	if currentTs > targetTs {
		timeDiffTargetCur = currentTs - targetTs
	} else {
		timeDiffTargetCur = targetTs - currentTs
	}

	avgBlocksPerMs := blockCountPrevCur / float64(timeDiffPrevCur)

	if avgBlocksPerMs > 0 {
		return uint64(math.Ceil(avgBlocksPerMs * float64(timeDiffTargetCur)))
	}

	// Blocks are being produced so slowly that the current block is still expected
	// to be the latest block at any future timestamp. Return 1 to avoid getting stuck
	// in the current block.
	return 1
}
