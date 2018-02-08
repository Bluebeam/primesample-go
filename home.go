// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		return
	}

	html, err := Asset("assets/home.html")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	t, _ := template.New("login").Parse(string(html))

	ctx := r.Context()
	u := ctx.Value("user").(user)

	client := env.OAuthConfig.Client(ctx, u.Token)

	req, err := http.NewRequest("GET", "https://studioapi.bluebeam.com/publicapi/v1/projects", nil)
	resp, err := client.Do(req)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		redirectToError(w, r, err)
		return
	}

	projects := ProjectsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	fmt.Println(projects)

	homeData := struct {
		UserID   string
		Projects []*Project
	}{u.UserID, projects.Projects}

	t.Execute(w, homeData)
}
