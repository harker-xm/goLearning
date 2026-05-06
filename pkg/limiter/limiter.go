package limiter

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// LimiterIface is the interface that all rate limiters must implement.
type LimiterIface interface {
	Key(c *gin.Context) string
	GetBucket(key string) (*ratelimit.Bucket, bool)
	AddBuckets(rules ...LimiterBucketRule) LimiterIface
}

// Limiter holds the mapping from bucket name to token bucket.
type Limiter struct {
	limiterBuckets map[string]*ratelimit.Bucket
}

// LimiterBucketRule defines the configuration for a single rate limit bucket.
type LimiterBucketRule struct {
	Key          string
	FillInterval time.Duration
	Capacity     int64
	Quantum      int64
}
