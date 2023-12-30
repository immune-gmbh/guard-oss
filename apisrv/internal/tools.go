//go:build tools
// +build tools

// dummy package to pull in build dependencies
// it is necessary for tools that can not be installed using the
// current go install pkg@version syntax, which is true for tools
// that have a replace directive inside their go.mod
package tools

import (
	_ "git.sr.ht/~emersion/go-jsonschema"
	_ "github.com/99designs/gqlgen"
	_ "github.com/google/go-licenses"
)
