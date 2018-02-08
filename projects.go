// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Project struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
}

type ProjectsResponse struct {
	Projects   []*Project
	TotalCount int
}

type ProjectFilesRequest struct {
	Name           string `json:"Name"`
	ParentFolderID int    `json:"ProjectFolderId"`
	Size           int    `json:"Size"`
	CRC            string `json:"CRC"`
}

type ProjectFilesResponse struct {
	ID                int    `json:"Id"`
	UploadUrl         string `json:"UploadUrl"`
	UploadContentType string `json:"UploadContentType"`
}

type CheckoutToSession struct {
	SessionID string `json:"SessionId"`
}

type CheckoutToSessionResponse struct {
	SessionID string `json:"SessionId"`
	ID        int    `json:"Id"`
}

type SessionResponse struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	Restricted     bool   `json:"Restricted"`
	ExpirationDate string `json:"ExpirationDate"`
	SessionEndDate string `json:"SessionEndDate"`
	Version        int    `json:"Version"`
	Created        string `json:"Created"`
	InviteURL      string `json:"InviteUrl"`
	OwnerEmail     string `json:"OwnerEmail"`
	Status         string `json:"Status"`
}

type CheckinFromSession struct {
	Comment string `json:"Comment"`
}

type SnapshotResponse struct {
	Status           string `json:"Status"`
	StatusTime       string `json:"StatusTime"`
	LastSnapshotTime string `json:"StatusSnapshotTime"`
	DownloadURL      string `json:"DownloadUrl"`
}

type JobFlattenOptions struct {
	Image             bool `json:"Image"`
	Ellipse           bool `json:"Ellipse"`
	Stamp             bool `json:"Stamp"`
	Snapshot          bool `json:"Snapshot"`
	TextAndCallout    bool `json:"TextAndCallout"`
	InkAndHighlighter bool `json:"InkAndHighlighter"`
	LineAndDimension  bool `json:"LineAndDimension"`
	MeasureArea       bool `json:"MeasureArea"`
	Polyline          bool `json:"Polyline"`
	PolygonAndCloud   bool `json:"PolygonAndCloud"`
	Rectangle         bool `json:"Rectangle"`
	TextMarkups       bool `json:"TextMarkups"`
	Group             bool `json:"Group"`
	FileAttachment    bool `json:"FileAttachment"`
	Flags             bool `json:"Flags"`
	Notes             bool `json:"Notes"`
	FormFields        bool `json:"FormFields"`
}

type JobFlatten struct {
	Recoverable     bool              `json:"Recoverable"`
	PageRange       string            `json:"PageRange,omitempty"`
	LayerName       string            `json:"LayerName,omitempty"`
	Options         JobFlattenOptions `json:"Options"`
	CurrentPassword string            `json:"CurrentPassword,omitempty"`
	OutputPath      string            `json:"OutputPath,omitempty"`
	OutputFileName  string            `json:"OutputFileName,omitempty"`
	Priority        int               `json:"Priority"`
}

type JobFlattenResponse struct {
	ID int `json:"Id"`
}

type ShareLink struct {
	ProjectFileID     int    `json:"ProjectFileID"`
	PasswordProtected bool   `json:"PasswordProtected"`
	Password          string `json:"Password"`
	Expires           string `json:"Expires"`
	Flatten           bool   `json:"Flatten"`
}

type SharedLinkResponse struct {
	ID        int    `json:"Id"`
	ShareLink string `json:"ShareLink"`
}

func startFileUpload(client *http.Client, projectID string, filename string) (*ProjectFilesResponse, error) {
	projectFile := ProjectFilesRequest{Name: filename, ParentFolderID: 0}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(projectFile)

	fmt.Println("Project ID: ", projectID)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files", projectID)
	req, err := http.NewRequest("POST", url, b)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return nil, err
	}

	projectFilesResponse := &ProjectFilesResponse{}
	err = json.NewDecoder(resp.Body).Decode(projectFilesResponse)
	if err != nil {
		return nil, err
	}

	fmt.Println("startFileUpload: ", projectFilesResponse.UploadUrl)

	return projectFilesResponse, nil
}

func uploadToAWS(projectFilesResponse *ProjectFilesResponse, file io.Reader, size int64) error {
	sterileClient := http.Client{}
	fmt.Println(projectFilesResponse)
	req, err := http.NewRequest("PUT", projectFilesResponse.UploadUrl, file)
	req.ContentLength = size
	req.Header.Add("Content-Type", projectFilesResponse.UploadContentType)
	req.Header.Add("x-amz-server-side-encryption", "AES256")
	resp, err := sterileClient.Do(req)
	if err != nil {
		return fmt.Errorf("uploadToAWS: %v", err)
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("uploadToAWS: ", string(body))

	return nil
}

func confirmUpload(client *http.Client, projectID string, projectFileID int) error {
	confirmURL := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files/%v/confirm-upload", projectID, projectFileID)
	req, err := http.NewRequest("POST", confirmURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return err
	}

	return nil
}

func checkoutFileToSession(client *http.Client, sessionID, projectID string, fileID int) (*CheckoutToSessionResponse, error) {
	checkoutToSession := CheckoutToSession{SessionID: sessionID}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(checkoutToSession)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files/%v/checkout-to-session", projectID, fileID)
	req, err := http.NewRequest("POST", url, b)
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

	response := &CheckoutToSessionResponse{}
	json.NewDecoder(resp.Body).Decode(response)

	return response, nil
}

func flattenProjectFile(client *http.Client, projectID string, fileID int) (*JobFlattenResponse, error) {
	job := &JobFlatten{
		Recoverable: true,
		PageRange:   "-1",
		Options: JobFlattenOptions{
			Image:             true,
			Ellipse:           true,
			Stamp:             true,
			Snapshot:          true,
			TextAndCallout:    true,
			InkAndHighlighter: true,
			LineAndDimension:  true,
			MeasureArea:       true,
			Polyline:          true,
			PolygonAndCloud:   true,
			Rectangle:         true,
			TextMarkups:       true,
			Group:             true,
			FileAttachment:    true,
			Flags:             true,
			Notes:             true,
			FormFields:        true,
		},
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(job)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files/%v/jobs/flatten", projectID, fileID)
	req, err := http.NewRequest("POST", url, b)
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

	response := &JobFlattenResponse{}
	json.NewDecoder(resp.Body).Decode(response)

	return response, nil
}

func confirmProjectCheckin(client *http.Client, projectID string, fileID int, comment string) error {
	checkin := CheckinFromSession{Comment: comment}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(checkin)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files/%v/confirm-checkin", projectID, fileID)
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return err
	}

	return nil
}

func checkinProjectFile(client *http.Client, projectID string, fileID int) (*ProjectFilesResponse, error) {
	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/files/%v/checkin", projectID, fileID)
	req, err := http.NewRequest("POST", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkHTTPResponse(resp); !ok {
		return nil, err
	}

	projectFilesResponse := &ProjectFilesResponse{}
	json.NewDecoder(resp.Body).Decode(projectFilesResponse)

	projectFilesResponse.ID = fileID

	return projectFilesResponse, nil
}

func getSharedLink(client *http.Client, projectID string, fileID int) (*SharedLinkResponse, error) {
	shareLink := ShareLink{ProjectFileID: fileID}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(shareLink)

	url := fmt.Sprintf("https://studioapi.bluebeam.com/publicapi/v1/projects/%s/sharedlinks", projectID)

	req, err := http.NewRequest("POST", url, b)
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

	shareLinkResponse := &SharedLinkResponse{}
	json.NewDecoder(resp.Body).Decode(shareLinkResponse)

	return shareLinkResponse, nil
}
