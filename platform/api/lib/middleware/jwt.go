package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope         string   `json:"scope"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"https://localhost:3000/roles"`
	UserMetadata  struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		CompanyName string `json:"company_name"`
		PhoneNumber string `json:"phone_number"`
	} `json:"https://localhost:3000/user_metadata"`
}

// Auth0Management handles Auth0 Management API operations
type Auth0Management struct {
	domain       string
	clientID     string
	clientSecret string
	roleID       string
	token        string
	tokenExpiry  time.Time
}

// NewAuth0Management creates a new Auth0Management instance
func NewAuth0Management() *Auth0Management {
	// read from .env file 
	godotenv.Load("../../.env")
	return &Auth0Management{
		domain:       os.Getenv("AUTH0_DOMAIN"),
		clientID:     os.Getenv("AUTH0_MANAGEMENT_CLIENT_ID"),
		clientSecret: os.Getenv("AUTH0_MANAGEMENT_CLIENT_SECRET"),
		roleID:       os.Getenv("AUTH0_ROLE_ID"),
	}
}

// getManagementToken obtains or refreshes the Management API token
func (a *Auth0Management) getManagementToken() error {
	if a.token != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	url := fmt.Sprintf("https://%s/oauth/token", a.domain)
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
		"audience":      fmt.Sprintf("https://%s/api/v2/", a.domain),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal token request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to get management token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	a.token = result.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return nil
}

// AssignRole assigns a role to a user
func (a *Auth0Management) AssignRole(userID, roleID string) error {
	if err := a.getManagementToken(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/api/v2/users/%s/roles", a.domain, userID)
	payload := map[string][]string{
		"roles": {roleID},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal role assignment request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create role assignment request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to assign role, status: %d", resp.StatusCode)
	}

	return nil
}

// CustomClaims defines any custom data / claims wanted.
// The Validator will call the Validate function which
// is where custom validation logic can be defined.
func (c CustomClaims) Validate(ctx context.Context) error {
	// Check if the email is present and has a valid format
	log.Println("Validating email")
	if c.Email == "" {
		return errors.New("email is required")
	}
	_, err := url.ParseRequestURI(c.Email)
	if err != nil {
		return errors.New("email format is invalid")
	}

	// Check if email is verified
	if !c.EmailVerified {
		return errors.New("email is not verified")
	}

	// Add any other custom validations you may need for UserMetadata
	if len(c.Roles) == 0 {
		if c.UserMetadata.FirstName == "" || c.UserMetadata.LastName == "" {
			return errors.New("first name and last name are required")
		}

		if c.UserMetadata.CompanyName == "" {
			return errors.New("company name is required")
		}

		// Initialize Auth0 Management API client
		auth0Management := NewAuth0Management()

		// Extract user ID from the context
		claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok {
			return errors.New("failed to get claims from context")
		}

		// Get the user ID from the subject claim
		userID := claims.RegisteredClaims.Subject

		// Assign the employer role
		err := auth0Management.AssignRole(userID, auth0Management.roleID)
		if err != nil {
			log.Printf("Failed to assign employer role: %v", err)
			// Don't fail validation if role assignment fails
			// The role will be assigned on next request
		}

		// todo add the role employer to the roles array of role id rol_lz7KugKHb6tiTJVl
	}
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	godotenv.Load("../../.env")
	log.Println("Setting up jwt middleware")
	issuerURL, err := url.Parse(os.Getenv("AUTH0_ISSUER_BASE_URL"))
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	// todo: do i need to create a new caching provider every time?
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

func (c *CustomClaims) HasRole(role string) ([]string, error) {
	if !c.EmailVerified {
		return nil, errors.New("email not verified")
	}
	return c.Roles, nil
}
