package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
    "github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/config"
)

type role string 


const EMPLOYER role = "rol_lz7KugKHb6tiTJVl"

type Handler struct {
	svc service.Service
    cfg *config.Config
}

func NewHandler(svc service.Service, cfg *config.Config) *Handler {
	return &Handler{
        svc: svc,
        cfg: cfg,
    }
}

type CreateUserReq struct {
	User  service.User `json:"user"`
	Secret string `json:"secret"`
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
    token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
    customClaims := token.CustomClaims.(*middleware.CustomClaims)
    if customClaims.Roles[0] != string(h.cfg.Auth0RoleID) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // todo update this by getting the id 
    user, err := h.svc.GetUserByID(r.Context(), customClaims.DBUserId)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        slog.Error("failed to get user", "error", err)
        http.Error(w, "failed to get user", http.StatusInternalServerError)
        return
    }

    sendJSONResponse(w, http.StatusOK, user)
}

func (h* Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
    // Get the user from the context
    slog.Info("creating user")
    var reqBody CreateUserReq
    err := json.NewDecoder(r.Body).Decode(&reqBody)
    if err != nil {
        slog.Error("failed to decode user", "error", err)
        http.Error(w, "failed to decode user", http.StatusBadRequest)
        return
    }

    slog.Info("secret", "req.Secret", reqBody.Secret)
    slog.Info("secret","cfg.Auth0HookSecret", h.cfg.Auth0HookSecret)

    	// 2. Validate secret
	if reqBody.Secret != h.cfg.Auth0HookSecret {
		sendJSONResponse(w, http.StatusForbidden, "You must provide the secret")
		return
	}

    // Create the user
    userID, err := h.svc.CreateUser(r.Context(), &reqBody.User)
    if err != nil {
        slog.Error("failed to create user", "error", err)
        http.Error(w, "failed to create user", http.StatusInternalServerError)
        return
    }
    
    type response struct {
        UserID string `json:"user_id"`
    }

    payload := &response{
        UserID: userID,
    }

    // Return the user
    sendJSONResponse(w, http.StatusOK, payload)
}

func (h *Handler) HandleFetchInvoices (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
    slog.Info("fetching invoices")
    token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
    customClaims := token.CustomClaims.(*middleware.CustomClaims)

    slog.Info("fetching invoices", "user_id", customClaims)
    if customClaims.Roles[0] != string(EMPLOYER) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
	searchTerm := r.URL.Query().Get("search")

	invoices, err := h.svc.FetchInvoices(r.Context(), customClaims.DBUserId, searchTerm)
	if err != nil {
		slog.Error("failed to fetch invoices", "error", err)
		http.Error(w, "failed to fetch invoices", http.StatusInternalServerError)
		return
	}

	// Return the invoices
    sendJSONResponse(w, http.StatusOK, invoices)
}

func sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}