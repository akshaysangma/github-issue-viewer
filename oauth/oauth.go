// Package oauth provides OAuth2 authentication for the GitHub API.
// Note: This package is using OAuth2 client credentials flow but using Github
// Apps and not actually Github OAuth2. Based on need might change to OAuth2
// later.
package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrOAuthMissingClientID     = errors.New("missing client ID")
	ErrOAuthMissingClientSecret = errors.New("missing client secret")
	ErrOAuthMissingRedirectURI  = errors.New("missing redirect URI")
	ErrOAuthMissingAuthURL      = errors.New("missing auth URL")
	ErrOAuthMissingTokenURL     = errors.New("missing token URL")
)

type OAuth struct {
	endpoint struct {
		authURL  string
		tokenURL string
	}
	scope        string
	clientID     string
	clientSecret string
	redirectURI  string
}

type OAuthResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	TokenType             string `json:"token_type"`
	Scope                 string `json:"scope"`
}

type OAuthOpt func(*OAuth)

func WithScope(scope string) OAuthOpt {
	return func(o *OAuth) {
		o.scope = scope
	}
}

func WithEndpointURL(authURL string, tokenURL string) OAuthOpt {
	return func(o *OAuth) {
		o.endpoint.authURL = authURL
		o.endpoint.tokenURL = tokenURL
	}
}

func WithClientID(id string) OAuthOpt {
	return func(o *OAuth) {
		o.clientID = id
	}
}

func WithClientSecret(secret string) OAuthOpt {
	return func(o *OAuth) {
		o.clientSecret = secret
	}
}

func WithRedirectURI(uri string) OAuthOpt {
	return func(o *OAuth) {
		o.redirectURI = uri
	}
}

func New(opts ...OAuthOpt) (*OAuth, error) {
	o := &OAuth{}
	for _, opt := range opts {
		opt(o)
	}

	if o.clientID == "" {
		return nil, ErrOAuthMissingClientID
	}

	if o.clientSecret == "" {
		return nil, ErrOAuthMissingClientSecret
	}

	if o.redirectURI == "" {
		return nil, ErrOAuthMissingRedirectURI
	}

	if o.endpoint.authURL == "" {
		return nil, ErrOAuthMissingAuthURL
	}

	if o.endpoint.tokenURL == "" {
		return nil, ErrOAuthMissingTokenURL
	}

	return o, nil
}

func (o *OAuth) GenerateAuthURL() (string, error) {
	state, err := generateOAuthState(32)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		o.endpoint.authURL,
		o.clientID,
		o.redirectURI,
		o.scope,
		state), nil
}

func generateOAuthState(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(randomBytes)
	return state, nil
}

func (o *OAuth) Exchange(code string) (*OAuthResponse, error) {
	tokenURL := fmt.Sprintf("%s?client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		o.endpoint.tokenURL,
		o.clientID,
		o.clientSecret,
		code,
		o.redirectURI)

	req, err := http.NewRequest(http.MethodPost, tokenURL, nil)
	if err != nil {
		return nil, err
	}
	// We set this header since we want the response
	// as JSON
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	httpClient := http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	OAuthResponse := &OAuthResponse{}
	err = json.NewDecoder(res.Body).Decode(&OAuthResponse)
	if err != nil {
		return nil, err
	}
	return OAuthResponse, nil
}
