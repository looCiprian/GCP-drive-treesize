package authentication

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

// RedirectURL is the OAuth2 callback registered for this application.
// The local web server must be reachable on this address.
const RedirectURL = "http://localhost:8080/oauth2callback"

// Config returns the OAuth2 configuration used to read Google Drive metadata.
func Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "99542254178-s7hc2ejf0h8n57lk1he7fiu78stbrfup.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.metadata.readonly"},
		Endpoint:     google.Endpoint,
		RedirectURL:  RedirectURL,
	}
}

// Flow holds the state of a single sign-in attempt: the anti-CSRF state
// value and the PKCE verifier used when exchanging the authorization code.
type Flow struct {
	State    string
	verifier string
}

// NewFlow starts a new sign-in attempt with fresh random values.
func NewFlow() (*Flow, error) {
	state, err := randomString()
	if err != nil {
		return nil, err
	}
	verifier, err := randomString()
	if err != nil {
		return nil, err
	}
	return &Flow{State: state, verifier: verifier}, nil
}

// AuthURL returns the Google consent page URL for this flow.
func (f *Flow) AuthURL(config *oauth2.Config) string {
	sum := sha256.Sum256([]byte(f.verifier))
	return config.AuthCodeURL(f.State, oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", base64.RawURLEncoding.EncodeToString(sum[:])),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))
}

// Exchange trades the authorization code received on the callback for a token.
func (f *Flow) Exchange(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", f.verifier))
}

// NewService creates a Drive client from a token. The token is kept in memory
// only and refreshed automatically by the oauth2 client when needed.
func NewService(ctx context.Context, config *oauth2.Config, tok *oauth2.Token) (*drive.Service, error) {
	return drive.NewService(ctx, option.WithHTTPClient(config.Client(ctx, tok)))
}

func randomString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
