package handler


import (
	"encoding/json"
	"log/slog"
	"net/http"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
)


type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h* Handler) HandleCreateUser (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	userId := r.Context().Value("user_id").(string)
	// Get the user from the request
	var user service.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		slog.Error("failed to decode user", "error", err)
		http.Error(w, "failed to decode user", http.StatusBadRequest)
		return
	}
	// Create the user
	err = h.svc.CreateUser(userId, user)
	if err != nil {
		slog.Error("failed to create user", "error", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	// Return success
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) HandleFetchInvoices (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	userId := r.Context().Value("user_id").(string)
	// Get the invoices from the service

	searchTerm := r.URL.Query().Get("search")

	invoices, err := h.svc.FetchInvoices(userId, searchTerm)
	if err != nil {
		slog.Error("failed to fetch invoices", "error", err)
		http.Error(w, "failed to fetch invoices", http.StatusInternalServerError)
		return
	}
	// Return the invoices
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}