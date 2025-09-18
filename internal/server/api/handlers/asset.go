package handlers

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/server"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type assetHandler struct {
	defaultLogger slog.Logger
}

func (ah *assetHandler) createAsset(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	var req model.AssetRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}
}

func (ah *assetHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/asset", ah.createAsset)
}

func NewAssetHandler(defaultLogger slog.Logger) RequestHandler {
	return &assetHandler{defaultLogger}
}
