package service

import (
	"database/sql"

)

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}


type Service interface {
	FetchInvoices(userId string, searchTerm string) ([]Invoice, error)
}

type service struct {
	db *sql.DB
}

var _ Service = &service{}

type Invoice struct {
	ID          string `json:"id"`	
	Amount      int    `json:"amount"`
	Description string `json:"description"`
}

func (s *service) FetchInvoices(userId string, searchTerm string) ([]Invoice, error) {
	// todo implement the logic to fetch invoices
	return nil, nil
}

