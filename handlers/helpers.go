package handlers

import (
	"hoagie-profile/auth"
	"os"
	"strings"
)

func getUser(authorizationHeader string) (user string, success bool) {
	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")

	user, err := auth.GetUser(accessToken)
	if os.Getenv("HOAGIE_MODE") == "debug" {
		user = "test@princeton.edu"
	} else if err != nil {
		return "", false
	}
	return user, true
}
