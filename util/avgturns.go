package util

import (
	"math"
	"sync"
	"time"
)

const buffSize = 3

type AvgTurns struct {
	count             uint
	lastCompleteTurns int
	lastCalled        time.Time
	turns             [buffSize]int
	durations         [buffSize]time.Duration
	mutex             sync.Mutex
}

func NewAvgTurns() *AvgTurns {
	return &AvgTurns{
		count:             0,
		lastCompleteTurns: 0,
		lastCalled:        time.Now(),
		turns:             [buffSize]int{},
		durations:         [buffSize]time.Duration{},
		mutex:             sync.Mutex{},
	}
}

func (avg *AvgTurns) TurnsPerSec(completedTurns int) int {
	avg.mutex.Lock()
	avg.turns[avg.count%buffSize] = completedTurns - avg.lastCompleteTurns
	avg.durations[avg.count%buffSize] = time.Since(avg.lastCalled)
	avg.lastCalled = time.Now()
	avg.lastCompleteTurns = completedTurns
	avg.count++
	sumTurns := 0
	sumDurations := time.Duration(0)
	for i := 0; i < buffSize; i++ {
		sumTurns += avg.turns[i]
		sumDurations += avg.durations[i]
	}
	avg.mutex.Unlock()
	if sumDurations.Seconds() <= 0 {
		return 0
	}
	return int(math.Round(float64(sumTurns) / sumDurations.Seconds()))
}
