package handlers

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// This is a temporary process-based limit implementation;
// In the future, we probably want to switch to a database-powered approach

// Mail request limit is 1 request every 6 hours
const mailLimitNumber = 6 * time.Hour

var mailLimit = rate.Every(mailLimitNumber / 1)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.
func init() {
	go cleanupVisitors()
}

func getVisitor(email string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[email]
	if !exists {
		limiter := rate.NewLimiter(mailLimit, 1)
		// Include the current time when creating a new visitor.
		visitors[email] = &visitor{limiter, time.Now()}
		return limiter
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Deletes visitor limit, noop if not existent
func deleteVisitor(email string) {
	mu.Lock()
	defer mu.Unlock()
	delete(visitors, email)
}

// Every 6 hours check the map for visitors that haven't been seen for
// more than 6 hours and delete the entries.
func cleanupVisitors() {
	for {
		time.Sleep(mailLimitNumber)

		mu.Lock()
		for email, v := range visitors {
			if time.Since(v.lastSeen) > mailLimitNumber {
				// Mail request limit is 1 request every 6 hours
				delete(visitors, email)
			}
		}
		mu.Unlock()
	}
}
