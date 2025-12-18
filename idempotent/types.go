package idempotent

import "context"

type IdempotencyService interface {
	// Exists 这里的Exist是包含添加语义的，返回true表示已经存在，返回false表示不存在，且将key添加到缓存中，下面的MExusts也是同理
	Exists(ctx context.Context, key string) (bool, error)
	MExists(ctx context.Context, keys ...string) ([]bool, error)
}
