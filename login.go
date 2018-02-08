// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

type user struct {
	Token  *oauth2.Token
	UserID string
}

func authHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ""
		cookie, err := r.Cookie("userId")
		if err == nil {
			userID = cookie.Value
		}

		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		token, err := env.DataStore.GetToken(userID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// This will possibly do a token refresh
		tokenSource := env.OAuthConfig.TokenSource(r.Context(), token)
		newToken, err := tokenSource.Token()

		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}

		// Store the token in the context for use by a page handler
		u := user{Token: newToken, UserID: userID}

		ctx := context.WithValue(r.Context(), "user", u)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

func getOAuthClient(ctx context.Context) *http.Client {
	u := ctx.Value("user").(user)
	token := u.Token

	return env.OAuthConfig.Client(ctx, token)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	html, err := Asset("assets/login.html")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	t, _ := template.New("login").Parse(string(html))
	loginData := struct {
	}{}

	state := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Int())

	cookie := &http.Cookie{Name: "state", Value: state}
	http.SetCookie(w, cookie)

	t.Execute(w, loginData)
}

func oauthRedirect(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("state")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	url := env.OAuthConfig.AuthCodeURL(cookie.Value, oauth2.AccessTypeOnline)

	http.Redirect(w, r, url, http.StatusFound)
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	e := r.URL.Query().Get("error")

	// Check error
	if e != "" {
		redirectToError(w, r, errors.New(e))
		return
	}

	// Validate state
	cookie, err := r.Cookie("state")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	if state != cookie.Value {
		redirectToError(w, r, errors.New("Authorization Error"))
		return
	}

	// Exchange Token
	token, err := env.OAuthConfig.Exchange(ctx, code)

	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Get the username out of the token
	userName := token.Extra("userName").(string)

	// Send a cookie back to the client
	expiration := time.Now().Add(30 * 24 * time.Hour) // 1 month
	cookie = &http.Cookie{Name: "userId", Value: userName, Expires: expiration}

	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}
