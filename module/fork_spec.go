package module

type Network string

const (
	Localnet Network = "localnet"
	Testnet  Network = "testnet"
	Mainnet  Network = "mainnet"
)

// After Pascal HF only
func GetForkParameters(network Network) []*ForkSpec {
	switch network {
	case Localnet:
		return []*ForkSpec{
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 0},
				AdditionalHeaderItemCount: 21,
			},
		}
	case Testnet:
		return []*ForkSpec{
			{
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 1740452880},
				AdditionalHeaderItemCount: 21,
			},
		}
	case Mainnet:
		return []*ForkSpec{
			{
				//TODO
				HeightOrTimestamp:         &ForkSpec_Timestamp{Timestamp: 0},
				AdditionalHeaderItemCount: 21,
			},
		}
	}
	return nil
}
