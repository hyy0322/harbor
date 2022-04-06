package mysql

import (
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
)

type Artifact struct {
}

func (a *Artifact) Connect() error {
	return nil
}

func (a *Artifact) Dump() ([]dao.Artifact, error) {
	return nil, nil
}

func (a *Artifact) Insert([]dao.Artifact) error {
	return nil
}
