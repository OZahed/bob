/*
Package circuitbreaker provides a simple circuit breaker implementation.

State transfer should be statistically evaluated to avoid false positives and negatives.
*/

package circuitbreaker

import (
	"sync"
	"time"
)

type Bucket struct {
	requests int
	failures int
}

type RetryPolicy struct {
	Count int
	Wait  time.Duration
}

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

// TODO: struct padding could be better to reduce the foot print by almost half
type CircuitBreaker struct {
	lastBucketTime       time.Time
	buckets              []Bucket
	currentRate          float64
	changeBucketDuration time.Duration
	stateStepInterval    time.Duration
	threshold            float64
	currentRPS           uint32
	lastIndex            int
	windowInSeconds      int
	bucketPerSecond      int
	totalRequests        int
	totalFailures        int
	currentState         State
	mu                   sync.RWMutex
}

// NewCircuitBreaker creates a new CircuitBreaker with the given windowInSeconds, bucketPerSecond and breakigThreshold.
// The windowInSeconds is the total time in seconds that the CircuitBreaker will keep track of.
// The bucketPerSecond is the number of buckets that the windowInSeconds will be divided into.
// The breakigThreshold is the percentage of failures that will cause the CircuitBreaker to open.
// The StateHandler is the handler that will be used to evaluate the state of the CircuitBreaker.
func NewCircuitBreaker(windowInSeconds, bucketsPerSecond int,
	threshold float64, stateStepInterval time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		windowInSeconds:   windowInSeconds,
		bucketPerSecond:   bucketsPerSecond,
		threshold:         threshold,
		stateStepInterval: stateStepInterval,
		buckets:           make([]Bucket, windowInSeconds*bucketsPerSecond),
	}
}

func (cb *CircuitBreaker) getBucketIndex() int {
	if cb.lastBucketTime.IsZero() {
		cb.lastBucketTime = time.Now()
		cb.buckets[cb.lastIndex] = Bucket{}
	}

	if time.Since(cb.lastBucketTime) < cb.changeBucketDuration {
		return cb.lastIndex
	}

	outDatedBucket := cb.buckets[cb.lastIndex]

	// clean up the outdated values
	cb.totalRequests -= outDatedBucket.requests
	cb.totalFailures -= outDatedBucket.failures

	// reset the bucket and recalculating the last index and current rate
	cb.lastIndex = (cb.lastIndex + 1) % len(cb.buckets)
	cb.buckets[cb.lastIndex] = Bucket{}

	cb.lastBucketTime = time.Now()

	cb.updateStats()
	return cb.lastIndex
}

// MakeRequest registers a request and a failure in the current bucket.
// It then updates the stats and evaluates the state of the CircuitBreaker.
// If the CircuitBreaker is in the Open state, it will return an error.
//
// Client is responisble for handling the error and determining which errors should be counted as
// error for circuit breaker
// e.x:
//
//	err := cb.MakeRequest(&cb.RetryPolicy{Count: 3, Wailt: time.Second*3},func() error {
//		res, err := http.Get("http://example.com")
//		if err != nil {
//			return err
//		}
//
//
//		// check the status code and return an error if it is not 200
//		if !(res.StatusCode >= 200 && res.StatusCode < 400){
//			return errors.New("server returned non-200 status code")
//		}
//
//		// read response body
//		defer res.Body.Close()
//		body, err := ioutil.ReadAll(res.Body)
//
//		// if you want to ignore bad response value for circuit breaker, thats up to you
//		if err != nil {
//			return nil
//		}
//		return nil
//	})
func (cb *CircuitBreaker) MakeRequest(f func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	defer cb.StateEval()

	if !cb.Allow() {
		return ErrRequestDropped
	}

	idx := cb.getBucketIndex()

	cb.totalRequests++
	cb.buckets[idx].requests++

	err := f()
	if err != nil {
		cb.totalFailures++
		cb.buckets[idx].failures++
	}

	cb.updateStats()
	cb.StateEval()

	return err
}

func (cb *CircuitBreaker) updateStats() {
	cb.currentRate = float64(cb.totalFailures) / float64(cb.totalRequests)
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.currentState {
	case Closed:
		return cb.closedAllow()
	case Open:
		return false
	case HalfOpen:
		return cb.halfOpenAllow(cb.currentRPS)
	default:
		return false
	}
}

func (cb *CircuitBreaker) closedAllow() bool {
	return cb.currentRate < cb.threshold
}

func (cb *CircuitBreaker) halfOpenAllow(_ uint32) bool {
	return false
}

// Bring everuything here
func (cb *CircuitBreaker) StateEval() {
	panic("should be evaluated")
}
