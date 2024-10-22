package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"database/sql"

	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
)


type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleFetchUser(w http.ResponseWriter, r *http.Request) {
    customClaims, ok := r.Context().Value("user").(*middleware.CustomClaims)
    if !ok {
        http.Error(w, "Failed to get user claims", http.StatusInternalServerError)
        return
    }

    user, err := h.svc.GetUserByEmail(r.Context(), customClaims.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            // User doesn't exist, create a new one
            newUser := service.User{
                Email:       customClaims.Email,
                FirstName:   customClaims.UserMetadata.FirstName,
                LastName:    customClaims.UserMetadata.LastName,
                CompanyName: customClaims.UserMetadata.CompanyName,
                PhoneNumber: customClaims.UserMetadata.PhoneNumber,
            }

            err := h.svc.CreateUser(r.Context(), &newUser)
            if err != nil {
                http.Error(w, "Failed to create user", http.StatusInternalServerError)
                return
            }

            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(newUser)
        } else {
            http.Error(w, "Failed to get user", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) HandleFetchInvoices (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	userId := r.Context().Value("user_id").(string)
	// Get the invoices from the service

	searchTerm := r.URL.Query().Get("search")

	invoices, err := h.svc.FetchInvoices(r.Context(), userId, searchTerm)
	if err != nil {
		slog.Error("failed to fetch invoices", "error", err)
		http.Error(w, "failed to fetch invoices", http.StatusInternalServerError)
		return
	}
	// Return the invoices
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}