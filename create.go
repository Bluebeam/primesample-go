/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Bluebeam Inc. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package main

import (
	"html/template"
	"net/http"
)

func createPage(w http.ResponseWriter, r *http.Request) {
	client := getOAuthClient(r.Context())

	projectID := r.FormValue("project")
	sessionName := r.FormValue("session")

	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("sessionFile")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	defer file.Close()

	projectFilesResponse, err := startFileUpload(client, projectID, handler.Filename)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	err = uploadToAWS(projectFilesResponse, file, handler.Size)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	err = confirmUpload(client, projectID, projectFilesResponse.ID)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	sessionResponse, err := createSession(client, sessionName)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	checkoutResponse, err := checkoutFileToSession(client, sessionResponse.ID, projectID, projectFilesResponse.ID)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	html, err := Asset("assets/create.html")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	t, _ := template.New("createSession").Parse(string(html))

	createSessionData := struct {
		SessionName   string
		SessionID     string
		ProjectID     string
		FileSessionID int
		FileProjectID int
	}{SessionName: sessionName, SessionID: sessionResponse.ID, ProjectID: projectID, FileSessionID: checkoutResponse.ID, FileProjectID: projectFilesResponse.ID}

	t.Execute(w, createSessionData)
}
