package migrator

import (
	"fmt"
	"github.com/goharbor/harbor/src/migration/migrator/mysql"
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
)

type ArtifactMigrator interface {
	Connect() error
	Dump() ([]dao.Artifact, error)
	Insert([]dao.Artifact) error
}

var artifactMigrators ArtifactMigrators

func RegisterArtifactMigrator(name string, artifactMigrator ArtifactMigrator) {
	artifactMigrators[name] = artifactMigrator
}

type ArtifactMigrators map[string]ArtifactMigrator

func (a ArtifactMigrators) Migrate(dbType string) error {
	err := a[dbType].Connect()
	fmt.Println(err)
	data, err := a[dbType].Dump()
	fmt.Println(err)
	err = a[dbType].Insert(data)
	fmt.Println(err)
	return nil
}

func init() {
	RegisterArtifactMigrator("mysql", &mysql.Artifact{})
}
