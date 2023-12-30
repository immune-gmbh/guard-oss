package testing

import (
	"path"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func GetSchemaPath() string {
	// $REPODIR/guard/internal/testing
	return path.Join(basepath, "..", "..", "..", "_schemas")
}
