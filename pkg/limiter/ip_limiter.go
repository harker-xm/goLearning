package limiter

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// IpLimiter implements LimiterIface based on client IP address.
// Buckets are created lazily on first request from each IP.
// All fields are protected by mu to ensure concurrent safety.
type IpLimiter struct {
	mu             sync.RWMutex
	limiterBuckets map[string]*ratelimit.Bucket
	fillInterval   time.Duration
	capacity       int64
	quantum        int64
}

// NewIpLimiter creates a new IpLimiter with no rules set yet.
// Call AddBuckets to configure the default token bucket parameters.
func NewIpLimiter() LimiterIface {
	return &IpLimiter{
		limiterBuckets: make(map[string]*ratelimit.Bucket),
	}
}

// Key returns the client IP address as the rate limit key.
func (l *IpLimiter) Key(c *gin.Context) string {
	return c.ClientIP()
}

// GetBucket retrieves the token bucket for the given IP key,
// creating one lazily if it does not yet exist.
func (l *IpLimiter) GetBucket(key string) (*ratelimit.Bucket, bool) {
	// Read fillInterval under lock to avoid data race with AddBuckets.
	l.mu.RLock()
	fillInterval := l.fillInterval
	bucket, ok := l.limiterBuckets[key]
	l.mu.RUnlock()

	if fillInterval == 0 {
		// No rule configured — do not limit.
		return nil, false
	}
	if ok {
		return bucket, true
	}

	// Bucket not found — acquire write lock and create it.
	l.mu.Lock()
	defer l.mu.Unlock()
	// Double-check after acquiring write lock.
	if bucket, ok = l.limiterBuckets[key]; ok {
		return bucket, true
	}
	bucket = ratelimit.NewBucketWithQuantum(l.fillInterval, l.capacity, l.quantum)
	l.limiterBuckets[key] = bucket
	return bucket, true
}

// AddBuckets stores the first rule's parameters as the default for all IP buckets.
// IpLimiter uses a single global rule applied to every IP address.
// Note: the Key field in LimiterBucketRule is ignored by IpLimiter.
func (l *IpLimiter) AddBuckets(rules ...LimiterBucketRule) LimiterIface {
	if len(rules) == 0 {
		return l
	}
	// IpLimiter only uses one global rule — take the first one.
	r := rules[0]
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fillInterval = r.FillInterval
	l.capacity = r.Capacity
	l.quantum = r.Quantum
	return l
}
