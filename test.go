// build+ ignore
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func tmain() {

	resp, err := http.Get("http://fed.princeton.edu/cas/oauth2.0/authorize")

	// Error checking of the http.Get() request
	if err != nil {
		log.Fatal(err)
	}

	// Resource leak if response body isn't closed
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	// Error checking of the ioutil.ReadAll() request
	if err != nil {
		log.Fatal(err)
	}

	bodyString := string(bodyBytes)

	// Print Statements

	fmt.Println(resp.StatusCode)

	fmt.Println(resp.Header)
	// Note if you try to access a key in the map that doesn't exist you will get an error, a check on the key should be made to prevent this, this code isn't meant for production!
	fmt.Println(resp.Header["Content-Type"])
	fmt.Println(resp.Header["Content-Type"][0]) // This is just to show you how to access items in the hash map

	fmt.Println(bodyString)
}
