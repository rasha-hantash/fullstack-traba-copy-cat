package middleware

import (
	"context"
	"errors"
	"log"
	// "log/slog"
	"net/http"
	"net/url"
	"time"

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
	}
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	log.Println("Setting up jwt middleware")
	issuerURL, err := url.Parse("https://dev-sjsi88vcdyupj8oq.us.auth0.com/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	// todo: do i need to create a new caching provider every time? 
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{"https://traba-api/"},
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

// GetRoles extracts roles from the validated token claims
// func GetRoles(r *http.Request) ([]string, error) {
// 	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
// 	if !ok {
// 		return nil, errors.New("no claims found in request context")
// 	}

// 	customClaims, ok := claims.CustomClaims.(*CustomClaims)
// 	if !ok {
// 		return nil, errors.New("failed to cast custom claims")
// 	}

// 	return customClaims.Roles, nil
// }

func (c *CustomClaims) HasRole(role string) ([]string, error) {
	if !c.EmailVerified {
		return nil, errors.New("email not verified")
	}
	return c.Roles, nil
}
