package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/config"
)

// todo: look into when i would use https://{yourDomain}/.well-known/jwks.json

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope    string   `json:"scope"`
	Email    string   `json:"https://traba.fs0ciety.dev/email"`
	DBUserId string   `json:"https://traba.fs0ciety.dev/db_user_id"`
	Roles    []string `json:"https://traba.fs0ciety.dev/roles"`
}

// CustomClaims defines any custom data / claims wanted.
// The Validator will call the Validate function which
// is where custom validation logic can be defined.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken(cfg *config.Config) func(next http.Handler) http.Handler {
	log.Println("ensuring valid token")
	issuerURL, err := url.Parse(cfg.Auth0IssuerBaseURL)
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	// todo: do i need to create a new caching provider every time?
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{cfg.Auth0Audience},
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
