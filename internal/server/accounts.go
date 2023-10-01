// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

type Account struct {
	Id           int
	Email        string
	Handle       string // display name
	HashedSecret []byte
}

func (a Account) IsAuthenticated() bool {
	return false
}

func (a Account) IsAuthorized(role string) bool {
	return false
}
