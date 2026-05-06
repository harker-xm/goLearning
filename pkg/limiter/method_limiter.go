package limiter

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// MethodLimiter implements LimiterIface based on request URI path.
type MethodLimiter struct {
	*Limiter
}

// NewMethodLimiter creates a new MethodLimiter.
func NewMethodLimiter() LimiterIface {
	return &MethodLimiter{
		Limiter: &Limiter{limiterBuckets: make(map[string]*ratelimit.Bucket)},
	}
}

// Key extracts the URI path (without query string) from the request context.
func (l *MethodLimiter) Key(c *gin.Context) string {
	uri := c.Request.RequestURI
	index := strings.Index(uri, "?")
	if index == -1 {
		return uri
	}
	return uri[:index]
}

// GetBucket retrieves the token bucket for the given key.
func (l *MethodLimiter) GetBucket(key string) (*ratelimit.Bucket, bool) {
	bucket, ok := l.limiterBuckets[key]
	return bucket, ok
}

// AddBuckets registers one or more rate limit rules and returns the limiter for chaining.
func (l *MethodLimiter) AddBuckets(rules ...LimiterBucketRule) LimiterIface {
	for _, rule := range rules {
		if _, ok := l.limiterBuckets[rule.Key]; !ok {
			l.limiterBuckets[rule.Key] = ratelimit.NewBucketWithQuantum(
				rule.FillInterval,
				rule.Capacity,
				rule.Quantum,
			)
		}
	}
	return l
}
