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

package middleware

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor/src/lib/log"

	"github.com/google/uuid"
)

type requestIDKey struct{}

const (
	// HeaderXRequestID is the standard X-Request-ID header
	HeaderXRequestID = "X-Request-ID"
	// HeaderXTtLogID is the X-TT-LOGID header for unified data plane
	HeaderXTtLogID = "X-TT-LOGID"
)

// RequestID returns a middleware that extracts the request ID from headers
// (X-TT-LOGID first, then X-Request-ID, then generates a UUID), sets both
// headers on the response, and injects the ID into context and logger.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := ResolveRequestID(r)
			setResponseHeaders(w, rid)
			ctx := injectIntoContext(r.Context(), rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ResolveRequestID extracts the request ID from headers with X-TT-LOGID priority.
// Falls back to generating a UUID if both headers are empty.
func ResolveRequestID(r *http.Request) string {
	rid := r.Header.Get(HeaderXTtLogID)
	if rid == "" {
		rid = r.Header.Get(HeaderXRequestID)
	}
	if rid == "" {
		rid = uuid.New().String()
	}
	return rid
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
		return rid
	}
	return ""
}

// SetRequestIDHeaders sets both X-Request-ID and X-TT-LOGID on the given
// http.Header for outbound requests from registryctl.
func SetRequestIDHeaders(header http.Header, rid string) {
	header.Set(HeaderXRequestID, rid)
	header.Set(HeaderXTtLogID, rid)
}

func setResponseHeaders(w http.ResponseWriter, rid string) {
	w.Header().Set(HeaderXRequestID, rid)
	w.Header().Set(HeaderXTtLogID, rid)
}

func injectIntoContext(ctx context.Context, rid string) context.Context {
	ctx = context.WithValue(ctx, requestIDKey{}, rid)
	logger := log.GetLogger(ctx).WithFields(log.Fields{"requestID": rid})
	return log.WithLogger(ctx, logger)
}