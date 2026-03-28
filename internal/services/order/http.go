package order

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"

    "github.com/arda/distributed-commerce-core/internal/platform/httpx"
)

type Handler struct {
    Service *Service
    Repo    *Repository
}

type createOrderRequest struct {
    CustomerID string `json:"customer_id"`
    Currency   string `json:"currency"`
    Items      []Item `json:"items"`
}

func (h Handler) Routes(r chi.Router) {
    r.Post("/v1/orders", h.create)
    r.Get("/v1/orders/{id}", h.get)
}

func (h Handler) create(w http.ResponseWriter, r *http.Request) {
    var req createOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    order, err := h.Service.Create(r.Context(), req.CustomerID, req.Currency, httpx.CorrelationID(r.Context()), req.Items)
    if err != nil {
        httpx.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    httpx.JSON(w, http.StatusAccepted, order)
}

func (h Handler) get(w http.ResponseWriter, r *http.Request) {
    o, err := h.Repo.Get(r.Context(), chi.URLParam(r, "id"))
    if err != nil {
        httpx.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
        return
    }
    httpx.JSON(w, http.StatusOK, o)
}
