package middleware

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server"
	"xrf197ilz35aq/internal/server/api/response"
)

type AuthenticationMiddleware struct {
	logger        slog.Logger
	authProcessor processor.AuthProcessor
}

func (m *AuthenticationMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if m.shouldCheckRouteAuth(r) {
			authToken := r.Header.Get(internal.XrfAuthToken)
			if authToken == "" {
				externalErr := &internal.ExternalError{Message: "invalid auth token", Code: 401}
				response.WriteErrorResponse(externalErr, w, m.logger)
				return
			}

			req := model.VerifyRevokeTokenReq{Token: authToken}
			userCtx, err := m.authProcessor.ValidateAuthToken(r.Context(), m.logger, req)
			if err != nil {
				externalErr := &internal.ExternalError{Message: err.Error(), Code: 401}
				response.WriteErrorResponse(externalErr, w, m.logger)
				return
			}
			if userCtx == nil {
				externalErr := &internal.ExternalError{Message: "invalid auth token", Code: 401}
				response.WriteErrorResponse(externalErr, w, m.logger)
				return
			}
			// set context to the enriched context with the user context obj
			ctx = server.ContextWithUserCtx(r.Context(), *userCtx)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthenticationMiddleware) shouldCheckRouteAuth(r *http.Request) bool {
	unCheckedRoutes := make(map[string]string)
	unCheckedRoutes["/health"] = "ANY"
	unCheckedRoutes["/api/v1/user"] = "POST"
	unCheckedRoutes["/api/v1/auth"] = "POST"
	unCheckedRoutes["/api/v1/auth/token"] = "POST"

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
