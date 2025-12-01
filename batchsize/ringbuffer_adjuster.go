package batchsize

import (
	"github.com/rermrf/emo/ringbuffer"
	"sync"
)

type RingBufferAdjuster struct {
	mutex      *sync.Mutex
	timeBuffer *ringbuffer.TimeDurationRingBuffer
}
