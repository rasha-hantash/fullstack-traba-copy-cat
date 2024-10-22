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
	UserPrefix Prefix = "user_"
	ShiftPrefix Prefix = "shift_"
	InvoicePrefix Prefix = "invoice_"
)

type User struct {
    ID          string    `json:"id" db:"id"`
    FirstName   string    `json:"first_name" db:"first_name"`
    LastName    string    `json:"last_name" db:"last_name"`
    Email       string    `json:"email" db:"email"`
	CompanyName string    `json:"company_name" db:"company_name"`
    PhoneNumber string    `json:"phone_number" db:"phone_number"`
    Role        string    `json:"role" db:"role"`
    CreatedBy   string    `json:"created_by" db:"created_by"`
    UpdatedBy   string    `json:"updated_by" db:"updated_by"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
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
    InvoiceAmount float64      `json:"invoice_amount" db:"invoice_amount"`
    Status        string    `json:"status" db:"status"`
    UserID        string    `json:"user_id" db:"user_id"`
    ShiftID       string    `json:"shift_id" db:"shift_id"`
    CreatedBy     string    `json:"created_by" db:"created_by"`
    UpdatedBy     string    `json:"updated_by" db:"updated_by"`
    InvoiceName   string    `json:"invoice_name" db:"invoice_name"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}


type Service interface {
	FetchInvoices(ctx context.Context, userId string, searchTerm string) ([]Invoice, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}

type service struct {
	db *sql.DB
}

var _ Service = &service{}


func (s *service) CreateUser(ctx context.Context, user *User) error {
	// create new ksuid for user 
	userID := generateID(UserPrefix)
	// todo: create onboarding flow to collect the following user information: first name, last name, email, phone_number
	
	_, err := s.db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, phone_number, role, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, userID, "Rasha", "Hantash", user.Email, "571-226-7109", user.Role, userID)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	
	err = s.initializeData(userID)
	if err != nil {
		return fmt.Errorf("error initializing data: %w", err)
	}

	return nil
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	
	user.Email = email
	err := s.db.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, phone_number
		FROM users 
		WHERE email = $1`,
		email,
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
		return nil, fmt.Errorf("error fetching user with email %s: %w", email, err)
	}

	return &user, nil
}


func (s *service) FetchInvoices(ctx context.Context, userId string, searchTerm string) ([]Invoice, error) {
    var query string
    var args []interface{}
    var invoices []Invoice

    // Base query
    query = `
        SELECT id, start_date, end_date, invoice_amount_cents, status, user_id, shift_id, invoice_name, created_at, updated_at
        FROM invoices
        WHERE user_id = $1
    `
    args = append(args, userId)

    // If search term is provided, add it to the query
    if searchTerm != "" {
        query += " AND invoice_name ILIKE $2"
        args = append(args, "%"+searchTerm+"%")
    }

    // Execute the query
    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("error querying invoices: %w", err)
    }
    defer rows.Close()

    // Iterate over the rows
    for rows.Next() {
        var inv Invoice
        var amountCents int
        err := rows.Scan(
            &inv.ID,
            &inv.StartDate,
            &inv.EndDate,
            &amountCents,
            &inv.Status,
            &inv.UserID,
            &inv.ShiftID,
            &inv.InvoiceName,
            &inv.CreatedAt,
            &inv.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning invoice row: %w", err)
        }
        inv.InvoiceAmount = float64(amountCents) / 100 // Convert cents to dollars
        invoices = append(invoices, inv)
    }

    // Check for errors from iterating over rows
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating invoice rows: %w", err)
    }

    return invoices, nil
}

func (s *service) initializeData(employerID string) error {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create a worker user
	userID := generateID(UserPrefix)
	_, err = tx.Exec(`
		INSERT INTO users (id, first_name, last_name, email, phone_number, role, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $1, $1)
	`, userID, "John", "Doe", "john.doe@example.com", "1234567890", userID, userID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Create a shift
	// todo make the 
	shiftID := generateID(ShiftPrefix)
	_, err = tx.Exec(`
		INSERT INTO shifts (id, start_date, end_date, location, shift_name, shifts_filled, shift_description, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, shiftID, time.Now(), time.Now().AddDate(0, 0, 7), "Main Street", "Day Shift", 4,  "Regular day shift", time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to insert shift: %w", err)
	}

	// Create 10 invoices
	generateInvoices(tx, shiftID, employerID)

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Data initialization completed successfully")
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

		_, err := tx.Exec(`
			INSERT INTO invoices (id, start_date, end_date, invoice_amount, status, shift_id, invoice_name, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		`, invoiceID, time.Now(), time.Now().AddDate(0, 0, 7), randomAmount, "pending", shiftID, randomShiftName, employerID)
		
		if err != nil {
			return fmt.Errorf("failed to insert invoice %d: %w", i+1, err)
		}
	}

	return nil
}


