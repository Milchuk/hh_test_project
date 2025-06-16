package handler

import (
	"encoding/json"
	"hh_test_project/internal/models"
	"hh_test_project/internal/service"
	"net/http"
	"strconv"
)

type Handler struct {
	Service *service.GoodsService
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createGoods(w, r)
	case http.MethodGet:
		h.getGoods(w, r)
	case http.MethodPatch:
		h.updateGoods(w, r)
	case http.MethodDelete:
		h.deleteGoods(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getGoods(w http.ResponseWriter, r *http.Request) {
	goodsID, _ := strconv.Atoi(r.URL.Path[len("/goods/"):])
	projectID, _ := strconv.Atoi(r.URL.Query().Get("projectId"))

	goods, err := h.Service.GetByID(r.Context(), projectID, goodsID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, goods)
}

func (h *Handler) createGoods(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil || projectID <= 0 {
		respondError(w, err)
		return
	}

	var input models.GoodsCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, err)
		return
	}

	goods, err := h.Service.Create(r.Context(), projectID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, goods)
}

func (h *Handler) updateGoods(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil || projectID <= 0 {
		respondError(w, err)
		return
	}

	goodsID, _ := strconv.Atoi(r.URL.Path[len("/goods/"):])

	var input models.GoodsUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, err)
		return
	}

	goods, err := h.Service.Update(r.Context(), projectID, goodsID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, goods)
}

func (h *Handler) deleteGoods(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil || projectID <= 0 {
		respondError(w, err)
		return
	}

	goodsID, _ := strconv.Atoi(r.URL.Path[len("/goods/"):])

	if err := h.Service.Delete(r.Context(), projectID, goodsID); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":        goodsID,
		"projectId": projectID,
		"removed":   true,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err error) {
	switch err {
	case service.ErrNotFound:
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"code":    3,
			"message": "errors.common.notFound",
			"details": map[string]interface{}{},
		})
	default:
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
}
