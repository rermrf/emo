package batchsize

import (
	"context"
	"time"
)

type FixedStepAdjuster struct {
	batchSize         int           // 当前批次大小
	adjustStep        int           // 每次调整的步长
	minBatchSize      int           // 最小批次大小
	maxBatchSize      int           // 最大批次大小
	lastAdjustTime    time.Time     // 上次调整时间
	minAdjustInterval time.Duration // 两次调整的最小间隔

	// 响应时间阈值
	fastThreshold time.Duration // 响应时间低于此阈值时，批次大小增加
	slowThreshold time.Duration // 响应时间高于此阈值时，批次大小减少
}

func NewFixedStepAdjuster(initialSize int, adjustStep int, minBatchSize int, maxBatchSize int, minAdjustInterval time.Duration, fastThreshold time.Duration, slowThreshold time.Duration) *FixedStepAdjuster {
	if initialSize < minBatchSize {
		initialSize = minBatchSize
	}
	if initialSize > maxBatchSize {
		initialSize = maxBatchSize
	}
	return &FixedStepAdjuster{
		batchSize:         initialSize,
		adjustStep:        adjustStep,
		minBatchSize:      minBatchSize,
		maxBatchSize:      maxBatchSize,
		lastAdjustTime:    time.Time{},
		minAdjustInterval: minAdjustInterval,
		fastThreshold:     fastThreshold,
		slowThreshold:     slowThreshold,
	}
}

// Adjust 根据相应时间动态调整批次大小
func (f *FixedStepAdjuster) Adjust(ctx context.Context, responseTime time.Duration) (int, error) {
	// 检查是否允许调整（满足最小间隔要求）
	if !f.lastAdjustTime.IsZero() && time.Since(f.lastAdjustTime) < f.minAdjustInterval {
		return f.batchSize, nil
	}

	// 根据相应时间调整批次大小
	if responseTime < f.fastThreshold {
		// 响应快，可以增加批次大小
		if f.batchSize < f.maxBatchSize {
			f.batchSize = min(f.batchSize+f.adjustStep, f.maxBatchSize)
			f.lastAdjustTime = time.Now()
		}
	} else if responseTime > f.slowThreshold {
		// 响应慢，需要减小批次大小
		f.batchSize -= f.adjustStep
		if f.batchSize < f.minBatchSize {
			f.batchSize = f.minBatchSize
		}
	}
	// 响应时间在阈值之间，不调整
	return f.batchSize, nil
}
