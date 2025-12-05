package retry

import (
	"fmt"
	"time"

	"github.com/rermrf/emo/retry/strategy"
)

type FixedIntervalConfig struct {
	MaxRetries int32         `json:"maxRetries" yaml:"maxRetries"`
	Interval   time.Duration `json:"interval" yaml:"interval"`
}

type ExponentialBackoffConfig struct {
	// 初始重试间隔 单位ms
	InitialInterval time.Duration `json:"initialInterval" yaml:"initialInterval"`
	// 最大重试间隔 单位ms
	MaxInterval time.Duration `json:"maxInterval" yaml:"maxInterval"`
	// 最大重试次数
	MaxRetries int32 `json:"maxRetries" yaml:"maxRetries"`
}

type Config struct {
	Type               string                    `json:"type" yaml:"type"`
	FixedInterval      *FixedIntervalConfig      `json:"fixedInterval" yaml:"fixedInterval"`
	ExponentialBackoff *ExponentialBackoffConfig `json:"exponentialBackoff" yaml:"exponentialBackoff"`
}

func NewRetry(cfg Config) (strategy.Strategy, error) {
	// 根据 config 中的字段来检测
	switch cfg.Type {
	case "fixed":
		return strategy.NewFixedIntervalRetryStrategy(cfg.FixedInterval.Interval, cfg.FixedInterval.MaxRetries), nil
	case "exponential":
		return strategy.NewExponentialBackoffRetryStrategy(cfg.ExponentialBackoff.InitialInterval, cfg.ExponentialBackoff.MaxInterval, cfg.ExponentialBackoff.MaxRetries), nil
	default:
		return nil, fmt.Errorf("未知重试类型：%s", cfg.Type)
	}
}
