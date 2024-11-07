package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/segmentio/ksuid"
)

type Prefix string

const (
	UserPrefix    Prefix = "user_"
	ShiftPrefix   Prefix = "shift_"
	InvoicePrefix Prefix = "invoice_"
)

type User struct {
	ID          string `json:"id" db:"id"`
	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	Email       string `json:"email" db:"email"`
	CompanyName string `json:"company_name" db:"company_name"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
}

type Shift struct {
	ID               string    `json:"id" db:"id"`
	StartDate        time.Time `json:"start_date" db:"start_date"`
	EndDate          time.Time `json:"end_date" db:"end_date"`
	Location         string    `json:"location" db:"location"`
	ShiftName        string    `json:"shift_name" db:"shift_name"`
	ShiftsFilled     time.Time `json:"shifts_filled" db:"shifts_filled"`
	ShiftDescription string    `json:"shift_description" db:"shift_description"`
	CreatedBy        string    `json:"created_by" db:"created_by"`
	UpdatedBy        string    `json:"updated_by" db:"updated_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Invoice struct {
	ID            string    `json:"id" db:"id"`
	StartDate     time.Time `json:"start_date" db:"start_date"`
	EndDate       time.Time `json:"end_date" db:"end_date"`
	InvoiceAmount float64   `json:"invoice_amount" db:"invoice_amount"`
	Status        string    `json:"status" db:"status"`
	UserID        string    `json:"user_id" db:"user_id"`
	ShiftID       string    `json:"shift_id" db:"shift_id"`
	CreatedBy     string    `json:"created_by" db:"created_by"`
	UpdatedBy     string    `json:"updated_by" db:"updated_by"`
	InvoiceName   string    `json:"invoice_name" db:"invoice_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type InvoiceResponse struct {
	ID            string    `json:"id" db:"id"`
	StartDate     time.Time `json:"start_date" db:"start_date"`
	EndDate       time.Time `json:"end_date" db:"end_date"`
	InvoiceAmount float64   `json:"invoice_amount" db:"invoice_amount"`
	Status        string    `json:"status" db:"status"`
	InvoiceName   string    `json:"invoice_name" db:"invoice_name"`
}

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}

type Service interface {
	FetchInvoices(ctx context.Context, userId string, searchTerm string) ([]InvoiceResponse, error)
	CreateUser(ctx context.Context, user *User) (string, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
}

type service struct {
	db *sql.DB
}

var _ Service = &service{}

func (s *service) CreateUser(ctx context.Context, user *User) (string, error) {
	// create new ksuid for user
	fmt.Println("creating user")
	userID := generateID(UserPrefix)
	// todo: create onboarding flow to collect the following user information: first name, last name, email, phone_number

	_, err := s.db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, phone_number, company_name, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.CompanyName, userID)
	if err != nil {
		return "", fmt.Errorf("error creating user: %w", err)
	}

	err = s.initializeData(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("error initializing data: %w", err)
	}

	return userID, nil
}

func (s *service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
	               SELECT id, first_name, last_name, phone_number
	               FROM users 
	               WHERE id = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("error fetching user with id %s: %w", userID, err)
	}

	return &user, nil
}

func (s *service) FetchInvoices(ctx context.Context, userId string, searchTerm string) ([]InvoiceResponse, error) {
	var query string
	var args []interface{}
	var invoices []InvoiceResponse

	// Base query
	query = `
		SELECT 
			i.id,
			i.invoice_amount,
			s.start_date,
			s.end_date,
			i.status,
			i.invoice_name
		FROM invoices i
		JOIN shifts s ON i.shift_id = s.id
		WHERE i.created_by = $1
		AND s.created_by = $1` // Removed the semicolon here
	args = append(args, userId)

	// If search term is provided, add it to the query
	if searchTerm != "" {
		query += ` AND invoice_name ILIKE $2`
		args = append(args, "%"+searchTerm+"%")
	}

	// Add the semicolon at the very end if needed
	query += `;`

	// Execute the query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying invoices: %w", err)
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var inv InvoiceResponse
		err := rows.Scan(
			&inv.ID,
			&inv.InvoiceAmount,
			&inv.StartDate,
			&inv.EndDate,
			&inv.Status,
			&inv.InvoiceName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning invoice row: %w", err)
		}
		// inv.InvoiceAmount = float64(amountCents) / 100 // Convert cents to dollars
		invoices = append(invoices, inv)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, nil
}

func (s *service) initializeData(ctx context.Context, employerID string) error {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create a worker user
	userID := generateID(UserPrefix)

	workerEmail := fmt.Sprintf("john.doe+%s@example.com", ksuid.New().String())
	_, err = tx.Exec(`
		INSERT INTO users (id, first_name, last_name, email, phone_number, company_name, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, "John", "Doe", workerEmail, "1234567890", "", userID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Create a shift
	shiftID := generateID(ShiftPrefix)
	_, err = tx.Exec(`
		INSERT INTO shifts (id, worker_id, start_date, end_date, location, shift_name, shifts_filled, shift_description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, shiftID, userID, time.Now(), time.Now().AddDate(0, 0, 7), "Main Street", "Day Shift", 4, "Regular day shift", employerID)
	if err != nil {
		return fmt.Errorf("failed to insert shift: %w", err)
	}

	// Create 10 invoices
	err = generateInvoices(tx, shiftID, employerID)
	if err != nil {
		return fmt.Errorf("failed to generate invoices: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "Data initialization completed successfully")
	return nil
}

func generateID(prefix Prefix) string {
	return fmt.Sprintf("%s%s", prefix, ksuid.New().String())
}

func generateInvoices(tx *sql.Tx, shiftID, employerID string) error {
	shiftNames := []string{
		"Morning Shift",
		"Afternoon Shift",
		"Night Shift",
		"Weekend Shift",
		"Holiday Shift",
		"Emergency Shift",
		"Overtime Shift",
		"On-Call Shift",
		"Training Shift",
		"Special Event Shift",
	}

	for i := 0; i < 10; i++ {
		invoiceID := generateID(InvoicePrefix)
		randomShiftName := shiftNames[rand.Intn(len(shiftNames))]
		randomAmount := rand.Intn(90001) + 10000 // Random number between 10000 and 100000
		status := "paid"
		if i%3 == 0 {
			status = "unpaid"
		}

		_, err := tx.Exec(`
			INSERT INTO invoices (id, invoice_amount, status, shift_id, invoice_name, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, invoiceID, randomAmount, status, shiftID, randomShiftName, employerID)

		if err != nil {
			return fmt.Errorf("failed to insert invoice %d: %w", i+1, err)
		}
	}

	return nil
}
