package handlers

import (
	"fmt"
	"net/http"
	"time"
)

func notBetween(w http.ResponseWriter, input string, inputName string, minChar int, maxChar int) bool {
	if len(input) < minChar {
		responseString := fmt.Sprintf("Please keep your %s over the %d-character requirement.", inputName, minChar)
		http.Error(w, responseString, http.StatusForbidden)
		return true
	}
	if len(input) > maxChar {
		responseString := fmt.Sprintf("Please keep your %s under the %d-character limit.", inputName, maxChar)
		http.Error(w, responseString, http.StatusForbidden)
		return true
	}
	return false
}

// Returns true if the date/time schedule is at least a minute
// after the current time, else false
func scheduleValid(schedule string) bool {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Could not load EST location", err)
		return false
	}
	scheduleTime, err := time.ParseInLocation(time.RFC3339, schedule, est)
	if err != nil {
		fmt.Println("Schedule string is not valid", err)
		return false
	}
	currentTime := time.Now().In(est).Add(time.Minute)
	return scheduleTime.After(currentTime)
}
