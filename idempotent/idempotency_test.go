package idempotent

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type IdempotencyServiceTest struct {
	Name       string
	NewService func() (IdempotencyService, func(), error)
}

func (ist IdempotencyServiceTest) RunTests(t *testing.T) {
	t.Helper()
	t.Run("TestExists", func(t *testing.T) {
		ist.TestExists(t)
	})
	t.Run("TestMExists", func(t *testing.T) {
		ist.TestMExists(t)
	})
}

func (ist IdempotencyServiceTest) TestExists(t *testing.T) {
	t.Parallel()
	service, cleanup, err := ist.NewService()
	require.NoError(t, err)
	t.Cleanup(cleanup)

	ctx := t.Context()

	// 第一次调用，键不应该存在
	key := "test-key-1"
	exists, err := service.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists, "Key should not exist")

	// 第二次调用，键应该存在
	exists, err = service.Exists(ctx, key)
	require.NoError(t, err)
	require.True(t, exists, "Key should exist after first call")

	// 使用不同的键，应该不存在
	newKey := "test-key-2"
	exists, err = service.Exists(ctx, newKey)
	require.NoError(t, err)
	require.False(t, exists, "New key should not exist")
}

func (ist IdempotencyServiceTest) TestMExists(t *testing.T) {
	t.Parallel()
	service, cleanup, err := ist.NewService()
	require.NoError(t, err)
	t.Cleanup(cleanup)

	ctx := t.Context()

	// 首次检查多个键
	keys := []string{"batch-key-1", "batch-key-2", "batch-key-3"}
	exists, err := service.MExists(ctx, keys...)
	require.NoError(t, err)
	require.Equal(t, len(keys), len(exists), "Results should have the same length")

	for i, res := range exists {
		require.False(t, res, "Key %s should not exist", keys[i])
	}

	// 再次检查同一批键
	exists, err = service.MExists(ctx, keys...)
	require.NoError(t, err)
	for i, res := range exists {
		require.True(t, res, "Key %s should exist", keys[i])
	}
}

func TestRedisImplementation(t *testing.T) {
	t.Parallel()
	t.Skip()
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	t.Cleanup(func() {
		client.Close()
	})

	// 检查 Redis 连接
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis server is not available, skipping test")
		return
	}

	// 先清空所有测试相关的key
	err := client.FlushDB(ctx).Err()
	if err != nil {
		t.Fatalf("Failed to flush DB: %v", err)
		return
	}

	// 创建测试套件
	redisTest := IdempotencyServiceTest{
		Name: "RedisIdempotencyService",
		NewService: func() (IdempotencyService, func(), error) {
			// 清理测试键前缀
			cleanup := func() {
				ctx := t.Context()
				iter := client.Scan(ctx, 0, "idempotency:*", 100).Iterator()
				for iter.Next(ctx) {
					client.Del(ctx, iter.Val())
				}
			}

			cleanup()

			// 创建服务实例
			service := NewRedisIdempotencyService(client, 10*time.Minute)
			return service, cleanup, nil
		},
	}

	redisTest.RunTests(t)
}

func TestBloomFilterImplementation(t *testing.T) {
	t.Parallel()
	t.Skip()
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	t.Cleanup(func() {
		client.Close()
	})

	// 检查 Redis 连接
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis server is not available, skipping test")
		return
	}

	// 检查是否支持布隆过滤器
	_, err := client.Do(ctx, "MODULE", "LIST").Result()
	if err != nil {
		t.Skip("Redis does not support Bloom filters, skipping test")
		return
	}

	// 设置布隆过滤器名称和参数
	filterName := "test-bloom-filter"
	capacity := uint64(10000)
	errorRate := 0.01

	// 尝试创建布隆过滤器名称和参数
	err = client.Del(ctx, filterName).Err()
	if err != nil {
		t.Skip("Failed to clean up previous Bloom filter, skipping test")
	}

	_, err = client.Do(ctx, "BF.RESERVE", filterName, errorRate, capacity).Result()
	if err != nil {
		t.Skip("Failed to create Bloom filter, skipping test")
		return
	}

	// 创建测试套件
	bloomTest := IdempotencyServiceTest{
		Name: "BloomIdempotencyService",
		NewService: func() (IdempotencyService, func(), error) {
			// 清理函数
			cleanup := func() {
				ctx := t.Context()
				client.Del(ctx, filterName)
				client.Do(ctx, "BF.RESERVE", filterName, errorRate, capacity)
			}

			// 先执行一次清理
			cleanup()

			// 创建服务实例
			service := NewBloomIdempotencyService(client, filterName, capacity, errorRate)
			return service, cleanup, nil
		},
	}

	bloomTest.RunTests(t)
}
