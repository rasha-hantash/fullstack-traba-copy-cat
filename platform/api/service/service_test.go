package service

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/test"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

var (
	db        *sql.DB
	container testcontainers.Container
)

func TestMain(m *testing.M) {
	// Setup
	db, container = test.SetupDatabaseContainer()

	// Run tests
	code := m.Run()

	// Teardown
	if err := test.TeardownDatabaseContainer(container); err != nil {
		log.Fatalf("failed to close container down: %v\n", err)
	}
	db.Close()

	os.Exit(code)
}

func Test_CreateUser(t *testing.T) {
	// Create service instance
	svc := NewService(db)

	// Define test cases
	tests := []struct {
		name          string
		input         *User
		expectedError bool
		errorMessage  string
		validate      func(t *testing.T, db *sql.DB, userID string)
	}{
		{
			name: "successful user creation",
			input: &User{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				PhoneNumber: "1234567890",
				CompanyName: "Test Company",
			},
			expectedError: false,
			validate: func(t *testing.T, db *sql.DB, userID string) {
				// Query the database to verify user was created
				var user User
				err := db.QueryRow(`
					SELECT first_name, last_name, email, phone_number, company_name 
					FROM users WHERE id = $1`, userID).Scan(
					&user.FirstName, &user.LastName, &user.Email,
					&user.PhoneNumber, &user.CompanyName,
				)

				assert.NoError(t, err)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
				assert.Equal(t, "john.doe@example.com", user.Email)
				assert.Equal(t, "1234567890", user.PhoneNumber)
				assert.Equal(t, "Test Company", user.CompanyName)
			},
		},
		{
			name: "duplicate email",
			input: &User{
				FirstName:   "Jane",
				LastName:    "Doe",
				Email:       "john.doe@example.com", // Same email as first test
				PhoneNumber: "0987654321",
				CompanyName: "Another Company",
			},
			expectedError: true,
			errorMessage:  "error creating user",
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create user
			userID, err := svc.CreateUser(context.Background(), tt.input)

			// Check error expectations
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				return
			}

			// Verify successful creation
			assert.NoError(t, err)
			assert.NotEmpty(t, userID)
			assert.True(t, len(userID) > 0)

			// Run validation if provided
			if tt.validate != nil {
				tt.validate(t, db, userID)
			}
		})
	}

	clearTestData(t, db)
}

func Test_FetchInvoices(t *testing.T) {
	// Create service instance
	svc := NewService(db)

	userID, err := svc.CreateUser(context.Background(), &User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		PhoneNumber: "1234567890",
		CompanyName: "Test Company",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, userID)
	assert.True(t, len(userID) > 0)

	// Define test cases
	tests := []struct {
		name           string
		userID         string
		searchTerm     string
		expectedCount  int
		expectedError  bool
		validateResult func(*testing.T, []InvoiceResponse)
	}{
		{
			name:          "fetch all invoices for user",
			userID:        userID,
			searchTerm:    "",
			expectedCount: 10, // Assuming we inserted 3 invoices in setupTestData
			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
				assert.Len(t, invoices, 10)
			},
		},
		{
			name:          "fetch invoices with search term",
			userID:        userID,
			searchTerm:    "Morning",
			expectedCount: 1,
			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
				assert.Len(t, invoices, 1)
				assert.Equal(t, "Morning Shift", invoices[0].InvoiceName)
			},
		},
		{
			name:          "fetch invoices with partial search term",
			userID:        userID,
			searchTerm:    "Shift",
			expectedCount: 10, // Should match all invoices containing "voice"
			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
				assert.Len(t, invoices, 10)
				for _, inv := range invoices {
					assert.Contains(t, inv.InvoiceName, "Shift")
				}
			},
		},
		{
			name:          "no results for search term",
			userID:        userID,
			searchTerm:    "nonexistent",
			expectedCount: 0,
			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
				assert.Empty(t, invoices)
			},
		},
		{
			name:          "invalid user ID",
			userID:        "invalid_user",
			searchTerm:    "",
			expectedCount: 0,
			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
				assert.Empty(t, invoices)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			expectedError: true,
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fetch invoices
			invoices, err := svc.FetchInvoices(context.Background(), tt.userID, tt.searchTerm)

			// Check error expectations
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			// Verify successful fetch
			assert.NoError(t, err)
			assert.Len(t, invoices, tt.expectedCount)

			// Run custom validation if provided
			if tt.validateResult != nil {
				tt.validateResult(t, invoices)
			}
		})
	}

	clearTestData(t, db)
}

// Helper function to clear test data
func clearTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`DELETE FROM invoices`)
	assert.NoError(t, err)
	_, err = db.Exec(`DELETE FROM shifts`)
	assert.NoError(t, err)
	_, err = db.Exec(`DELETE FROM users`)
	assert.NoError(t, err)
}
