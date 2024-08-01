package blob

import (
	"context"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

func ParseBlobInstance() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		project := distribution.ParseProjectName(r.URL.Path)
		parts := strings.Split(project, "__")
		instance := "0"
		if len(parts) >= 2 {
			instance = parts[0]
		}

		ctx := context.WithValue(r.Context(), "instance", instance)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
