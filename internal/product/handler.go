package product

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

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "productID"), 10, 64)
	if err != nil || productID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid_product_id")
		return
	}
	response, err := h.service.GetProduct(r.Context(), productID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "product_not_found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "product_read_failed")
		return
	}
	httpx.JSON(w, http.StatusOK, response)
}
