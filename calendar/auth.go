package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Auth handles Google OAuth2 authentication
type Auth struct {
	credentialsPath string
	tokenPath       string
	scopes          []string
}

// NewAuth creates a new Auth instance
func NewAuth(credentialsPath, tokenPath string) *Auth {
	return &Auth{
		credentialsPath: credentialsPath,
		tokenPath:       tokenPath,
		scopes: []string{
			calendar.CalendarEventsScope, // Read/write access to events
		},
	}
}

// GetClient returns an authenticated HTTP client
func (a *Auth) GetClient(ctx context.Context) (*http.Client, error) {
	config, err := a.getConfig()
	if err != nil {
		return nil, err
	}

	token, err := a.getToken(config)
	if err != nil {
		return nil, err
	}

	return config.Client(ctx, token), nil
}

// getConfig loads OAuth2 config from credentials file
func (a *Auth) getConfig() (*oauth2.Config, error) {
	b, err := os.ReadFile(a.credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w\n\nPlease download credentials.json from Google Cloud Console:\n1. Go to https://console.cloud.google.com/\n2. Create a project and enable Google Calendar API\n3. Create OAuth 2.0 credentials (Desktop app)\n4. Download and save as 'credentials.json' in the current directory", err)
	}

	config, err := google.ConfigFromJSON(b, a.scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %w", err)
	}

	return config, nil
}

// getToken retrieves token from file or initiates new authorization
func (a *Auth) getToken(config *oauth2.Config) (*oauth2.Token, error) {
	token, err := a.loadToken()
	if err == nil {
		return token, nil
	}

	// Token doesn't exist, get a new one
	return a.getTokenFromWeb(config)
}

// loadToken loads token from file
func (a *Auth) loadToken() (*oauth2.Token, error) {
	f, err := os.Open(a.tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// getTokenFromWeb initiates browser-based OAuth flow
func (a *Auth) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("\n[Auth] Authorization Required")
	fmt.Println("-----------------------------")
	fmt.Println("Go to the following link in your browser:")
	fmt.Printf("\n%s\n\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token: %w", err)
	}

	// Save token for future use
	if err := a.saveToken(token); err != nil {
		fmt.Printf("Warning: unable to save token: %v\n", err)
	}

	return token, nil
}

// saveToken saves the token to file
func (a *Auth) saveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(a.tokenPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(a.tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// GetCalendarService creates an authenticated Calendar service
func GetCalendarService(ctx context.Context, credentialsPath, tokenPath string) (*calendar.Service, error) {
	auth := NewAuth(credentialsPath, tokenPath)

	client, err := auth.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return srv, nil
}
