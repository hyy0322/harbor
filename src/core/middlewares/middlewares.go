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

package middlewares

import (
	"net/http"
	"regexp"

	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/artifactinfo"
	"github.com/goharbor/harbor/src/server/middleware/csrf"
	"github.com/goharbor/harbor/src/server/middleware/log"
	"github.com/goharbor/harbor/src/server/middleware/mergeslash"
	"github.com/goharbor/harbor/src/server/middleware/metric"
	"github.com/goharbor/harbor/src/server/middleware/notification"
	"github.com/goharbor/harbor/src/server/middleware/orm"
	"github.com/goharbor/harbor/src/server/middleware/readonly"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"github.com/goharbor/harbor/src/server/middleware/security"
	"github.com/goharbor/harbor/src/server/middleware/session"
	"github.com/goharbor/harbor/src/server/middleware/trace"
	"github.com/goharbor/harbor/src/server/middleware/transaction"
	"github.com/goharbor/harbor/src/server/middleware/url"

	"github.com/beego/beego"
)

var (
	match         = regexp.MustCompile
	numericRegexp = match(`[0-9]+`)

	// The ping endpoint will be blocked when DB conns reach the max open conns of the sql.DB
	// which will make ping request timeout, so skip the middlewares which will require DB conn.
	pingSkipper = middleware.MethodAndPathSkipper(http.MethodGet, match("^/api/v2.0/ping"))

	// dbTxSkippers skip the transaction middleware for PATCH Blob Upload, PUT Blob Upload and `Read` APIs
	// because the APIs may take a long time to run, enable the transaction middleware in them will hold the database connections
	// until the API finished, this behavior may eat all the database connections.
	// There are no database writing operations in the PATCH Blob APIs, so skip the transaction middleware is all ok.
	// For the PUT Blob Upload API, we will make a transaction manually to write blob info to the database when put blob upload successfully.
	dbTxSkippers = []middleware.Skipper{
		middleware.MethodAndPathSkipper(http.MethodPatch, distribution.BlobUploadURLRegexp),
		middleware.MethodAndPathSkipper(http.MethodPut, distribution.BlobUploadURLRegexp),
		func(r *http.Request) bool { // skip tx for GET, HEAD and Options requests
			m := r.Method
			return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
		},
	}

	// readonlySkippers skip the post request when harbor sets to readonly.
	readonlySkippers = []middleware.Skipper{
		middleware.MethodAndPathSkipper(http.MethodPut, match("^/api/v2.0/configurations")),
		middleware.MethodAndPathSkipper(http.MethodPut, match("^/api/internal/configurations")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/login")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/userExists")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/oidc/onboard")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/adminjob/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/replication/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/replication/task/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/webhook/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/retention/task/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/schedules/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/webhook/"+numericRegexp.String())),
		pingSkipper,
	}

	retrySkippers = []middleware.Skipper{
		middleware.MethodAndPathSkipper(http.MethodGet, match("^/api/.*")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/api/.*")),
		middleware.MethodAndPathSkipper(http.MethodPut, match("^/api/.*")),
		middleware.MethodAndPathSkipper(http.MethodDelete, match("^/api/.*")),

		middleware.MethodAndPathSkipper(http.MethodGet, match("^/c/.*")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/.*")),
		middleware.MethodAndPathSkipper(http.MethodPut, match("^/c/.*")),
		middleware.MethodAndPathSkipper(http.MethodDelete, match("^/c/.*")),

		middleware.MethodAndPathSkipper(http.MethodGet, match("^/service/.*")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/.*")),
		middleware.MethodAndPathSkipper(http.MethodPut, match("^/service/.*")),
		middleware.MethodAndPathSkipper(http.MethodDelete, match("^/service/.*")),

		middleware.MethodAndPathSkipper(http.MethodPatch, distribution.BlobUploadURLRegexp),
		middleware.MethodAndPathSkipper(http.MethodPut, distribution.BlobUploadURLRegexp),
	}
)

// MiddleWares returns global middlewares
func MiddleWares() []beego.MiddleWare {
	return []beego.MiddleWare{
		url.Middleware(),
		mergeslash.Middleware(),
		trace.Middleware(),
		metric.Middleware(),
		requestid.Middleware(),
		log.Middleware(),
		session.Middleware(),
		csrf.Middleware(),
		middleware.RetryMiddleware(retrySkippers...),
		orm.Middleware(pingSkipper),
		notification.Middleware(pingSkipper), // notification must ahead of transaction ensure the DB transaction execution complete
		transaction.Middleware(dbTxSkippers...),
		artifactinfo.Middleware(),
		security.Middleware(pingSkipper),
		security.UnauthorizedMiddleware(),
		readonly.Middleware(readonlySkippers...),
	}
}
