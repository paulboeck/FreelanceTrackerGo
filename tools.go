//go:build tools
// +build tools

package tools

// This file declares dependencies on build tools so they are tracked in go.mod
// The build constraint ensures this file is never actually compiled

import (
	_ "github.com/pressly/goose/v3/cmd/goose"
)
