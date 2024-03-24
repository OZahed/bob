package circuitbreaker

import (
	"time"
)

const (
	floatConfidanceDist = 0.01
)

type halfOpenStage uint8

const (
	First halfOpenStage = iota
	Second
	Third
	Final
)

type stateHadler struct {
	lastRequestTime      time.Time
	curretState          State
	currentCount         int
	reqPerInterval       float64
	secondsCount         int
	halfOpenCurrentStage int
	halfOpenStage        halfOpenStage
	avgTimeStep          time.Duration
	currentFlyingTests   int
	halfOpenTesting      bool
}

func (c *stateHadler) Name() State {
	return c.curretState
}

func (s *stateHadler) Allow(currentRate, thr float64) bool {
	if time.Since(s.lastRequestTime) > s.avgTimeStep {
		s.secondsCount += 1
		s.reqPerInterval = float64(s.currentCount) / float64(s.secondsCount)

		s.currentCount = 0
	}

	s.lastRequestTime = time.Now()

	switch s.curretState {
	case Closed:
		return (currentRate - thr) < floatConfidanceDist
	case Open:
		return false
	case HalfOpen:
		return s.halfOpenAllow()
	default:
		return false
	}
}

func NewStateHandler(avgTimeStep time.Duration) *stateHadler {
	if avgTimeStep < time.Second {
		avgTimeStep = time.Second
	}

	return &stateHadler{avgTimeStep: avgTimeStep}
}

func (s *stateHadler) StateEval(currentState State) {
	// implement later
}

func (s *stateHadler) halfOpenAllow() bool {
	expectedValue := int(s.reqPerInterval) + 1

	if expectedValue < 10 {
		return s.halfOpenTesting && s.currentFlyingTests > 1
	}

	var percentile int
	switch s.halfOpenStage {
	case First:
		percentile = (expectedValue * 10) / 10
	case Second:
		percentile = (expectedValue * 30) / 10
	case Third:
		percentile = (expectedValue * 50) / 10
	default:
		percentile = expectedValue
	}

	return s.currentFlyingTests <= percentile
}

// for half open we are going to only call one request if it made it, we will increase it to 10%
// of what we were expecting per second on average,keep it there for gradual step(a config value)
// then 30% then 50% then 70% and after that we should go to Closed state
