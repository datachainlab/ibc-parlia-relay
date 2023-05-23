package module

import "github.com/ethereum/go-ethereum/common"

type Ring [][]byte

func (r *Ring) IndexOf(element common.Address) int {
	for i, e := range *r {
		if common.BytesToAddress(e) == element {
			return i
		}
	}
	return -1
}

func (r *Ring) Range(start, size int) Ring {
	values := *r
	result := make([][]byte, 0, size)
	startAt := start % len(values)
	for {
		if len(result) == size {
			break
		}
		result = append(result, values[startAt])
		startAt = r.NextIndex(startAt)
	}
	return result
}

func (r *Ring) NextIndex(current int) int {
	next := current + 1
	if next >= len(*r) {
		return 0
	}
	return next
}

func (r *Ring) Get(index int) []byte {
	return (*r)[index%len(*r)]
}

func (r *Ring) Contains(target common.Address) bool {
	for _, e := range *r {
		if common.BytesToAddress(e) == target {
			return true
		}
	}
	return false
}
