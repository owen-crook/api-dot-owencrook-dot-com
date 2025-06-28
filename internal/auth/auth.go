package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

var (
	verifier *oidc.IDTokenVerifier
)

// AuthenticatedUser holds the decoded user info from the token
type AuthenticatedUser struct {
	Email string
	Sub   string
}

// Init sets up the OIDC token verifier.
// projectClientID is your Google OAuth Client ID from the Google Cloud Console.
func Init(ctx context.Context, projectClientID string) error {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return fmt.Errorf("oidc.NewProvider: %w", err)
	}

	verifier = provider.Verifier(&oidc.Config{
		ClientID: projectClientID,
	})

	return nil
}

// GetUserFromRequest extracts and verifies the bearer token from an HTTP request.
func GetUserFromRequest(r *http.Request) (*AuthenticatedUser, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, errors.New("invalid Authorization header format")
	}

	ctx := r.Context()
	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &AuthenticatedUser{
		Email: claims.Email,
		Sub:   claims.Sub,
	}, nil
}
