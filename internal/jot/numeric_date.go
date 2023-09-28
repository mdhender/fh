// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"encoding/json"
	"time"
)

// NumericDate is a numeric date representing seconds past 1970-01-01 00:00:00Z.
type NumericDate time.Time

// MarshalJSON implements the json.Marshaler interface.
func (d NumericDate) MarshalJSON() ([]byte, error) {
	n := time.Time(d).Unix()
	return json.Marshal(n)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *NumericDate) UnmarshalJSON(data []byte) error {
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*d = NumericDate(time.Unix(n, 0))
	return nil
}

// String implements the Stringer interface.
func (d NumericDate) String() string {
	return time.Time(d).Format(time.RFC3339)
}
