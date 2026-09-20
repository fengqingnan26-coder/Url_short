// Package snowflake 实现 Twitter 雪花算法，用于生成分布式唯一 ID。
//
// ID 的 64 位结构（从高位到低位）：
//
//	1 bit 符号位（恒为 0）
//	41 bit 毫秒级时间戳（相对自定义纪元 epoch）
//	 n bit 节点 ID（nodeBits，默认 10，最多 1024 个节点）
//	 m bit 同毫秒内序列号（stepBits，默认 12，每节点每毫秒 4096 个）
package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	// DefaultNodeBits 默认节点 ID 位数
	DefaultNodeBits uint8 = 10
	// DefaultStepBits 默认序列号位数
	DefaultStepBits uint8 = 12

	// maxBackwards 允许容忍的时钟回拨幅度，超过则直接报错
	maxBackwards = 5 * time.Millisecond
)

var (
	// ErrNodeOverload 节点 ID 超出 nodeBits 所能表示的范围
	ErrNodeOverload = errors.New("snowflake: node id exceeds the maximum allowed by node bits")
	// ErrInvalidBits 位数配置非法（总和不能超过 41，且不能为 0）
	ErrInvalidBits = errors.New("snowflake: node bits and step bits must be in (0, 41) and sum to at most 41")
	// ErrClockBackwards 时钟发生超出容忍范围的回拨
	ErrClockBackwards = errors.New("snowflake: clock moved backwards, refuse to generate id")
	// ErrInvalidEpoch 纪元时间晚于当前时间
	ErrInvalidEpoch = errors.New("snowflake: epoch must not be in the future")
)

// Node 表示一个雪花算法节点，每个工作实例应持有一个单例。
type Node struct {
	mu sync.Mutex

	epoch     int64 // 纪元（Unix 毫秒）
	node      int64 // 本节点 ID
	nodeBits  uint8
	stepBits  uint8
	nodeMax   int64 // 节点 ID 上限
	stepMax   int64 // 序列号上限
	timestamp int64 // 上次生成 ID 的时间戳（Unix 毫秒）
	step      int64 // 当前毫秒内已分配的序列号
}

// NewNode 创建一个雪花节点。
// nodeID 为本节点编号；epoch 为自定义纪元（建议设为项目上线时间附近，使 ID 尽可能短）；
// nodeBits、stepBits 传 0 时使用默认值 10 和 12。
func NewNode(nodeID int64, epoch time.Time, nodeBits, stepBits uint8) (*Node, error) {
	if nodeBits == 0 {
		nodeBits = DefaultNodeBits
	}
	if stepBits == 0 {
		stepBits = DefaultStepBits
	}
	if nodeBits == 0 || stepBits == 0 || nodeBits+stepBits > 41 {
		return nil, ErrInvalidBits
	}
	if epoch.After(time.Now()) {
		return nil, ErrInvalidEpoch
	}

	nodeMax := int64(-1) ^ (int64(-1) << nodeBits)
	if nodeID < 0 || nodeID > nodeMax {
		return nil, ErrNodeOverload
	}

	return &Node{
		epoch:    epoch.UnixMilli(),
		node:     nodeID,
		nodeBits: nodeBits,
		stepBits: stepBits,
		nodeMax:  nodeMax,
		stepMax:  int64(-1) ^ (int64(-1) << stepBits),
	}, nil
}

// NodeID 返回当前节点编号。
func (n *Node) NodeID() int64 {
	return n.node
}

// Generate 生成一个全局唯一的 uint64 ID。
// 同一毫秒内序列号耗尽时会自旋等待到下一毫秒；
// 发生小幅时钟回拨（<= maxBackwards）时等待时钟追平，大幅回拨返回 ErrClockBackwards。
func (n *Node) Generate() (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	// 时钟回拨处理
	if now < n.timestamp {
		diff := n.timestamp - now
		if diff <= maxBackwards.Milliseconds() {
			// 小幅度回拨：等待时钟追平上次时间
			time.Sleep(time.Duration(diff) * time.Millisecond)
			now = time.Now().UnixMilli()
		}
		if now < n.timestamp {
			return 0, ErrClockBackwards
		}
	}

	if now == n.timestamp {
		// 同一毫秒内递增序列号
		n.step++
		if n.step > n.stepMax {
			// 本毫秒序列号耗尽，自旋等待到下一毫秒
			for now <= n.timestamp {
				now = time.Now().UnixMilli()
			}
			n.step = 0
		}
	} else {
		// 进入新的一毫秒，序列号归零
		n.step = 0
	}

	n.timestamp = now

	shift := uint(n.nodeBits + n.stepBits)
	id := (now-n.epoch)<<shift |
		n.node<<n.stepBits |
		n.step
	return uint64(id), nil
}
