// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import (
	"encoding/json"
	"log"
	"os"
)

type JSONStore struct {
	Sessions map[string]JSONSession `json:"sessions,omitempty"`
	Signing  struct {
		Salt   []byte `json:"salt,omitempty"`
		Key    []byte `json:"key,omitempty"`
		Pepper []byte `json:"pepper,omitempty"`
	} `json:"signing,omitempty"`
}

type JSONSession struct {
	Account   string `json:"acct"`
	ExpiresAt string `json:"exp"`
}

func Load(path string) (*JSONStore, error) {
	js := &JSONStore{
		Sessions: make(map[string]JSONSession),
	}

	if path == ":memory:" {
		return js, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &js); err != nil {
		return nil, err
	}
	return js, nil
}

func Save(js *JSONStore, path string) error {
	log.Printf("[sessions] save: %q\n", path)
	data, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
