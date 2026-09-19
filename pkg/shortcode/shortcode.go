package shortcode

import "math/rand"

type RandomShortCodeGeneratorImpl struct {
	length int
}

func NewShortCodeGeneratorImpl(length int) *RandomShortCodeGeneratorImpl {
	return &RandomShortCodeGeneratorImpl{
		length: length,
	}
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 使用随机数生成的短码
func (s *RandomShortCodeGeneratorImpl) GenerateShortCode() string {
	length := len(chars)
	result := make([]byte, s.length)
	for i := 0; i < s.length; i++ {
		result[i] = chars[rand.Intn(length)]
	}
	return string(result)
}

// 解决重复长地址转换攻击
// TODO 使用布隆过滤器存储长地址，每次转换前先判断下长地址是否已经转换过。
