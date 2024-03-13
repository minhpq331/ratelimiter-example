package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"time"
)

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

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Read the first line for number of request and requests per hour.
	var n, r int
	fmt.Scan(&n, &r)

	// Initialize the rate limiter with a rate of r requests per hour.
	rateLimiter := NewLeakyBucketRateLimiter(r, time.Hour)

	for i := 0; i < n; i++ {
		if !scanner.Scan() {
			fmt.Println("Error reading time input")
			return
		}
		timestampStr := scanner.Text()
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			fmt.Printf("Error parsing time: %v\n", err)
			continue
		}

		allowed := rateLimiter.AllowRequest(timestamp)
		if allowed {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}
	}
}
