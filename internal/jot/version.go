// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import "github.com/mdhender/fh/internal/semver"

// Version is the version of this package.
func Version() string {
	return version.String()
}

var version = semver.Version{
	Major: 0,
	Minor: 1,
	Patch: 0,
}
