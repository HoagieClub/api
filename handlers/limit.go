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

// Test email limit is once per minute
const testMailLimitNumber = 1 * time.Minute

var mailLimit = rate.Every(mailLimitNumber / 1)
var testMailLimit = rate.Every(testMailLimitNumber / 1)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
// Also contains testLimiter for limiting rate of test emails
// this visitor is allowed to send
type visitor struct {
	limiter     *rate.Limiter
	testLimiter *rate.Limiter
	lastSeen    time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.
func init() {
	go cleanupVisitors()
}

// getVisitor returns visitor queried by their email handle
//
//	If the visitor does not already exist, a new entry for that visitor is created
func getVisitor(email string) *visitor {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[email]
	if !exists {
		limiter := rate.NewLimiter(mailLimit, 1)
		testLimiter := rate.NewLimiter(testMailLimit, 1)
		// Include the current time when creating a new visitor.
		visitors[email] = &visitor{limiter, testLimiter, time.Now()}
		return visitors[email]
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v
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
