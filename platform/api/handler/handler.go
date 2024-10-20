package handler


import (
	"encoding/json"
	"net/http"
	"github.com/z3r0-cool/monorepo/platform/api/service"
)


type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleFetchInvoices (w http.ResponseWriter, r *http.Request) {
	// Get the user from the context
	user := r.Context().Value("user").(string)
	// Get the invoices from the service

	searchTerm := r.URL.Query().Get("search")

	invoices, err := svc.FetchInvoices(user)
	if err != nil {
		slog.Error("failed to fetch invoices", "error", err)
		http.Error(w, "failed to fetch invoices", http.StatusInternalServerError)
		return
	}
	// Return the invoices
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}