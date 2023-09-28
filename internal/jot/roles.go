// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"encoding/json"
	"sort"
)

// Roles is a map of all roles assigned to a Subject.
type Roles map[string]bool

// MarshalJSON implements the json.Marshaler interface.
func (r Roles) MarshalJSON() ([]byte, error) {
	var roles []string
	for k := range r {
		roles = append(roles, k)
	}
	sort.Strings(roles)
	return json.Marshal(roles)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Roles) UnmarshalJSON(data []byte) error {
	var roles []string
	if err := json.Unmarshal(data, &roles); err != nil {
		return err
	}
	rr := make(map[string]bool)
	for _, role := range roles {
		rr[role] = true
	}
	*r = rr
	return nil
}
