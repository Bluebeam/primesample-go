// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DefaultSessionPermissions struct {
	Type  string `json:"Type"`
	Allow string `json:"Allow"`
}

type CreateSession struct {
	Name               string                      `json:"Name"`
	Notification       bool                        `json:"Notification"`
	Restricted         bool                        `json:"Restricted"`
	SessionEndDate     time.Time                   `json:"SessionEndDate"`
	DefaultPermissions []DefaultSessionPermissions `json:"DefaultPermissions"`
}

type Session struct {
	Name           string `json:"Name,omitempty"`
	Notification   *bool  `json:"Notification,omitempty"`
	Restricted     *bool  `json:"Restricted,omitempty"`
	SessionEndDate string `json:"SessionEndDate,omitempty"`
	OwnerEmailOrID string `json:"OwnerEmailOrId,omitempty"`
	Status         string `json:"Status,omitempty"`
}

type CreateSessionResponse struct {
	ID string `json:"Id"`
}

func createSession(client *http.Client, sessionName string) (*CreateSessionResponse, error) {
	createSessionData := CreateSession{
		Name:           sessionName,
		Notification:   true,
		Restricted:     false,
		SessionEndDate: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		DefaultPermissions: []DefaultSessionPermissions{
			DefaultSessionPermissions{
				Type:  "SaveCopy",
				Allow: "Allow",
			},
			DefaultSessionPermissions{
				Type:  "PrintCopy",
				Allow: "Allow",
			},
			DefaultSessionPermissions{
				Type:  "Markup",
				Allow: "Allow",
			},
			DefaultSessionPermissions{
				Type:  "MarkupAlert",
				Allow: "Allow",
			},
			DefaultSessionPermissions{
				Type:  "AddDocuments",
				Allow: "Allow",
			},
		},
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(createSessionData)

	req, err := http.NewRequest("POST", "https://studioapi.bluebeam.com/publicapi/v1/sessions", b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return nil, err
	}

	response := &CreateSessionResponse{}
	json.NewDecoder(resp.Body).Decode(response)

	return response, nil
}

func setSessionStatus(client *http.Client, sessionID, status string) (*SessionResponse, error) {
	session := Session{Status: status}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(session)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/sessions/%s", sessionID)
	req, err := http.NewRequest("PUT", url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return nil, err
	}

	response := &SessionResponse{}
	json.NewDecoder(resp.Body).Decode(response)

	return response, nil
}

func sessionDelete(client *http.Client, sessionID string) error {
	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/sessions/%s", sessionID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return err
	}

	return nil
}

func startSnapshot(client *http.Client, sessionID string, fileID int) error {
	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/sessions/%s/files/%v/snapshot", sessionID, fileID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return err
	}

	return nil
}

func getSnapshotStatus(client *http.Client, sessionID string, fileID int) (*SnapshotResponse, error) {
	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/sessions/%s/files/%v/snapshot", sessionID, fileID)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return nil, err
	}

	response := &SnapshotResponse{}
	json.NewDecoder(resp.Body).Decode(response)

	return response, nil
}
