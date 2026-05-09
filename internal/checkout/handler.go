package checkout

import (
	"encoding/json"
	"errors"
	"net/http"

	"inventory-cache-lab/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_json")
		return
	}
	response, err := h.service.PlaceOrder(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRequest):
			httpx.Error(w, http.StatusBadRequest, "invalid_checkout_request")
		case errors.Is(err, ErrInsufficientInventory):
			httpx.Error(w, http.StatusConflict, "insufficient_inventory")
		default:
			httpx.Error(w, http.StatusInternalServerError, "checkout_failed")
		}
		return
	}
	httpx.JSON(w, http.StatusOK, response)
}
