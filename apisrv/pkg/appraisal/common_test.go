package appraisal

import (
	_ "embed"
)

//go:embed seed.sql
var seedSql string

//go:embed seed-expire.sql
var seedExpireSql string

//go:embed seed-revoke.sql
var seedRevokeSql string

//go:embed seed-copy.sql
var seedCopySql string
