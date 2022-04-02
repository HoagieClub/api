package handlers

import (
	b64 "encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	CONSUMER_KEY    = os.Getenv("CONSUMER_KEY")
	CONSUMER_SECRET = os.Getenv("CONSUMER_SECRET")
)

const (
	BASE_URL          = "https://api.princeton.edu:443/mobile-app"
	COURSE_COURSES    = "/courses/courses"
	COURSE_TERMS      = "/courses/terms"
	DINING_LOCATIONS  = "/dining/locations"
	DINING_EVENTS     = "/dining/events"
	DINING_MENU       = "/dining/menu"
	PLACES_OPEN       = "/places/open"
	EVENTS_EVENTS     = "/events/events"
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
	princeton_token   = ""
)

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func _refreshToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", REFRESH_TOKEN_URL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+b64.StdEncoding.EncodeToString([]byte(CONSUMER_KEY+":"+CONSUMER_SECRET)))
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var refreshResponse RefreshTokenResponse
	err = json.Unmarshal(b, &refreshResponse)
	if err != nil {
		return "", err
	}

	return refreshResponse.AccessToken, nil
}
