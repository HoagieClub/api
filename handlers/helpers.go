package handlers

import (
	"hoagie-profile/auth"
	//"os"
	"strings"
)

func getUser(authorizationHeader string) (user auth.User, success bool) {
	accessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")

	user, err := auth.GetUser(accessToken)
	// if os.Getenv("HOAGIE_MODE") == "debug" {
	// 	user = auth.User{Email: "meatball@princeton.edu", Name: "Meatball Hoagie"}
	// } else 
	if err != nil {
		return auth.User{}, false
	}
	return user, true
}
