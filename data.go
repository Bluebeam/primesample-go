// Copyright (c) Bluebeam Inc. All rights reserved.
//
// Licensed under the MIT License. See LICENSE in the project root for license information.

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"golang.org/x/oauth2"
)

type DataStore interface {
	New()
	Close()
	StoreToken(userID string, token *oauth2.Token) error
	GetToken(userID string) (*oauth2.Token, error)
}

type BoltDBStore struct {
	DB *bolt.DB
}

func (s *BoltDBStore) New() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Tokens"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}

		return nil
	})

	s.DB = db
}

func (s *BoltDBStore) Close() {
	s.DB.Close()
}

func (s *BoltDBStore) StoreToken(userID string, token *oauth2.Token) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Tokens"))

		data, err := json.Marshal(token)
		if err != nil {
			return err
		}

		err = b.Put([]byte(userID), data)
		return err
	})
}

func (s *BoltDBStore) GetToken(userID string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Tokens"))

		data := b.Get([]byte(userID))
		if data == nil {
			return fmt.Errorf("Token Not Found: %s", userID)
		}

		return json.Unmarshal(data, token)
	})

	return token, err
}
