// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
)

func errorPage(w http.ResponseWriter, r *http.Request) {
	html, err := Asset("assets/error.html")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	t, _ := template.New("error").Parse(string(html))
	errorData := struct {
		Description string
	}{}

	desc, _ := url.QueryUnescape(r.URL.Query().Get("description"))

	errorData.Description = desc

	t.Execute(w, errorData)
}

func redirectToError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println(err)
	http.Redirect(w, r, "/error?description="+url.QueryEscape(err.Error()), http.StatusFound)
}

func checkHTTPResponse(resp *http.Response) (bool, error) {
	if resp.StatusCode >= http.StatusBadRequest {
		errBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, errors.New(resp.Status + " " + string(errBytes))
	}

	return true, nil
}
