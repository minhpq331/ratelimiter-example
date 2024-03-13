package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type SlidingWindowRateLimiter struct {
	rate           int           // Maximum number of requests allowed in the windowDuration.
	windowDuration time.Duration // Duration of the sliding window.
	requests       map[int64]int // Map to hold request counts for each second within the window.
}

// NewSlidingWindowRateLimiter creates a new rate limiter instance.
func NewSlidingWindowRateLimiter(rate int, windowDuration time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		rate:           rate,
		windowDuration: windowDuration,
		requests:       make(map[int64]int),
	}
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Read the first line for number of request and requests per hour.
	var n, r int
	fmt.Scan(&n, &r)

	// Initialize the rate limiter with a rate of r requests per hour.
	rateLimiter := NewSlidingWindowRateLimiter(r, time.Hour)

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
