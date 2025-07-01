package handlers

import "net/http"

type RequestHandler interface {
	RegisterRoutes(serveMux *http.ServeMux)
}
