# ratelimiter-example

Example of implementing a ratelimiter

## The problem

> Implement a simple ratelimiter in API Gateway to ensure resistance to Brute Force attacks and DDOS attacks.

## Coding challenge

### Prerequisites

- Golang 1.22

### Solution 1: The sliding window algorithm (with counter per second)

I choose grouping by second as the unit of time, and store the counter per second to ensure the minimum precision of the ratelimiter is one second.

**Pros:**

- Very accurate, the minimum precision of the ratelimiter is one second.
- This algorithm is simple and easy to implement both in the API Gateway itself or in the centralize store like Redis (using SortedSet)

**Cons:**

- Memory usage grows linearly with window duration (but can be controlled). We can also reduce memory usage and consumed time and sacrifice accuracy with this algorithm by increasing the grouping unit to minute.
- Not effective because we need to loop through all the counter to remove the expired ones and sum them to find the number of requests in the current window.

```golang
type SlidingWindowRateLimiter struct {
	rate           int           // Maximum number of requests allowed in the windowDuration.
	windowDuration time.Duration // Duration of the sliding window.
	requests       map[int64]int // Map to hold request counts for each second within the window.
}

// AllowRequest determines whether a new request at the current time should be allowed.
func (rl *SlidingWindowRateLimiter) AllowRequest(requestTime time.Time) bool {
	// Round the request time down to the nearest second to group requests by second.
	requestTimeSecond := requestTime.Truncate(time.Second).Unix()

	// Clean up old requests that are outside the current window and count the number of requests in the current window.
	startOfWindow := requestTime.Add(-rl.windowDuration).Unix()
	currentCount := 0

	for timestamp, count := range rl.requests {
		if timestamp < startOfWindow {
			delete(rl.requests, timestamp)
		} else {
			currentCount += count
		}
	}

	// If the current window is not full, allow the request and update the counter.
	if currentCount < rl.rate {
		rl.requests[requestTimeSecond]++
		return true
	}

	// Otherwise, deny the request.
	return false
}
```

To run this solution over sample test case, run the following command:

```bash
cat testcase-sample.txt | go run sliding-window-counter.go
```

### Solution 2: The leaky bucket algorithm

Since the sliding window algorithm is accurate but not very effective, I think in the real world, we need to consider other algorithms to achieve a better ratelimiter. The leaky bucket is one of them.

**Pros:**

- Very low memory usage
- This algorithm is simple and easy to implement both in the API Gateway itself or in the centralize store like Redis

**Cons:**

- Not very accurate, the ratelimiter will let some requests at the end of the time window pass through even if the maximum number of requests allowed in the windowDuration has been reached.

```golang
type LeakyBucketRateLimiter struct {
	capacity       float64       // The maximum capacity of the bucket.
	windowDuration time.Duration // The duration of the sliding window.
	lastUpdate     time.Time     // The last time the bucket was updated.
	current        float64       // The current amount of requests in the bucket.
}

// NewLeakyBucketRateLimiter creates a new rate limiter instance.
func NewLeakyBucketRateLimiter(rate int, windowDuration time.Duration) *LeakyBucketRateLimiter {
	return &LeakyBucketRateLimiter{
		capacity:       float64(rate),
		windowDuration: windowDuration,
		current:        0,
	}
}

// AllowRequest determines whether a new request should be allowed.
func (lb *LeakyBucketRateLimiter) AllowRequest(requestTime time.Time) bool {
	// Calculate time elapsed since the last request.
	elapsed := requestTime.Sub(lb.lastUpdate).Seconds() / lb.windowDuration.Seconds()

	// Leak the bucket based on the elapsed time.
	lb.current -= elapsed * float64(lb.capacity)
	if lb.current < 0 {
		lb.current = 0
	}
	lb.lastUpdate = requestTime

	// If the bucket is not full, allow the request and update the bucket current amount.
	if math.Ceil(lb.current) < lb.capacity {
		lb.current++
		return true
	}

	// Otherwise, the request is denied.
	return false
}
```

To run this solution over sample test case, run the following command:

```bash
cat testcase-sample.txt | go run leaky-bucket.go
```

