package netgear

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"ntgrrc/pkg/netgear/internal"
)

// Client represents a connection to a Netgear switch
type Client struct {
	address    string
	model      Model
	httpClient *internal.HTTPClient
	token      string
	tokenMgr   TokenManager
	detector   *internal.ModelDetector
	verbose    bool
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithTokenManager sets a custom token manager
func WithTokenManager(tm TokenManager) ClientOption {
	return func(c *Client) {
		c.tokenMgr = tm
	}
}

// WithTimeout sets the HTTP timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient = internal.NewHTTPClient(c.address, timeout, c.verbose)
	}
}

// WithVerbose enables verbose logging
func WithVerbose(verbose bool) ClientOption {
	return func(c *Client) {
		c.verbose = verbose
		if c.httpClient != nil {
			c.httpClient.SetVerbose(verbose)
		}
	}
}

// NewClient creates a new Netgear switch client
func NewClient(address string, opts ...ClientOption) (*Client, error) {
	client := &Client{
		address:    address,
		httpClient: internal.NewHTTPClient(address, 10*time.Second, false),
		tokenMgr:   NewMemoryTokenManager(),
		detector:   internal.NewModelDetector(),
		verbose:    false,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Try to load existing token first
	ctx := context.Background()
	token, model, err := client.tokenMgr.GetToken(ctx, address)
	if err == nil {
		client.token = token
		client.model = model
		if client.verbose {
			fmt.Printf("Loaded existing token for model %s\n", model)
		}
	} else {
		// Detect model if no token available
		model, err := client.detectModel(ctx)
		if err != nil {
			return nil, NewModelError("failed to detect switch model", err)
		}
		client.model = model
		if client.verbose {
			fmt.Printf("Detected model: %s\n", model)
		}
	}

	return client, nil
}

// detectModel attempts to detect the switch model by making a request to the root page
func (c *Client) detectModel(ctx context.Context) (Model, error) {
	resp, err := c.httpClient.Get(ctx, "/", nil)
	if err != nil {
		return "", NewNetworkError("failed to connect to switch", err)
	}

	body, err := c.httpClient.ReadBody(resp)
	if err != nil {
		return "", NewNetworkError("failed to read response", err)
	}

	modelString := c.detector.DetectFromHTML(body)
	if modelString == "" {
		return "", ErrModelNotDetected
	}

	model := Model(modelString)
	if !model.IsSupported() {
		return "", NewModelError(fmt.Sprintf("detected model %s is not supported", model), nil)
	}

	return model, nil
}

// Login authenticates with the switch
func (c *Client) Login(ctx context.Context, password string) error {
	if password == "" {
		return NewAuthError("password cannot be empty", nil)
	}

	// Perform authentication based on model type
	var token string
	var err error

	authType := GetAuthenticationType(c.model)
	switch authType {
	case AuthTypeSession:
		token, err = c.loginWithSession(ctx, password)
	case AuthTypeGambit:
		token, err = c.loginWithGambit(ctx, password)
	default:
		return NewAuthError(fmt.Sprintf("unsupported authentication type for model %s", c.model), nil)
	}

	if err != nil {
		return err
	}

	c.token = token

	// Store token for future use
	err = c.tokenMgr.StoreToken(ctx, c.address, token, c.model)
	if err != nil {
		// Log warning but don't fail login
		if c.verbose {
			fmt.Printf("Warning: failed to store token: %v\n", err)
		}
	}

	return nil
}

// loginWithSession performs session-based authentication (30x series)
func (c *Client) loginWithSession(ctx context.Context, password string) (string, error) {
	// Encrypt password using MD5
	encryptedPassword := internal.EncryptPassword(password)

	// Prepare login data
	data := url.Values{}
	data.Set("password", encryptedPassword)

	// Make login request
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, err := c.httpClient.Post(ctx, "/login.cgi", data, headers)
	if err != nil {
		return "", NewNetworkError("login request failed", err)
	}

	body, err := c.httpClient.ReadBody(resp)
	if err != nil {
		return "", NewNetworkError("failed to read login response", err)
	}

	// Extract session token from response
	token := internal.ExtractSessionToken(body)
	if token == "" {
		// Check for error message
		if errorMsg := internal.ExtractErrorMessage(body); errorMsg != "" {
			return "", NewAuthError(fmt.Sprintf("login failed: %s", errorMsg), nil)
		}
		return "", ErrInvalidCredentials
	}

	return token, nil
}

// loginWithGambit performs Gambit-based authentication (316 series)
func (c *Client) loginWithGambit(ctx context.Context, password string) (string, error) {
	// First, get the login page to extract any required tokens
	resp, err := c.httpClient.Get(ctx, "/", nil)
	if err != nil {
		return "", NewNetworkError("failed to get login page", err)
	}

	body, err := c.httpClient.ReadBody(resp)
	if err != nil {
		return "", NewNetworkError("failed to read login page", err)
	}

	// Encrypt password
	encryptedPassword := internal.EncryptPassword(password)

	// Prepare login data for Gambit authentication
	data := url.Values{}
	data.Set("password", encryptedPassword)

	// Make login request
	resp, err = c.httpClient.Post(ctx, "/login.cgi", data, nil)
	if err != nil {
		return "", NewNetworkError("gambit login request failed", err)
	}

	body, err = c.httpClient.ReadBody(resp)
	if err != nil {
		return "", NewNetworkError("failed to read gambit login response", err)
	}

	// Extract Gambit token
	token := internal.ExtractGambitToken(body)
	if token == "" {
		if errorMsg := internal.ExtractErrorMessage(body); errorMsg != "" {
			return "", NewAuthError(fmt.Sprintf("gambit login failed: %s", errorMsg), nil)
		}
		return "", ErrInvalidCredentials
	}

	return token, nil
}

// IsAuthenticated returns true if the client has a valid token
func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}

// GetModel returns the detected switch model
func (c *Client) GetModel() Model {
	return c.model
}

// GetAddress returns the switch address
func (c *Client) GetAddress() string {
	return c.address
}

// POE returns the POE management interface
func (c *Client) POE() *POEManager {
	return newPOEManager(c)
}

// Ports returns the port management interface
func (c *Client) Ports() *PortManager {
	return newPortManager(c)
}

// Logout clears the authentication token
func (c *Client) Logout(ctx context.Context) error {
	c.token = ""
	
	// Remove stored token
	err := c.tokenMgr.DeleteToken(ctx, c.address)
	if err != nil && c.verbose {
		fmt.Printf("Warning: failed to delete stored token: %v\n", err)
	}
	
	return nil
}

// makeAuthenticatedRequest makes an HTTP request with appropriate authentication
func (c *Client) makeAuthenticatedRequest(ctx context.Context, method, path string, data url.Values) (string, error) {
	if !c.IsAuthenticated() {
		return "", ErrNotAuthenticated
	}

	headers := make(map[string]string)

	// Add authentication based on model type
	authType := GetAuthenticationType(c.model)
	switch authType {
	case AuthTypeSession:
		// Use session cookie
		headers["Cookie"] = fmt.Sprintf("SID=%s", c.token)
	case AuthTypeGambit:
		// Add Gambit parameter to URL
		if data == nil {
			data = url.Values{}
		}
		data.Set("Gambit", c.token)
	}

	if method == "GET" {
		if len(data) > 0 {
			// Add query parameters for GET requests
			path += "?" + data.Encode()
		}
		httpResp, err := c.httpClient.Get(ctx, path, headers)
		if err != nil {
			return "", NewNetworkError("GET request failed", err)
		}
		return c.httpClient.ReadBody(httpResp)
	} else {
		httpResp, err := c.httpClient.Post(ctx, path, data, headers)
		if err != nil {
			return "", NewNetworkError("POST request failed", err)
		}
		return c.httpClient.ReadBody(httpResp)
	}
}