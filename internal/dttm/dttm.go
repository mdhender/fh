// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package dttm

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date is time formatted per RFC 3339.
type Date time.Time

// MarshalJSON implements the json.Marshaler interface.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).UTC().Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Date) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	} else if t, err := time.Parse(v, time.RFC3339); err != nil {
		return fmt.Errorf("exp: %w", err)
	} else {
		*d = Date(t)
	}
	return nil
}
