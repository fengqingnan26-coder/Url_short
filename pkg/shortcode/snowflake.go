package shortcode

import (
	"fmt"

	"github.com/jekyulll/url_shortener/pkg/snowflake"
)

// SnowflakeShortCodeGenerator 基于雪花算法的短码生成器：
// 先生成全局唯一的雪花 ID，再编码为 Base62 字符串。
// 生成的短码天然不重复，无需依赖随机重试。
type SnowflakeShortCodeGenerator struct {
	node      *snowflake.Node
	minLength int // 短码保底长度，自然长度不足时左侧补 chars[0]
}

// NewSnowflakeShortCodeGenerator 创建雪花短码生成器。
// minLength <= 0 时不做补齐。
func NewSnowflakeShortCodeGenerator(node *snowflake.Node, minLength int) *SnowflakeShortCodeGenerator {
	return &SnowflakeShortCodeGenerator{
		node:      node,
		minLength: minLength,
	}
}

// GenerateShortCode 生成一个 Base62 短码。
func (g *SnowflakeShortCodeGenerator) GenerateShortCode() (string, error) {
	id, err := g.node.Generate()
	if err != nil {
		return "", fmt.Errorf("generate snowflake id: %w", err)
	}
	code := base62Encode(id)
	if g.minLength > len(code) {
		pad := make([]byte, g.minLength-len(code))
		for i := range pad {
			pad[i] = chars[0]
		}
		code = string(pad) + code
	}
	return code, nil
}

// base62Encode 将 uint64 编码为 Base62 字符串，字母表复用短码字符集 chars。
func base62Encode(id uint64) string {
	if id == 0 {
		return string(chars[0])
	}
	var buf [11]byte // uint64 最长 11 位 Base62
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = chars[id%62]
		id /= 62
	}
	return string(buf[i:])
}
