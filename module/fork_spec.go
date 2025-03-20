package module

import (
	"os"
	"strconv"
)

type Network string

const (
	Localnet Network = "localnet"
	Testnet  Network = "testnet"
	Mainnet  Network = "mainnet"
)

func GetForkParameters(network Network) []*ForkSpec {

	switch network {
	case Localnet:

		localLorentzHFTimestamp := os.Getenv("LOCAL_LORENTZ_HF_TIMESTAMP")
		localLorentzHFTimestampInt := uint64(1)
		if localLorentzHFTimestamp != "" {
			result, err := strconv.Atoi(localLorentzHFTimestamp)
			if err != nil {
				panic(err)
			}
			localLorentzHFTimestampInt = uint64(result)
		}
		return []*ForkSpec{
			// Pascal HF
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 0},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			// Lorentz HF
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: localLorentzHFTimestampInt},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	case Testnet:
		return []*ForkSpec{
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 0},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 1},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	case Mainnet:
		return []*ForkSpec{
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 0},
				AdditionalHeaderItemCount: 1,
				EpochLength:               200,
			},
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 1},
				AdditionalHeaderItemCount: 1,
				EpochLength:               500,
			},
		}
	}
	return nil
}
