package randnum

import (
	"math/rand"

	"github.com/jekyulll/url_shortener/config"
)

const nums = "0123456789"

type RandNum struct {
	length int
}

func NewRandNum(cfg config.RandNumConfig) *RandNum {
	return &RandNum{
		length: cfg.Length,
	}
}

func (r *RandNum) Generate() string {
	result := make([]byte, r.length)
	length := len(nums)
	for i := range result {
		result[i] = nums[rand.Intn(length)]
	}
	return string(result)
}
