package token_bucket

import (
	"math"
	"sync"
	"time"
)

const (
	refillDuration = time.Second // default refill duration
	defaultTokensN = 1           // default weight for one requests or operation
)

// TokenBucket
//
//	implement the token bucket algorithm
//
//	Fields:
//
//	[refillRate]    number of tokens to be added in bucket per refill duration
//	[maxTokens]     maximum number of tokens in bucket
//	[currTokens]    current token number in bucket
//	[lastFillT]     time of the last refilling of the bucket
//	[refillT]       time for bucket refilling
//	[lock]          mutex for atomic operations
//
//	For Options:
//
//	[tokenN] weight for one request or operation. default: 1
//	[refillDur] bucket refill duration. default: 1 second
type TokenBucket struct {
	refillRate int
	maxTokens  int
	currTokens int
	lastFillT  time.Time
	refillT    time.Time
	lock       sync.Mutex

	tokenN    int
	refillDur time.Duration
}

// NewTokenBucket returns new TokenBucket entity instance
func NewTokenBucket(maxTokens, refillRate int, options ...Option) *TokenBucket {
	tb := &TokenBucket{
		refillRate: refillRate,
		maxTokens:  maxTokens,
		currTokens: maxTokens,
		lastFillT:  nowT(),

		tokenN:    defaultTokensN,
		refillDur: refillDuration,
	}

	for _, opt := range options {
		opt(tb)
	}

	tb.refillT = tb.nextT()

	return tb
}

// Option for TokenBucket entity
type Option func(*TokenBucket)

// SetRefillDuration set refill duration
func SetRefillDuration(dur time.Duration) Option {
	return func(tb *TokenBucket) {
		tb.refillDur = dur
	}
}

// SetTokenN set weight for one request or operation
func SetTokenN(n int) Option {
	return func(tb *TokenBucket) {
		tb.tokenN = n
	}
}

// nowT returns current time in UTC
func nowT() time.Time {
	return time.Now().UTC()
}

// nextT returns next filling time
func (tb *TokenBucket) nextT() time.Time {
	return tb.lastFillT.Add(tb.refillDur)
}

// refill fill the bucket if the 'refillDur' interval is reached
func (tb *TokenBucket) refill() {
	nowT := nowT()

	if tb.refillT.Unix() <= nowT.Unix() {

		filling := float64(tb.currTokens + tb.refillRate)
		max := float64(tb.maxTokens)

		tb.currTokens = int(math.Min(filling, max))

		tb.lastFillT = nowT
		tb.refillT = tb.nextT()
	}
}

// AllowN return 'true' if there are 'n' tokens in the bucket
func (tb *TokenBucket) AllowN(n int) bool {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	tb.refill()

	if tb.currTokens < n {
		return false
	}
	tb.currTokens -= n

	return true
}

// Allow returns 'true' if there are enough tokens in the bucket
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(tb.tokenN)
}
