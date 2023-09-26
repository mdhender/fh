// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package engine

type Options struct {
	root string // absolute path to root of file system
}

type Option func(*Options) error
