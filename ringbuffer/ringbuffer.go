package ringbuffer

import (
	"errors"
	"sync"
	"time"
)

// ErrInvalidCapacity 当创建环形缓冲区的容量大小等于0时返回
var ErrInvalidCapacity = errors.New("环形缓冲区容量必须大于0")

type TimeDurationRingBuffer struct {
	mu       sync.Mutex
	buffer   []time.Duration // 环形存储
	capacity int             // 固定容量
	index    int             // 下一个写入位置
	count    int             // 当前有效数据个数
	sum      time.Duration   // buffer 中元素之和，便于O(1)求平均
}

// NewTimeDurationRingBuffer 创建一个容量为 capacity 的环形缓冲区
func NewTimeDurationRingBuffer(capacity int) (*TimeDurationRingBuffer, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &TimeDurationRingBuffer{
		buffer:   make([]time.Duration, capacity),
		capacity: capacity,
	}, nil
}

// Push 向环形缓冲区写入一个元素
// 当缓冲区已满时会覆盖最老的数据，同时维护 sum
func (r *TimeDurationRingBuffer) Push(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == r.capacity {
		// 缓冲区已满，需要减去即将被覆盖的值
		r.sum -= r.buffer[r.index]
	} else {
		r.count++
	}
	r.buffer[r.index] = d
	r.sum += d
	r.index = (r.index + 1) % r.capacity
}

// Avg 返回环形缓冲区的平均值
// 当尚无样本时返回 0
func (r *TimeDurationRingBuffer) Avg() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == 0 {
		return 0
	}
	return r.sum / time.Duration(r.count)
}

// Len 返回环形缓冲区的有效数据个数
func (r *TimeDurationRingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Cap 获取环形缓冲区的容量
func (r *TimeDurationRingBuffer) Cap() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.capacity
}

// Clear 清空环形缓冲区
func (r *TimeDurationRingBuffer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.buffer {
		r.buffer[i] = 0
	}
	r.index = 0
	r.count = 0
	r.sum = 0
}
