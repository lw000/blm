package bloomf

import "github.com/willf/bitset"

const DEFAULT_SIZE = 2 << 24

var seeds = []uint{7, 11, 13, 31, 37, 61}

type SimpleHash struct {
	cap  uint
	seed uint
}

func (s SimpleHash) Hash(value string) uint {
	var result uint = 0
	for i := 0; i < len(value); i++ {
		result = result*s.seed + uint(value[i])
	}
	return (s.cap - 1) & result
}

type BloomFilter struct {
	b   *bitset.BitSet
	fns [6]SimpleHash
}

func New(size uint) *BloomFilter {
	bf := &BloomFilter{}
	bf.b = bitset.New(DEFAULT_SIZE)
	for i := 0; i < len(seeds); i++ {
		bf.fns[i] = SimpleHash{
			cap:  DEFAULT_SIZE,
			seed: seeds[i],
		}
	}
	return bf
}

func (bf *BloomFilter) Add(value string) {
	if value == "" {
		return
	}

	for _, fn := range bf.fns {
		bf.b.Set(fn.Hash(value))
	}
}

func (bf *BloomFilter) Contains(value string) bool {
	if value == "" {
		return false
	}

	ret := true
	for _, fn := range bf.fns {
		ret = bf.b.Test(fn.Hash(value))
	}
	return ret
}

func (bf *BloomFilter) Load(filename string) bool {
	if filename == "" {
		return false
	}

	return true
}
