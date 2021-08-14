package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/cas.v2"
)

type TemplateBinding struct {
	Username   string
	Attributes cas.UserAttributes
}

// NewRouter creeates a mux.Router which has login and logout
// functionality pre-configured for CAS.
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	s := r.Host("localhost").Subrouter()

	s.HandleFunc("/login", LoginUser)
	s.HandleFunc("/logout", LogoutUser)

	return r
}

// GetAttributes is an alias for cas.Attributes
func GetAttributes(r *http.Request) cas.UserAttributes {
	return cas.Attributes(r)
}

// IsAuthenticated is an alias for cas.isAuthenticated
func IsAuthenticated(r *http.Request) bool {
	return cas.IsAuthenticated(r)
}

// GetUsername is an alias for cas.Username
func GetUsername(r *http.Request) string {
	return cas.Username(r)
}

// LoginUser redirects the user to CAS authentication if they are not
// logged in, and back to the "return" parameter otherwise
func LoginUser(w http.ResponseWriter, r *http.Request) {
	if !cas.IsAuthenticated(r) {
		cas.RedirectToLogin(w, r)
		return
	}

	redirect := r.URL.Query().Get("return")
	http.Redirect(w, r, redirect, http.StatusFound)
}

// LogoutUser logs the user out from CAS, if they are logged in, and redirects to app root
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	if cas.IsAuthenticated(r) {
		cas.RedirectToLogout(w, r)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
