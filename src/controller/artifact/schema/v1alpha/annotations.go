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

package v1alpha

const (
	// AnnotationArtifactSkipList is the annotation key for user-defined artifact in manifest JSON config.annotations to specify which keys in artifact config JSON will not be extracted out to artifact model stored in database.
	// Values for this key should be type string separated by comma.
	// example:
	// "org.goharbor.artifact.skiplist": "keyA,keyB"
	AnnotationArtifactSkipList = "org.goharbor.artifact.skiplist"

	// AnnotationArtifactSkipList is the annotation key for user-defined artifact to specify layers in manifest JSON layers[].annotations to specify if this layer is a icon layer or not.
	// Values for this key should be type string, true or false
	// example:
	// "org.goharbor.artifact.icon": "true"
	AnnotationArtifactIcon = "org.goharbor.artifact.icon"
)
