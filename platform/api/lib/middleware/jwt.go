package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// todo: look into when i would use https://{yourDomain}/.well-known/jwks.json

// CustomClaims contains custom data we want from the token.
type Role struct {
	RoleID string `json:"role_id"`
}

type CustomClaims struct {
	Scope    string   `json:"scope"`
	Email    string   `json:"https://traba-staging.fs0ciety.dev/email"`
	DBUserId string   `json:"https://traba-staging.fs0ciety.dev/db_user_id"`
	Roles    []string `json:"https://traba-staging.fs0ciety.dev/roles"`
}

// CustomClaims defines any custom data / claims wanted.
// The Validator will call the Validate function which
// is where custom validation logic can be defined.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	log.Println("ensuring valid token")
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
