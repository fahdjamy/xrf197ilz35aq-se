package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type assetHandler struct {
	defaultLogger  slog.Logger
	assetProcessor processor.AssetProcessor
}

func (ah *assetHandler) createAsset(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	var req model.AssetRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(errors.New("invalid user context"), w, *logger)
		return
	}
}

func (ah *assetHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/asset", ah.createAsset)
}

func NewAssetHandler(defaultLogger slog.Logger, assetProcessor processor.AssetProcessor) RequestHandler {
	return &assetHandler{defaultLogger, assetProcessor}
}
