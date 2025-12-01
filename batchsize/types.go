package batchsize

import (
	"context"
	"time"
)

// Adjuster 根据相应时间动态调整批处理大小
type Adjuster interface {
	// Adjust 根据上次操作的相应时间计算下一批次的大小
	Adjust(ctx context.Context, responseTime time.Duration) (int, error)
}
