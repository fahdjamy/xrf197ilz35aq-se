package middleware

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api/response"
)

type AuthenticationMiddleware struct {
	logger        slog.Logger
	authProcessor processor.AuthProcessor
}

func (m *AuthenticationMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.shouldCheckRouteAuth(r) {
			authToken := r.Header.Get(internal.XrfAuthToken)
			if authToken == "" {
				externalErr := &internal.ExternalError{Message: "invalid/missing x-rf-se-token", Code: 401}
				response.WriteErrorResponse(externalErr, w, m.logger)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthenticationMiddleware) shouldCheckRouteAuth(r *http.Request) bool {
	unCheckedRoutes := make(map[string]string)
	unCheckedRoutes["/health"] = "ANY"
	unCheckedRoutes["/api/v1/user"] = "POST"
	unCheckedRoutes["/api/v1/auth"] = "POST"

	route := r.URL.Path

	method, ok := unCheckedRoutes[route]

	if !ok || (method == "" || method != "ANY" && method != r.Method) {
		return true
	}

	return false
}

func NewAuthenticationMiddleware(logger slog.Logger, authProcessor processor.AuthProcessor) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		logger:        logger,
		authProcessor: authProcessor,
	}
}
