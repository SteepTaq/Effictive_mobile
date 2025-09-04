package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"rest-service/internal/model"
	"rest-service/pkg/response"
)

type Repository interface {
	Create(ctx context.Context, s model.Subscription) (model.Subscription, error)
	Get(ctx context.Context, id int64) (model.Subscription, error)
	Update(ctx context.Context, id int64, s model.Subscription) (model.Subscription, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int, userID uuid.UUID, serviceName string) ([]model.Subscription, error)
	SumByPeriod(ctx context.Context, from, to time.Time, userID uuid.UUID, serviceName string) (int64, error)
}

type Handler struct {
	repo Repository
	log  *slog.Logger
}

func NewHandler(repo Repository, log *slog.Logger) *Handler {
	return &Handler{
		repo: repo,
		log:  log,
	}
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	h.log.Info("create subscription request")
	var in struct {
		ServiceName string    `json:"service_name"`
		Price       int64     `json:"price"`
		UserID      uuid.UUID `json:"user_id"`
		StartDate   string    `json:"start_date"`
		EndDate     *string   `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed to decode request", "error", err)
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	start, err := time.Parse("01-2006", in.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "start_date", in.StartDate, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid start_date, expected MM-YYYY")
		return
	}

	var end *time.Time
	if in.EndDate != nil {
		t, err := time.Parse("01-2006", *in.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "end_date", *in.EndDate, "error", err)
			response.Error(w, http.StatusBadRequest, "invalid end_date, expected MM-YYYY")
			return
		}
		end = &t
	}

	s, err := h.repo.Create(r.Context(), model.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	})
	if err != nil {
		h.log.Error("failed to create subscription", "error", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Json(w, http.StatusCreated, s)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	h.log.Info("get subscription request")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("invalid id", "id", idStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	s, err := h.repo.Get(r.Context(), id)
	if err != nil {
		h.log.Error("subscription not found", "id", id, "error", err)
		response.Error(w, http.StatusNotFound, "not found")
		return
	}

	response.Json(w, http.StatusOK, s)
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	h.log.Info("update subscription request")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("invalid id", "id", idStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var in struct {
		ServiceName string    `json:"service_name"`
		Price       int64     `json:"price"`
		UserID      uuid.UUID `json:"user_id"`
		StartDate   string    `json:"start_date"`
		EndDate     *string   `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed to decode request", "error", err)
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	start, err := time.Parse("01-2006", in.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "start_date", in.StartDate, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid start_date, expected MM-YYYY")
		return
	}

	var end *time.Time
	if in.EndDate != nil {
		t, err := time.Parse("01-2006", *in.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "end_date", *in.EndDate, "error", err)
			response.Error(w, http.StatusBadRequest, "invalid end_date, expected MM-YYYY")
			return
		}
		end = &t
	}

	s, err := h.repo.Update(r.Context(), id, model.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartDate:   start,
		EndDate:     end,
	})
	if err != nil {
		h.log.Error("failed to update subscription", "id", id, "error", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Json(w, http.StatusOK, s)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	h.log.Info("delete subscription request")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("invalid id", "id", idStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.log.Error("failed to delete subscription", "id", id, "error", err)
		response.Error(w, http.StatusNotFound, "not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	h.log.Info("list subscriptions request")
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		response.Error(w, http.StatusBadRequest, "user_id required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("invalid user_id", "user_id", userIDStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		response.Error(w, http.StatusBadRequest, "service_name required")
		return
	}

	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			limit = l
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if o, err := strconv.Atoi(v); err == nil {
			offset = o
		}
	}

	list, err := h.repo.List(r.Context(), limit, offset, userID, serviceName)
	if err != nil {
		h.log.Error("failed to list subscriptions", "error", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Json(w, http.StatusOK, list)
}

func (h *Handler) SummarySubscriptions(w http.ResponseWriter, r *http.Request) {
	h.log.Info("summary subscriptions request")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		response.Error(w, http.StatusBadRequest, "from and to are required (MM-YYYY)")
		return
	}

	from, err := time.Parse("01-2006", fromStr)
	if err != nil {
		h.log.Error("invalid from date", "from", fromStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid from date format, expected MM-YYYY")
		return
	}

	to, err := time.Parse("01-2006", toStr)
	if err != nil {
		h.log.Error("invalid to date", "to", toStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid to date format, expected MM-YYYY")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		response.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("invalid user_id", "user_id", userIDStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid user_id format")
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		response.Error(w, http.StatusBadRequest, "service_name is required")
		return
	}

	h.log.Info("calculating subscription sum",
		"user_id", userID,
		"service_name", serviceName,
		"from", fromStr,
		"to", toStr)

	total, err := h.repo.SumByPeriod(r.Context(), from, to, userID, serviceName)
	if err != nil {
		h.log.Error("failed to calculate sum", "error", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Json(w, http.StatusOK, map[string]interface{}{
		"total":        total,
		"user_id":      userID,
		"service_name": serviceName,
		"period": map[string]string{
			"from": fromStr,
			"to":   toStr,
		},
	})
}
