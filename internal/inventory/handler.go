package inventory

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"inventory-cache-lab/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}
	response, err := h.service.GetInventory(r.Context(), productID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "inventory_not_found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "inventory_read_failed")
		return
	}
	httpx.JSON(w, http.StatusOK, response)
}

func (h *Handler) InvalidateInventory(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseProductID(w, r)
	if !ok {
		return
	}
	if err := h.service.InvalidateInventory(r.Context(), productID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "inventory_invalidation_failed")
		return
	}
	httpx.JSON(w, http.StatusOK, InvalidateResponse{ProductID: productID, Invalidated: true})
}

func parseProductID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "productID"), 10, 64)
	if err != nil || productID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid_product_id")
		return 0, false
	}
	return productID, true
}
