// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// StudioConfig extends the oauth2.Config to get a hook for storing refreshed tokens
type StudioConfig struct {
	*oauth2.Config
}

// StoreToken is called whenever a new Token is received and when a Token is refreshed. Studio Tokens are refreshed regularly and are one time use. It is important to always store the latest token.
func (c *StudioConfig) StoreToken(token *oauth2.Token) error {
	userName := token.Extra("userName").(string)
	fmt.Println("Saving Token: " + userName)
	env.DataStore.StoreToken(userName, token)
	return nil
}

func (c *StudioConfig) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	if err := c.StoreToken(token); err != nil {
		return nil, err
	}
	return token, nil
}

// Client retrieves an http.Client that is all setup for providing OAuth tokens and refreshing them as needed
func (c *StudioConfig) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx, t))
}

// TokenSource creates a ReuseTokenSource that will call back to StoreToken when a refresh happens
func (c *StudioConfig) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	rts := &StudioTokenSource{
		source: c.Config.TokenSource(ctx, t),
		config: c,
	}
	return oauth2.ReuseTokenSource(t, rts)
}

// StudioTokenStore extends the oauth2.TokenSource to provide a handle back to the StudioConfig so that StoreToken can be called
type StudioTokenSource struct {
	source oauth2.TokenSource
	config *StudioConfig
}

func (t *StudioTokenSource) Token() (*oauth2.Token, error) {
	token, err := t.source.Token()
	if err != nil {
		return nil, err
	}
	if err := t.config.StoreToken(token); err != nil {
		return nil, err
	}
	return token, nil
}
