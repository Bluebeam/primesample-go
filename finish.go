// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func finishPage(w http.ResponseWriter, r *http.Request) {
	client := getOAuthClient(r.Context())

	sessionID := r.FormValue("sessionId")
	projectID := r.FormValue("projectId")
	fileSessionID, _ := strconv.ParseInt(r.FormValue("fileSessionId"), 10, 32)
	fileProjectID, _ := strconv.ParseInt(r.FormValue("fileProjectId"), 10, 32)

	// Set Session to Finalizing to boot people
	_, err := setSessionStatus(client, sessionID, "Finalizing")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Initiate Snapshot
	err = startSnapshot(client, sessionID, int(fileSessionID))
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Poll the snapshot status every 5 seconds until complete or an error
	var snapshotResponse *SnapshotResponse

outer:
	for {
		snapshotResponse, err = getSnapshotStatus(client, sessionID, int(fileSessionID))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(snapshotResponse)

		switch snapshotResponse.Status {
		case "Complete":
			break outer
		case "Error":
			redirectToError(w, r, errors.New("Shapshot error"))
			return

		}

		time.Sleep(5 * time.Second)
	}

	// Download Snapshot
	resp, err := http.Get(snapshotResponse.DownloadURL)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	fileReader := resp.Body

	// Delete Session
	err = sessionDelete(client, sessionID)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Start checkin
	projectFilesResponse, err := checkinProjectFile(client, projectID, int(fileProjectID))
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Upload new revision to Aws
	err = uploadToAWS(projectFilesResponse, fileReader, resp.ContentLength)
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Confirm checkin
	err = confirmProjectCheckin(client, projectID, int(fileProjectID), "Checkin from Roundtripper")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Kick off job to flatten the file
	_, err = flattenProjectFile(client, projectID, int(fileProjectID))
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	// Generate a share link to the file
	sharedLinkResponse, err := getSharedLink(client, projectID, int(fileProjectID))
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	html, err := Asset("assets/finish.html")
	if err != nil {
		redirectToError(w, r, err)
		return
	}

	t, _ := template.New("finishSession").Parse(string(html))

	finishSessionData := struct {
		ProjectLink string
	}{ProjectLink: sharedLinkResponse.ShareLink}

	t.Execute(w, finishSessionData)
}
