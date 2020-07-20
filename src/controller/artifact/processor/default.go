// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	schemaV1 "github.com/goharbor/harbor/src/controller/artifact/schema/v1"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/mitchellh/mapstructure"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ArtifactTypeUnknown defines the type for the unknown artifacts
const ArtifactTypeUnknown = "UNKNOWN"

var (
	// DefaultProcessor is to process artifact which has no specific processor
	DefaultProcessor = &defaultProcessor{RegCli: registry.Cli}

	artifactTypeRegExp = regexp.MustCompile(`^application/vnd\.[^.]*\.(.*)\.config\.[^.]*\+json$`)
)

// the default processor to process artifact
// currently, it only tries to parse the artifact type from media type
type defaultProcessor struct {
	RegCli registry.Client
}

func (d *defaultProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	// try to parse the type from the media type
	strs := artifactTypeRegExp.FindStringSubmatch(artifact.MediaType)
	if len(strs) == 2 {
		return strings.ToUpper(strs[1])
	}
	// can not get the artifact type from the media type, return unknown
	return ArtifactTypeUnknown
}
func (d *defaultProcessor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	artifactConfig := artifact.ExtraAttrs
	schemaMap, ok := artifactConfig[schemaV1.XHarborAttributes]
	if !ok {
		return nil
	}
	schema := &schemaV1.Schema{}
	err := mapstructure.Decode(schemaMap, schema)
	if err != nil {
		log.Warningf("unsupported artifact config schema: %v", err)
		return nil
	}
	additionTypes := make([]string, 0)
	for _, addition := range schema.Additions {
		additionTypes = append(additionTypes, addition.Name)
	}
	return additionTypes
}
func (d *defaultProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	if artifact.ManifestMediaType != v1.MediaTypeImageManifest && artifact.ManifestMediaType != schema2.MediaTypeManifest {
		return nil
	}
	// get manifest
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		log.Infof("%v", mani)
		log.Warningf("unmarshal error: %+v", err)
		return err
	}
	// get config layer
	_, blob, err := d.RegCli.PullBlob(artifact.RepositoryName, mani.Config.Digest.String())
	if err != nil {
		log.Warningf("PullBlob error: %+v", err)
		return err
	}
	defer blob.Close()
	// parse metadata from config layer
	metadata := map[string]interface{}{}
	if err := json.NewDecoder(blob).Decode(&metadata); err != nil {
		log.Warningf("decode blob error: %+v", err)
		return err
	}

	// populate all metadata into the ExtraAttrs first
	// if there are xHarborAttribute in artifact config
	// metadata will be processed in specific rules
	artifact.ExtraAttrs = metadata

	// parse schema
	schemaMap, ok := metadata[schemaV1.XHarborAttributes]
	if !ok {
		return nil
	}
	schema := &schemaV1.Schema{}
	err = mapstructure.Decode(schemaMap, schema)
	if err != nil {
		return fmt.Errorf("unsupported artifact config schema: %v", err)
	}

	// check schema version
	if schema.SchemaVersion != schemaV1.SchemaVersionV1 {
		return fmt.Errorf("unsupported artifact config schema version %d", schema.SchemaVersion)
	}

	// move keys in skipKeyList
	for _, skipKey := range schema.SkipKeyList {
		delete(metadata, skipKey)
	}

	artifact.ExtraAttrs = metadata

	return nil
}
func (d *defaultProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*Addition, error) {
	// no xHarborAttributes in artifact extraAttrs
	schemaMap, ok := artifact.ExtraAttrs[schemaV1.XHarborAttributes]
	if !ok {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("no addition defined in config layer for artifact %s, cannot get the addition", artifact.Type)
	}

	schema := &schemaV1.Schema{}
	err := mapstructure.Decode(schemaMap, schema)
	if err != nil {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("addition %s isn't supported for %s", addition, artifact.Type)
	}

	var (
		contentType string
		layerDigest string
	)

	// get addition
	for _, add := range schema.Additions {
		if add.Name == addition {
			contentType = add.ContentType
			layerDigest = add.Digest
		}
	}

	if layerDigest == "" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("addition %s isn't supported for %s", addition, artifact.Type)
	}

	// get addition from layer
	_, blob, err := d.RegCli.PullBlob(artifact.RepositoryName, layerDigest)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(blob)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	return &Addition{
		Content:     content,
		ContentType: contentType,
	}, nil
}
