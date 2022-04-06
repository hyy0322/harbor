package migrator

import "fmt"

type Migrator interface {
	Migrate(dbType string) error
}

var migrators = make(map[string]Migrator)

func RegisterMigrator(name string, migrator Migrator) {
	migrators[name] = migrator
}

func Migrate(dbType string) {
	for _, m := range migrators {
		err := m.Migrate(dbType)
		fmt.Println(err)
	}
}

func init() {
	RegisterMigrator("artifact", artifactMigrators)
}
