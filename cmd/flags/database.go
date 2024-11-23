package flags

import (
	"fmt"
	"strings"
)

type Database string

// These are all the current databases supported.
// If you want to add any, just add a line below here and remember
// include it in the `AllowedDBDrivers` slice.
const (
	Postgres Database = "postgres"
	Sqlite   Database = "sqlite"
)

var AllowedDBDrivers = []string{string(Postgres), string(Sqlite)}

func (f Database) String() string {
	return string(f)
}

func (f *Database) Type() string {
	return "Database"
}

func (f *Database) Set(value string) error {
	for _, database := range AllowedDBDrivers {
		if database == value {
			*f = Database(value)
			return nil
		}
	}

	return fmt.Errorf("Database to use. Allowed values: %s", strings.Join(AllowedDBDrivers, ", "))
}
