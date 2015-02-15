package breaker

import (
	"sync/atomic"
	"time"
)

type sample struct {
	concurrentActions int64
	totalActions      int64
	totalFailures     int64
	performances      []int64
}

// TODO newSample(sync.mutex)

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
	// TODO lock
	s.performances = append(s.performances, actionEnd-actionStart)
	// TODO unlock
}
