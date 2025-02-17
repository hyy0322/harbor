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

package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/distribution"
)

// ParseProjectName parse project name from v2 and v2.0 API URL path
func ParseProjectName(r *http.Request) string {
	path := path.Clean(r.URL.EscapedPath())

	var projectName string

	prefixes := []string{
		fmt.Sprintf("/api/%s/projects/", api.APIVersion), // v2.0 management APIs
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
			if len(parts) > 0 {
				projectName = parts[0]
				break
			}
		}
	}

	if projectName == "" && strings.HasPrefix(path, "/v2/") {
		// v2 APIs
		projectName = distribution.ParseProjectName(path)
	}

	return projectName
}

// SkipPolicyChecking ...
func SkipPolicyChecking(r *http.Request, projectID, artID int64) (bool, error) {
	secCtx, ok := security.FromContext(r.Context())

	// 1, scanner pull access can bypass.
	// 2, cosign pull can bypass, it needs to pull the manifest before pushing the signature.
	// 3, pull cosign signature can bypass.
	if ok && secCtx.Name() == "v2token" {
		if secCtx.Can(r.Context(), rbac.ActionScannerPull, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) ||
			(secCtx.Can(r.Context(), rbac.ActionPush, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) &&
				strings.Contains(r.UserAgent(), "cosign")) {
			return true, nil
		}
	}

	accs, err := accessory.Mgr.List(r.Context(), q.New(q.KeyWords{"ArtifactID": artID}))
	if err != nil {
		return false, err
	}
	if len(accs) > 0 && accs[0].GetData().Type == model.TypeCosignSignature {
		return true, nil
	}

	return false, nil
}

var errInvalidSecret = fmt.Errorf("invalid secret")

// UnpackUploadState unpacks and validates the blob upload state from the
// token, using the hmacKey secret.
// 如果不挂载secret，因为mac长度就是32，所以只需要获取32:后面的数据即可，不做数据校验
func UnpackUploadState(secret string, token string) (blobUploadState, error) {
	var state blobUploadState

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}
	mac := hmac.New(sha256.New, []byte(secret))

	if len(tokenBytes) < mac.Size() {
		return state, fmt.Errorf("token too short: %d,%d", len(tokenBytes), mac.Size())
	}

	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, fmt.Errorf("token invalid: %x, %x", mac.Sum(nil), macBytes)
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}

// blobUploadState captures the state serializable state of the blob upload.
type blobUploadState struct {
	// name is the primary repository under which the blob will be linked.
	Name string

	// UUID identifies the upload.
	UUID string

	// offset contains the current progress of the upload.
	Offset int64

	// StartedAt is the original start time of the upload.
	StartedAt time.Time
}
