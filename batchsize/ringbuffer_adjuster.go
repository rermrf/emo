package batchsize

import (
	"context"
	"sync"
	"time"

	"github.com/rermrf/emo/ringbuffer"
)

// RingBufferAdjuster 基于过去执行时间维护的环形缓冲区动态调整批次大小
// 当前执行时间超过历史平均值时减少批大小，低于平均值时增加批大小
type RingBufferAdjuster struct {
	mutex          *sync.RWMutex
	timeBuffer     *ringbuffer.TimeDurationRingBuffer
	batchSize      int           // 当前批大小
	minBatchSize   int           // 最小批次大小
	maxBatchSize   int           // 最大批次大小
	adjustStep     int           // 调整步长
	cooldownPeriod time.Duration // 调整后的冷却时间
	lastAdjustTime time.Time     // 上次调整时间
}

// NewRingBufferAdjuster 创建基于环形缓冲区的批大小调整器
// initialSize：初始批次大小
// minBatchSize：最小批次大小
// maxBatchSize：最大批次大小
// adjustStep：调整步长
// cooldownPeriod：调整后的冷却时间
// bufferSize：环形缓冲区大小
func NewRingBufferAdjuster(initialSize int, minBatchSize int, maxBatchSize int, adjustStep int, cooldownPeriod time.Duration, bufferSize int) *RingBufferAdjuster {
	if initialSize < minBatchSize {
		initialSize = minBatchSize
	} else if initialSize > maxBatchSize {
		initialSize = maxBatchSize
	}

	if bufferSize <= 0 {
		bufferSize = 128
	}

	timeBuffer, _ := ringbuffer.NewTimeDurationRingBuffer(bufferSize)

	return &RingBufferAdjuster{
		mutex:          &sync.RWMutex{},
		timeBuffer:     timeBuffer,
		batchSize:      initialSize,
		minBatchSize:   minBatchSize,
		maxBatchSize:   maxBatchSize,
		adjustStep:     adjustStep,
		cooldownPeriod: cooldownPeriod,
		lastAdjustTime: time.Time{},
	}
}

// Adjust 根据相应时间动态调整批次大小
// 1. 记录响应十斤啊到环形缓冲区
// 2. 如果当前时间比平均时间长，且不在冷却期，则减少批大小
// 3. 如果当前时间比平均时间短，且不在冷却期，则增加批大小
func (r *RingBufferAdjuster) Adjust(ctx context.Context, responseTime time.Duration) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// 记录当前时间到环形缓冲区
	r.timeBuffer.Push(responseTime)

	// 需要至少收集一轮才能开始调整
	if r.timeBuffer.Len() < r.timeBuffer.Cap() {
		return r.batchSize, nil
	}

	// 如果处于冷却期内，不调整批大小
	if !r.lastAdjustTime.IsZero() && time.Since(r.lastAdjustTime) < r.cooldownPeriod {
		return r.batchSize, nil
	}

	// 获取平均执行时间
	avgTime := r.timeBuffer.Avg()

	// 根据响应时间调整批大小
	if responseTime > avgTime {
		// 相应时间高于平均值，减小批大小
		if r.batchSize > r.minBatchSize {
			r.batchSize = max(r.batchSize-r.adjustStep, r.minBatchSize)
			r.lastAdjustTime = time.Now()
		}
	} else if responseTime < avgTime {
		// 响应时间低于平均值，增加批大小
		if r.batchSize < r.maxBatchSize {
			r.batchSize = min(r.batchSize+r.adjustStep, r.maxBatchSize)
			r.lastAdjustTime = time.Now()
		}
	}
	// 响应时间等于平均值，不调整
	return r.batchSize, nil
}
