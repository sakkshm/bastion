package api

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	healthCheckResult := h.Engine.HealthCheck()

	resp := HealthCheckResponse{
		DockerAlive: healthCheckResult.DockerHealthy,
		DBAlive:     healthCheckResult.DatabaseHealthy,
		APIAlive:    true,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}
}
