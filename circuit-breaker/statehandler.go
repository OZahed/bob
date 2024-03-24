package circuitbreaker

import "time"

const avgTimeStep = time.Second

type stateHadler struct {
	curretState     string
	currentCount    int
	lastRequestTime time.Time
}

func (c *stateHadler) Name() string {
	return c.curretState
}

// implement later
func (c *stateHadler) Allow(currentRate, thr float64) bool {
	if time.Since(c.lastRequestTime) > avgTimeStep {
		c.lastRequestTime = time.Now()
		c.currentCount = 0
	}

	return true
}

func NewStateHandler() *stateHadler {
	return &stateHadler{}
}

func (c *stateHadler) MakeRequest(f func() (interface{}, error)) (interface{}, error) {
	return f()
}

func (c *stateHadler) StateEval(currentState State) {
	// implement later
}

// for half open we are going to only call one request if it made it, we will increase it to 10%
// of what we were expecting per second on average,keep it there for gradual step(a config value)
// then 30% then 50% then 70% and after that we should go to Closed state
