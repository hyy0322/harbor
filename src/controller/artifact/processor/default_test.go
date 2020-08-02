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
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/registry"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
)

const (
	ormbManifest = `{
    "schemaVersion":2,
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "config":{
        "mediaType":"application/vnd.caicloud.model.config.v1alpha1+json",
        "digest":"sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
        "size":187
    },
    "layers":[
        {
            "mediaType":"application/tar+gzip",
            "digest":"sha256:eb6063fecbb50a9d98268cb61746a0fd62a27a4af9e850ffa543a1a62d3948b2",
            "size":166022
        }
    ]
}`
	ormbManifestWithoutSkipKeyList = `{
    "schemaVersion":2,
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "config":{
        "mediaType":"application/vnd.caicloud.model.config.v1alpha1+json",
        "digest":"sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
        "size":187,
        "annotations": {
            "org.goharbor.artifact.schema.version": "v1/alpha"
        }
    },
    "layers":[
        {
            "mediaType":"application/tar+gzip",
            "digest":"sha256:eb6063fecbb50a9d98268cb61746a0fd62a27a4af9e850ffa543a1a62d3948b2",
            "size":166022
        }
    ]
}`
	ormbConfig = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    }
}`
	ormbConfigWithoutSkipKeyList = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "xHarborAttributes": {
        "schemaVersion": 1,
        "icon": "https://github.com/caicloud/ormb/raw/master/docs/images/logo.png",
        "additions": [
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "yaml",
                "digest": "sha256:c2b304e60b7aec6a32d50b0d2c064933a7554db9d5d55259ac236f630a1c1f86"
            },
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "readme",
                "digest": "sha256:6dba1ad7ead7a5ee681441ec4b56b6a24690de6411d4574b927ce654c303f3c6"
            }
		]
    }
}`
	ormbConfigWithEmptySkipKeyList = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "xHarborAttributes": {
        "schemaVersion": 1,
        "icon": "https://github.com/caicloud/ormb/raw/master/docs/images/logo.png",
        "additions": [
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "yaml",
                "digest": "sha256:c2b304e60b7aec6a32d50b0d2c064933a7554db9d5d55259ac236f630a1c1f86"
            },
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "readme",
                "digest": "sha256:6dba1ad7ead7a5ee681441ec4b56b6a24690de6411d4574b927ce654c303f3c6"
            }
		],
        "skipKeyList": []
    }
}`
	ormbConfigUnsupportedSchemaVersion = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "xHarborAttributes": {
        "schemaVersion": 2,
        "icon": "https://github.com/caicloud/ormb/raw/master/docs/images/logo.png",
        "additions": [
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "yaml",
                "digest": "sha256:c2b304e60b7aec6a32d50b0d2c064933a7554db9d5d55259ac236f630a1c1f86"
            },
            {
                "contentType": "text/plain; charset=utf-8",
                "name": "readme",
                "digest": "sha256:6dba1ad7ead7a5ee681441ec4b56b6a24690de6411d4574b927ce654c303f3c6"
            }
		],
        "skipKeyList": [
            "metrics",
			"hyperparameters"
        ]
    }
}`
	ormbConfigWithoutAdditions = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "xHarborAttributes": {
        "schemaVersion": 1,
        "icon": "https://github.com/caicloud/ormb/raw/master/docs/images/logo.png",
        "skipKeyList": [
            "metrics",
			"dataset"
        ]
    }
}`
	ormbConfigWithoutXHarborAttributes = `{
    "created": "2015-10-31T22:22:56.015925234Z",
    "author": "Ce Gao <gaoce@caicloud.io>",
    "description": "CNN Model",
    "tags": [
        "cv"
    ],
    "labels": {
        "tensorflow.version": "2.0.0"
    },
    "framework": "TensorFlow",
    "format": "SavedModel",
    "size": 9223372036854775807,
    "metrics": [
        {
            "name": "acc",
            "value": "0.9"
        }
    ],
    "hyperparameters": [
        {
            "name": "batch_size",
            "value": "32"
        }
    ],
    "signature": {
        "inputs": [
            {
                "name": "input_1",
                "size": [
                    224,
                    224,
                    3
                ],
                "dtype": "float64"
            }
        ],
        "outputs": [
            {
                "name": "output_1",
                "size": [
                    1,
                    1000
                ],
                "dtype": "float64"
            }
        ],
        "layers": [
            {
                "name": "conv"
            }
        ]
    },
    "training": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    },
    "dataset": {
        "git": {
            "repository": "git@github.com:caicloud/ormb.git",
            "revision": "22f1d8406d464b0c0874075539c1f2e96c253775"
        }
    }
}`
)

type defaultProcessorTestSuite struct {
	suite.Suite
	processor *defaultProcessor
	regCli    *registry.FakeClient
}

func (d *defaultProcessorTestSuite) SetupTest() {
	d.regCli = &registry.FakeClient{}
	d.processor = &defaultProcessor{
		RegCli: d.regCli,
	}
}

func (d *defaultProcessorTestSuite) TestGetArtifactType() {
	mediaType := ""
	art := &artifact.Artifact{MediaType: mediaType}
	processor := &defaultProcessor{}
	typee := processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "unknown"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal(ArtifactTypeUnknown, typee)

	mediaType = "application/vnd.oci.image.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("IMAGE", typee)

	mediaType = "application/vnd.cncf.helm.chart.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("HELM.CHART", typee)

	mediaType = "application/vnd.sylabs.sif.config.v1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("SIF", typee)

	mediaType = "application/vnd.caicloud.model.config.v1alpha1+json"
	art = &artifact.Artifact{MediaType: mediaType}
	processor = &defaultProcessor{}
	typee = processor.GetArtifactType(nil, art)
	d.Equal("MODEL", typee)
}

func (d *defaultProcessorTestSuite) TestAbstractMetadata() {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ormbManifest))
	d.Require().Nil(err)
	manifestMediaType, content, err := manifest.Payload()
	d.Require().Nil(err)
	art := &artifact.Artifact{ManifestMediaType: manifestMediaType}
	d.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(ormbConfig)), nil)
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Require().Nil(err)
	d.Len(art.ExtraAttrs, 12)
	d.Equal("CNN Model", art.ExtraAttrs["description"])
	d.Equal("TensorFlow", art.ExtraAttrs["framework"])
	d.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])

	// reset the mock
	d.SetupTest()
	art = &artifact.Artifact{ManifestMediaType: manifestMediaType}
	d.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(ormbConfigWithEmptySkipKeyList)), nil)
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Require().Nil(err)
	d.Len(art.ExtraAttrs, 14)
	d.Equal("CNN Model", art.ExtraAttrs["description"])
	d.Equal("TensorFlow", art.ExtraAttrs["framework"])
	d.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])

	// reset the mock
	d.SetupTest()
	art = &artifact.Artifact{ManifestMediaType: manifestMediaType}
	d.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(ormbConfigWithoutSkipKeyList)), nil)
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Require().Nil(err)
	d.Len(art.ExtraAttrs, 14)
	d.Equal("CNN Model", art.ExtraAttrs["description"])
	d.Equal("TensorFlow", art.ExtraAttrs["framework"])
	d.Equal([]interface{}{map[string]interface{}{"name": "batch_size", "value": "32"}}, art.ExtraAttrs["hyperparameters"])

	// reset the mock
	d.SetupTest()
	art = &artifact.Artifact{ManifestMediaType: manifestMediaType}
	d.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(ormbConfigUnsupportedSchemaVersion)), nil)
	err = d.processor.AbstractMetadata(nil, art, content)
	d.Equal(fmt.Errorf("unsupported artifact config schema version 2"), err)
}

func TestDefaultProcessorTestSuite(t *testing.T) {
	suite.Run(t, &defaultProcessorTestSuite{})
}
