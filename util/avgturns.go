package util

import (
	"math"
	"time"
)

const BUF_SIZE = 3

type AvgTurns struct {
	count             uint
	lastCompleteTurns int
	lastCalled        time.Time
	bufTurns          [BUF_SIZE]int
	bufDurations      [BUF_SIZE]time.Duration
}

func NewAvgTurns() *AvgTurns {
	return &AvgTurns{
		count:             0,
		lastCompleteTurns: 0,
		lastCalled:        time.Now(),
		bufTurns:          [BUF_SIZE]int {},
		bufDurations:      [BUF_SIZE]time.Duration {},
	}
}

func (avg *AvgTurns) Get(completedTurns int) int {
	avg.bufTurns[avg.count%BUF_SIZE] = completedTurns - avg.lastCompleteTurns
	avg.bufDurations[avg.count%BUF_SIZE] = time.Since(avg.lastCalled)
	avg.lastCalled = time.Now()
	avg.lastCompleteTurns = completedTurns
	avg.count++
	sumTurns := 0
	for _, turns := range avg.bufTurns {
		sumTurns += turns
	}
	sumDurations := time.Duration(0)
	for _, durations := range avg.bufDurations {
		sumDurations += durations
	}
	return int(sumTurns) / int(math.Round(math.Max(sumDurations.Seconds(), 1)))
}
