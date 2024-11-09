package service

import (
	"context"
	"database/sql"
	"testing"
	"log"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/test"
)

func Test_CreateUser(t *testing.T) {
	db, container := test.SetupAndFillDatabaseContainer("")
	defer func(container testcontainers.Container) {
		err := test.TeardownDatabaseContainer(container)
		if err != nil {
			log.Fatalf("failed to close container down: %v\n", err)
		}
	}(container)
	defer db.Close()

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
				Email:      "john.doe@example.com",
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
				Email:      "john.doe@example.com", // Same email as first test
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
	
				// Verify initialization data (assuming initializeData creates some default records)
				// This would depend on what initializeData does, but here's an example:
				/*
				var count int
				err = db.QueryRow(`SELECT COUNT(*) FROM some_initialized_table WHERE user_id = $1`, userID).Scan(&count)
				assert.NoError(t, err)
				assert.Greater(t, count, 0)
				*/
			})
		}
}


// func Test_FetchInvoices(t *testing.T) {
// 	// Setup
// 	db, container := test.SetupAndFillDatabaseContainer("")
// 	defer func(container testcontainers.Container) {
// 		err := test.TeardownDatabaseContainer(container)
// 		if err != nil {
// 			t.Fatalf("failed to close container down: %v\n", err)
// 		}
// 	}(container)
// 	defer db.Close()

// 	// Create service instance
// 	svc := NewService(db)

// 	// Setup test data
// 	testUserID := "user_123"
// 	setupTestData(t, db, testUserID)

// 	// Define test cases
// 	tests := []struct {
// 		name           string
// 		userID         string
// 		searchTerm     string
// 		expectedCount  int
// 		expectedError  bool
// 		validateResult func(*testing.T, []InvoiceResponse)
// 	}{
// 		{
// 			name:          "fetch all invoices for user",
// 			userID:        testUserID,
// 			searchTerm:    "",
// 			expectedCount: 3, // Assuming we inserted 3 invoices in setupTestData
// 			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
// 				assert.Len(t, invoices, 3)
// 				// Validate first invoice
// 				assert.Equal(t, "Invoice 1", invoices[0].InvoiceName)
// 				assert.Equal(t, float64(100.00), invoices[0].InvoiceAmount)
// 			},
// 		},
// 		{
// 			name:          "fetch invoices with search term",
// 			userID:        testUserID,
// 			searchTerm:    "Invoice 1",
// 			expectedCount: 1,
// 			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
// 				assert.Len(t, invoices, 1)
// 				assert.Equal(t, "Invoice 1", invoices[0].InvoiceName)
// 			},
// 		},
// 		{
// 			name:          "fetch invoices with partial search term",
// 			userID:        testUserID,
// 			searchTerm:    "voice",
// 			expectedCount: 3, // Should match all invoices containing "voice"
// 			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
// 				assert.Len(t, invoices, 3)
// 				for _, inv := range invoices {
// 					assert.Contains(t, inv.InvoiceName, "voice")
// 				}
// 			},
// 		},
// 		{
// 			name:          "no results for search term",
// 			userID:        testUserID,
// 			searchTerm:    "nonexistent",
// 			expectedCount: 0,
// 			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
// 				assert.Empty(t, invoices)
// 			},
// 		},
// 		{
// 			name:          "invalid user ID",
// 			userID:        "invalid_user",
// 			searchTerm:    "",
// 			expectedCount: 0,
// 			validateResult: func(t *testing.T, invoices []InvoiceResponse) {
// 				assert.Empty(t, invoices)
// 			},
// 		},
// 		{
// 			name:          "empty user ID",
// 			userID:        "",
// 			expectedError: true,
// 		},
// 	}

// 	// Execute test cases
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Fetch invoices
// 			invoices, err := svc.FetchInvoices(context.Background(), tt.userID, tt.searchTerm)

// 			// Check error expectations
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 				return
// 			}

// 			// Verify successful fetch
// 			assert.NoError(t, err)
// 			assert.Len(t, invoices, tt.expectedCount)

// 			// Run custom validation if provided
// 			if tt.validateResult != nil {
// 				tt.validateResult(t, invoices)
// 			}
// 		})
// 	}
// }


// // Helper function to setup test data
// func setupTestData(t *testing.T, db *sql.DB, userID string) {
// 	// First, clear any existing data
// 	clearTestData(t, db)

// 	// Setup shifts first (since invoices reference shifts)
// 	shifts := []struct {
// 		id        string
// 		startDate time.Time
// 		endDate   time.Time
// 	}{
// 		{
// 			id:        "shift_1",
// 			startDate: time.Now().AddDate(0, 0, -7),
// 			endDate:   time.Now(),
// 		},
// 		{
// 			id:        "shift_2",
// 			startDate: time.Now().AddDate(0, 0, -14),
// 			endDate:   time.Now().AddDate(0, 0, -7),
// 		},
// 		{
// 			id:        "shift_3",
// 			startDate: time.Now().AddDate(0, 0, -21),
// 			endDate:   time.Now().AddDate(0, 0, -14),
// 		},
// 	}

// 	// Insert shifts
// 	for _, shift := range shifts {
// 		_, err := db.Exec(`
// 			INSERT INTO shifts (id, start_date, end_date, created_by)
// 			VALUES ($1, $2, $3, $4)
// 		`, shift.id, shift.startDate, shift.endDate, userID)
// 		assert.NoError(t, err)
// 	}

// 	// Insert test invoices
// 	testInvoices := []struct {
// 		id            string
// 		invoiceAmount float64
// 		shiftID       string
// 		status        string
// 		invoiceName   string
// 	}{
// 		{
// 			id:            "invoice_1",
// 			invoiceAmount: 100.00,
// 			shiftID:       "shift_1",
// 			status:        "pending",
// 			invoiceName:   "Invoice 1",
// 		},
// 		{
// 			id:            "invoice_2",
// 			invoiceAmount: 200.00,
// 			shiftID:       "shift_2",
// 			status:        "paid",
// 			invoiceName:   "Invoice 2",
// 		},
// 		{
// 			id:            "invoice_3",
// 			invoiceAmount: 300.00,
// 			shiftID:       "shift_3",
// 			status:        "pending",
// 			invoiceName:   "Invoice 3",
// 		},
// 	}

// 	for _, inv := range testInvoices {
// 		_, err := db.Exec(`
// 			INSERT INTO invoices (id, invoice_amount, shift_id, status, invoice_name, created_by)
// 			VALUES ($1, $2, $3, $4, $5, $6)
// 		`, inv.id, inv.invoiceAmount, inv.shiftID, inv.status, inv.invoiceName, userID)
// 		assert.NoError(t, err)
// 	}
// }

// // Helper function to clear test data
// func clearTestData(t *testing.T, db *sql.DB) {
// 	_, err := db.Exec(`DELETE FROM invoices`)
// 	assert.NoError(t, err)
// 	_, err = db.Exec(`DELETE FROM shifts`)
// 	assert.NoError(t, err)
// }