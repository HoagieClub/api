package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/rs/cors"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(domain string, token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(domain + ".well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := fmt.Errorf("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

type User struct {
	Email string `json:"email"`
	Name  string
}

func GetUser(accessToken string) (User, error) {
	parser := jwt.Parser{}
	claims := jwt.MapClaims{}

	// Parsing unverified is okay as long as handlers all go through
	// a JWT middleware before running this.
	_, _, err := parser.ParseUnverified(accessToken, claims)

	if err != nil {
		return User{}, err
	}
	email := claims["https://hoagie.io/email"]
	emailString, ok := email.(string)
	if !ok {
		return User{}, fmt.Errorf("email not string")
	}
	name := claims["https://hoagie.io/name"]
	nameString, ok := name.(string)
	if !ok {
		return User{}, fmt.Errorf("name not string")
	}
	return User{
		Email: emailString,
		Name:  nameString,
	}, nil
}

func Middleware(domain string, audience string) *jwtmiddleware.JWTMiddleware {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Verify 'aud' claim
		checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(audience, false)
		if !checkAud {
			return token, errors.New("Invalid audience.")
		}
		// Verify 'iss' claim
		iss := domain
		checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
		if !checkIss {
			return token, errors.New("Invalid issuer.")
		}

		cert, err := getPemCert(domain, token)
		if err != nil {
			panic(err.Error())
		}

		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		return result, nil
	}
	return jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: keyFunc,
		SigningMethod:       jwt.SigningMethodRS256,
	})
}

func CorsWrapper(runtimeMode string) *cors.Cors {
	var corsWrapper *cors.Cors
	if runtimeMode == "debug" {
		// CORS for development only
		corsWrapper = cors.New(cors.Options{
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
		})
	} else {
		corsWrapper = cors.New(cors.Options{
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
			AllowedOrigins: []string{"https://*.hoagie.io"},
		})
	}
	return corsWrapper
}
