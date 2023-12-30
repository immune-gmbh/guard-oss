package report

import "errors"

var (
	notFoundErr     = errors.New("not-found")
	invalidValueErr = errors.New("invalid")
	missingValueErr = errors.New("missing")
	dupValueErr     = errors.New("duplicate")
	noResponseErr   = errors.New("no-resp")
)

type insecureValueErr struct {
	Tag string
}

func (err insecureValueErr) Error() string {
	return err.Tag
}
