package breaker

import (
	"sync"
	"sync/atomic"
	"time"
)

type sampleConfig struct {
	mutex *sync.Mutex
}

type sample struct {
	sampleConfig

	concurrentActions int64
	totalActions      int64
	totalFailures     int64
	performances      []int64
}

func newSample(c sampleConfig) *sample {
	return &sample{sampleConfig: c}
}

func (s *sample) actionStart() int64 {
	s.totalActions = atomic.AddInt64(&s.totalActions, 1)
	s.concurrentActions = atomic.AddInt64(&s.concurrentActions, 1)

	return time.Now().UnixNano()
}

func (s *sample) actionSuccess() int64 {
	s.concurrentActions = atomic.AddInt64(&s.concurrentActions, -1)

	return time.Now().UnixNano()
}

func (s *sample) actionFailure() int64 {
	s.totalFailures = atomic.AddInt64(&s.totalFailures, 1)
	s.concurrentActions = atomic.AddInt64(&s.concurrentActions, -1)

	return time.Now().UnixNano()
}

func (s *sample) actionPerformance(actionStart, actionEnd int64) {
	s.mutex.Lock()
	s.performances = append(s.performances, actionEnd-actionStart)
	s.mutex.Unlock()
}
