package strategy

import (
	"sync/atomic"
	"time"
)

var _ Strategy = (*FixedIntervalRetryStrategy)(nil)

type FixedIntervalRetryStrategy struct {
	maxRetries int32         // 最大重试次数，如果是 0 或者负数则表示无限重试
	interval   time.Duration // 重试间隔时间
	retries    int32         // 当前重试次数
}

func NewFixedIntervalRetryStrategy(interval time.Duration, maxRetries int32) *FixedIntervalRetryStrategy {
	return &FixedIntervalRetryStrategy{
		maxRetries: maxRetries,
		interval:   interval,
	}
}

func (f *FixedIntervalRetryStrategy) NextWithRetries(retries int32) (time.Duration, bool) {
	return f.nextWithRetries(retries)
}

func (f *FixedIntervalRetryStrategy) nextWithRetries(retries int32) (time.Duration, bool) {
	if f.maxRetries <= 0 || retries <= f.maxRetries {
		return f.interval, true
	}
	return 0, false
}

func (f *FixedIntervalRetryStrategy) Next() (time.Duration, bool) {
	retries := atomic.AddInt32(&f.retries, 1)
	return f.nextWithRetries(retries)
}

func (f *FixedIntervalRetryStrategy) Report(_ error) Strategy {
	return f
}
