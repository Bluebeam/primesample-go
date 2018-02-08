// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var env *environment

// Root singleton to store our application state
type environment struct {
	OAuthConfig *StudioConfig
	DataStore   DataStore
}

func main() {
	conf, err := initOauth()
	if err != nil {
		log.Fatal(err)
		return
	}

	// BoltDB is used as a simple Database to store OAuth tokens
	dataStore := &BoltDBStore{}
	dataStore.New()

	env = &environment{OAuthConfig: conf, DataStore: dataStore}

	// These pages are protected by authentication
	http.Handle("/", authHandler(http.HandlerFunc(homePage)))
	http.Handle("/create", authHandler(http.HandlerFunc(createPage)))
	http.Handle("/finish", authHandler(http.HandlerFunc(finishPage)))

	// The pages are all part of the OAuth flow
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/oauth", oauthRedirect)
	http.HandleFunc("/callback", oauthCallback)

	http.HandleFunc("/error", errorPage)

	// CSS and JS are found in assets.go
	http.HandleFunc("/style.css", cssHandler)
	http.HandleFunc("/script.js", scriptHandler)

	http.ListenAndServe(":5000", nil)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	css, _ := Asset("assets/style.css")
	w.Header().Set("Content-Type", "text/css")
	w.Write(css)
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	script, _ := Asset("assets/script.js")
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(script)
}

func initOauth() (*StudioConfig, error) {
	// Do not be alarmed by this call. This is simply because the Studio Auth server expects the clientId and secretId to be in query parameters rather than the Authorization header
	oauth2.RegisterBrokenAuthHeaderProvider("https://authserver.bluebeam.com/auth/token")

	config := struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		URL          string `json:"url"`
	}{}

	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		// Try Environment Variables
		config.ClientID = os.Getenv("CLIENT_ID")
		config.ClientSecret = os.Getenv("CLIENT_SECRET")
		config.URL = os.Getenv("URL")
	} else {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return nil, err
		}
	}

	conf := &StudioConfig{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Scopes:       []string{"full_user", "jobs"},
			RedirectURL:  config.URL + "/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://authserver.bluebeam.com/auth/oauth/authorize",
				TokenURL: "https://authserver.bluebeam.com/auth/token",
			},
		},
	}

	return conf, nil
}
