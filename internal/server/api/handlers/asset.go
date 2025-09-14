package handlers

import (
	"log/slog"
	"net/http"
)

type assetHandler struct {
	defaultLogger slog.Logger
}

func (ah *assetHandler) RegisterRoutes(serveMux *http.ServeMux) {}

func NewAssetHandler(defaultLogger slog.Logger) RequestHandler {
	return &assetHandler{defaultLogger}
}
