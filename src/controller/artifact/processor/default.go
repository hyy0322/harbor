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
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/controller/artifact/schema"
	schemav1alpha "github.com/goharbor/harbor/src/controller/artifact/schema/v1alpha"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"

	"github.com/docker/distribution/manifest/schema2"
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
	return nil
}

// The default processor will process user-defined artifact.
// AbstractMetadata will abstract data in a specific way.
// Annotation keys in artifact annotation will decide which content will be processed in artifact.
// Here is a manifest example:
//{
//   "schemaVersion": 2,
//   "config": {
//       "mediaType": "application/vnd.caicloud.model.config.v1alpha1+json",
//       "digest": "sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
//       "size": 187,
//       "annotations": {
//           "org.goharbor.artifact.schema.version": "v1/alpha",
//           "org.goharbor.artifact.skiplist": "metrics,git"
//       }
//   },
//   "layers": [
//       {
//           "mediaType": "image/png",
//           "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
//           "size": 166015,
//           "annotations": {
//               "org.goharbor.artifact.icon": "true"
//           }
//       },
//       {
//           "mediaType": "application/tar+gzip",
//           "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
//           "size": 166015
//       }
//   ]
//}
func (d *defaultProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	if artifact.ManifestMediaType != v1.MediaTypeImageManifest && artifact.ManifestMediaType != schema2.MediaTypeManifest {
		return nil
	}
	// get manifest
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
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

	// Populate all metadata into the ExtraAttrs first.
	// If there are annotation key AnnotationArtifactSchemaVersion in artifact manifest.config, metadata will be processed in specific rules.
	artifact.ExtraAttrs = metadata

	// check manifest.config.annotation key AnnotationArtifactSchemaVersion
	schemaVersion, ok := mani.Config.Annotations[schema.AnnotationArtifactSchemaVersion]
	if !ok {
		return nil
	}

	// check schema version
	if schemaVersion != schema.SchemaVersionV1Alpha {
		return fmt.Errorf("unsupported artifact config schema version %s", schemaVersion)
	}

	// check manifest.config.annotation key AnnotationArtifactSkipList
	skipKeyListStr, ok := mani.Config.Annotations[schemav1alpha.AnnotationArtifactSkipList]
	if !ok {
		return nil
	}

	skipKeyList := strings.Split(skipKeyListStr, ",")

	// move keys in skipKeyList
	for _, skipKey := range skipKeyList {
		delete(metadata, skipKey)
	}

	artifact.ExtraAttrs = metadata

	return nil
}
func (d *defaultProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*Addition, error) {
	// In schema version v1/alpha, addition not support for user-defined artifact yet.
	// It will be support in the future
	// return error directly
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage("the processor for artifact %s not found, cannot get the addition", artifact.Type)
}
