package filter

import (
	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilter interface {
	Add(shortCode string)
	Exists(shortCode string) bool
}

type bloomFilterImpl struct {
	b *bloom.BloomFilter
}

func New(capacity uint, errorRate float64) BloomFilter {
	b := bloom.NewWithEstimates(capacity, errorRate)
	return &bloomFilterImpl{
		b: b,
	}
}

func (f *bloomFilterImpl) Add(shortCode string) {
	f.b.Add([]byte(shortCode))
}

func (f *bloomFilterImpl) Exists(shortCode string) bool {
	return f.b.Test([]byte(shortCode))
}
