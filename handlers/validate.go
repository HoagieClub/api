package handlers

import (
	"fmt"
	"net/http"
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
