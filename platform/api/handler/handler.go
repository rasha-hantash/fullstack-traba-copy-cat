package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
    "fmt"
	"net/http"
    "os"

    "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
)


type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

type CreateUserReq struct {
	User  service.User `json:"user"`
	Secret string `json:"secret"`
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("Fetching user")
    token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)

    customClaims := token.CustomClaims.(*middleware.CustomClaims)

    user, err := h.svc.GetUserByEmail(r.Context(), customClaims.Email)
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
    var reqBody CreateUserReq
    err := json.NewDecoder(r.Body).Decode(&reqBody)
    if err != nil {
        slog.Error("failed to decode user", "error", err)
        http.Error(w, "failed to decode user", http.StatusBadRequest)
        return
    }

    	// 2. Validate secret
	if reqBody.Secret != os.Getenv("AUTH0_HOOK_SECRET") {
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

    // Return the user
    sendJSONResponse(w, http.StatusOK, fmt.Sprintf(`{"user_id":"%s"}`, userID))
}

func (h *Handler) HandleFetchInvoices (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	userId := r.Context().Value("user_id").(string)
	searchTerm := r.URL.Query().Get("search")

	invoices, err := h.svc.FetchInvoices(r.Context(), userId, searchTerm)
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