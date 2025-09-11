package handlers

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// This is a temporary process-based limit implementation;
// In the future, we probably want to switch to a database-powered approach

// Mail request limit is 1 per 6 hours
const mailLimitNumber = 6 * time.Hour
var mailLimit = rate.Every(mailLimitNumber)

// Test email limit is 1 per minute
const testMailLimitNumber = 1 * time.Minute
var testMailLimit = rate.Every(testMailLimitNumber)

// Holds rate limiters for normal emails and test emails
type visitor struct {
	emailLimiter     *rate.Limiter
	testEmailLimiter *rate.Limiter
	lastSeen    time.Time
}

// Map email handle to visitor pointers
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map
func init() {
	go cleanupVisitors()
}

// getVisitor returns a visitor queried by their email handle.
// If the visitor does not already exist, create a new entry for that visitor
func getVisitor(email string) *visitor {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[email]
	if !exists {
		emailLimiter := rate.NewLimiter(mailLimit, 1)
		testEmailLimiter := rate.NewLimiter(testMailLimit, 1)

		// Include the current time when creating a new visitor.
		visitors[email] = &visitor{emailLimiter, testEmailLimiter, time.Now()}
		return visitors[email]
	}

	// Update the last seen time for the visitor
	v.lastSeen = time.Now()
	return v
}

// Deletes visitor limit, noop if not existent
func deleteVisitor(email string) {
	mu.Lock()
	defer mu.Unlock()
	delete(visitors, email)
}

// Every 6 hours, check the map for visitors that haven't been seen for
// more than 6 hours and delete the entries
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
