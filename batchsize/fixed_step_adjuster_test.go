package batchsize

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewFixedStepAdjuster(t *testing.T) {
	testCases := []struct {
		name              string
		initialSize       int
		minBatchSize      int
		maxBatchSize      int
		adjustStep        int
		minAdjustInterval time.Duration
		fastThreshold     time.Duration
		slowThreshold     time.Duration
		expectSize        int
	}{
		{
			name:              "正常初始化",
			initialSize:       50,
			minBatchSize:      10,
			maxBatchSize:      100,
			adjustStep:        5,
			minAdjustInterval: time.Second * 10,
			fastThreshold:     time.Millisecond * 150,
			slowThreshold:     time.Millisecond * 300,
			expectSize:        50,
		},
		{
			name:              "初始值小于最小值",
			initialSize:       5,
			minBatchSize:      10,
			maxBatchSize:      100,
			adjustStep:        5,
			minAdjustInterval: time.Second * 10,
			fastThreshold:     time.Millisecond * 150,
			slowThreshold:     time.Millisecond * 300,
			expectSize:        10,
		},
		{
			name:              "初始值大于最小值",
			initialSize:       500,
			minBatchSize:      10,
			maxBatchSize:      100,
			adjustStep:        5,
			minAdjustInterval: time.Second * 10,
			fastThreshold:     time.Millisecond * 150,
			slowThreshold:     time.Millisecond * 300,
			expectSize:        100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adjuster := NewFixedStepAdjuster(tc.initialSize, tc.adjustStep, tc.minBatchSize, tc.maxBatchSize, tc.minAdjustInterval, tc.fastThreshold, tc.slowThreshold)

			firstSize, err := adjuster.Adjust(t.Context(), 175*time.Millisecond)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectSize, firstSize)
		})
	}
}

func TestAdjustBatchSize(t *testing.T) {
	t.Run("响应时间影响批次大小", func(t *testing.T) {
		t.Parallel()
		// 创建一个无间隔限制的调整器
		adjuster := NewFixedStepAdjuster(50, 10, 10, 100, 0, time.Millisecond*150, time.Millisecond*200)

		// 1. 响应时间在中间范围 - 保持不变
		size, err := adjuster.Adjust(t.Context(), 175*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 50, size, "中间响应时间不应改变批次大小")

		// 2. 快速响应 - 增加批次大小
		size, err = adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 60, size, "快速响应应增加批次大小")

		// 3. 再次快速响应 - 继续增加批次大小
		size, err = adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 70, size, "连续快速响应应增加批次大小")

		// 4. 慢速响应 - 减少批次大小
		size, err = adjuster.Adjust(t.Context(), 300*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 60, size, "慢速响应应减少批次大小")
	})

	t.Run("批次大小有边界限制", func(t *testing.T) {
		t.Parallel()

		adjuster := NewFixedStepAdjuster(90, 10, 10, 100, 0, time.Millisecond*150, time.Millisecond*200)

		// 1. 接近最大值时快速响应
		size, err := adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 100, size, "应增加但不超过最大值")

		// 2. 已达最大值时快速响应
		size, err = adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 100, size, "应保持不变")

		// 重新创建一个接近最小值的调整器
		adjuster = NewFixedStepAdjuster(30, 10, 20, 100, 0, time.Millisecond*150, time.Millisecond*200)

		// 3. 接近最小值时慢速响应
		size, err = adjuster.Adjust(t.Context(), 300*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 20, size, "应减少但不超过最小值")

		// 4. 已达最小值时慢速响应
		size, err = adjuster.Adjust(t.Context(), 300*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 20, size, "应保持不变")
	})

	t.Run("调整间隔限制", func(t *testing.T) {
		t.Parallel()

		// 创建一个有间隔限制的调整器
		adjuster := NewFixedStepAdjuster(50, 10, 10, 100, time.Millisecond*100, time.Millisecond*150, time.Millisecond*200)

		// 1. 首次调整应正常调整
		size, err := adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 60, size, "首次调用应正常调整")

		// 2. 紧接着调整不应调整
		size, err = adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 60, size, "间隔内不应调整")

		// 3. 等待间隔后调用应正常调整
		time.Sleep(time.Millisecond * 150)
		size, err = adjuster.Adjust(t.Context(), 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 70, size, "等待足够间隔后应可调整")
	})
}

func TestContinuousAdjustment(t *testing.T) {
	t.Run("连续调整行为", func(t *testing.T) {
		// 初始化一个间隔设为0的调整器
		adjuster := NewFixedStepAdjuster(50, 10, 10, 100, 0, time.Millisecond*150, time.Millisecond*200)

		// 验证初始大小
		initialSize, _ := adjuster.Adjust(t.Context(), 175*time.Millisecond)
		assert.Equal(t, 50, initialSize)

		// 连续增长直到上限
		sizes := []int{60, 70, 80, 90, 100}
		for _, size := range sizes {
			expectedSize, _ := adjuster.Adjust(t.Context(), 100*time.Millisecond)
			assert.Equal(t, size, expectedSize)
		}

		// 验证连续减少直到下限
		sizes = []int{90, 80, 70, 60, 50, 40, 30, 20, 10}
		for _, size := range sizes {
			expectedSize, _ := adjuster.Adjust(t.Context(), 300*time.Millisecond)
			assert.Equal(t, size, expectedSize)
		}
	})
}
