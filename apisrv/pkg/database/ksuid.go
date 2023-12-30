package database

import "regexp"

const (
	MinKsuid = "000000000000000000000000000"
	MaxKsuid = "aWgEPTl1tmebfsQzFP4bxwgy80V"
)

var validKsuidPattern = regexp.MustCompile("^[a-zA-Z0-9]{27}$")

func ValidKsuid(s string) bool {
	return validKsuidPattern.MatchString(s)
}
