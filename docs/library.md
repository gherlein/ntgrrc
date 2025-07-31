# Library Architecture Design for ntgrrc

## Overview

This document outlines the design for refactoring ntgrrc to separate the CLI interface from the core functionality, enabling it to be used as a library in other Go programs. The refactoring will follow clean architecture principles to create a well-structured, maintainable, and reusable library.

## Current Architecture Problems

The current codebase has several issues that prevent library usage:

1. **Tight CLI Coupling**: Business logic is embedded in command structures
2. **Global State**: Uses global options passed through command execution
3. **Direct Output**: Functions print directly to stdout instead of returning data
4. **Mixed Concerns**: HTTP, parsing, and formatting logic are intertwined

## Proposed Architecture

### Package Structure

```
ntgrrc/
├── cmd/
│   └── ntgrrc/
│       └── main.go          # CLI application
├── pkg/
│   └── netgear/
│       ├── client.go        # Main client interface
│       ├── auth.go          # Authentication logic
│       ├── models.go        # Data structures
│       ├── poe.go           # POE management
│       ├── port.go          # Port management
│       ├── errors.go        # Error types
│       └── internal/
│           ├── http.go      # HTTP client
│           ├── parser.go    # HTML parsing
│           └── crypto.go    # Password encryption
└── examples/
    ├── basic_usage.go
    └── advanced_usage.go
```

### Core API Design

```go
// pkg/netgear/client.go
package netgear

import (
    "context"
    "time"
)

// Client represents a connection to a Netgear switch
type Client struct {
    address  string
    model    Model
    http     *httpClient
    token    string
    tokenMgr TokenManager
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
        c.http.timeout = timeout
    }
}

// NewClient creates a new Netgear switch client
func NewClient(address string, opts ...ClientOption) (*Client, error) {
    client := &Client{
        address:  address,
        http:     newHTTPClient(),
        tokenMgr: NewMemoryTokenManager(),
    }
    
    for _, opt := range opts {
        opt(client)
    }
    
    // Detect model
    model, err := client.detectModel()
    if err != nil {
        return nil, err
    }
    client.model = model
    
    return client, nil
}

// Login authenticates with the switch
func (c *Client) Login(ctx context.Context, password string) error {
    // Implementation
}

// POE returns the POE management interface
func (c *Client) POE() *POEManager {
    return &POEManager{client: c}
}

// Ports returns the port management interface
func (c *Client) Ports() *PortManager {
    return &PortManager{client: c}
}
```

### Model Definitions

```go
// pkg/netgear/models.go
package netgear

// Model represents a Netgear switch model
type Model string

const (
    ModelGS305EP  Model = "GS305EP"
    ModelGS305EPP Model = "GS305EPP"
    ModelGS308EP  Model = "GS308EP"
    ModelGS308EPP Model = "GS308EPP"
    ModelGS316EP  Model = "GS316EP"
    ModelGS316EPP Model = "GS316EPP"
)

// POEPortStatus represents the status of a POE port
type POEPortStatus struct {
    PortID               int
    PortName             string
    Status               string
    PowerClass           string
    VoltageV             float64
    CurrentMA            float64
    PowerW               float64
    TemperatureC         float64
    ErrorStatus          string
}

// POEPortSettings represents POE port configuration
type POEPortSettings struct {
    PortID              int
    PortName            string
    Enabled             bool
    Mode                POEMode
    Priority            POEPriority
    PowerLimitType      POELimitType
    PowerLimitW         float64
    DetectionType       string
    LongerDetectionTime bool
}

// PortSettings represents switch port configuration
type PortSettings struct {
    PortID        int
    PortName      string
    Speed         PortSpeed
    IngressLimit  string
    EgressLimit   string
    FlowControl   bool
    Status        PortStatus
    LinkSpeed     string
}

// Enums
type POEMode string
const (
    POEMode8023af    POEMode = "802.3af"
    POEMode8023at    POEMode = "802.3at"
    POEModeLegacy    POEMode = "legacy"
    POEModePre8023at POEMode = "pre-802.3at"
)

type POEPriority string
const (
    POEPriorityLow      POEPriority = "low"
    POEPriorityHigh     POEPriority = "high"
    POEPriorityCritical POEPriority = "critical"
)

type POELimitType string
const (
    POELimitTypeNone  POELimitType = "none"
    POELimitTypeClass POELimitType = "class"
    POELimitTypeUser  POELimitType = "user"
)

type PortSpeed string
const (
    PortSpeedAuto      PortSpeed = "auto"
    PortSpeed10MHalf   PortSpeed = "10M half"
    PortSpeed10MFull   PortSpeed = "10M full"
    PortSpeed100MHalf  PortSpeed = "100M half"
    PortSpeed100MFull  PortSpeed = "100M full"
    PortSpeedDisable   PortSpeed = "disable"
)

type PortStatus string
const (
    PortStatusAvailable PortStatus = "available"
    PortStatusConnected PortStatus = "connected"
    PortStatusDisabled  PortStatus = "disabled"
)
```

### POE Management Interface

```go
// pkg/netgear/poe.go
package netgear

import "context"

// POEManager handles POE-related operations
type POEManager struct {
    client *Client
}

// GetStatus retrieves POE status for all ports
func (m *POEManager) GetStatus(ctx context.Context) ([]POEPortStatus, error) {
    // Implementation
}

// GetSettings retrieves POE settings for all ports
func (m *POEManager) GetSettings(ctx context.Context) ([]POEPortSettings, error) {
    // Implementation
}

// UpdatePort updates settings for specific ports
func (m *POEManager) UpdatePort(ctx context.Context, updates ...POEPortUpdate) error {
    // Implementation
}

// POEPortUpdate represents changes to apply to a POE port
type POEPortUpdate struct {
    PortID         int
    Enabled        *bool
    Mode           *POEMode
    Priority       *POEPriority
    PowerLimitType *POELimitType
    PowerLimitW    *float64
    DetectionType  *string
}

// CyclePower performs a power cycle on specified ports
func (m *POEManager) CyclePower(ctx context.Context, portIDs ...int) error {
    // Implementation
}
```

### Port Management Interface

```go
// pkg/netgear/port.go
package netgear

import "context"

// PortManager handles port-related operations
type PortManager struct {
    client *Client
}

// GetSettings retrieves port settings
func (m *PortManager) GetSettings(ctx context.Context) ([]PortSettings, error) {
    // Implementation
}

// UpdatePort updates settings for specific ports
func (m *PortManager) UpdatePort(ctx context.Context, updates ...PortUpdate) error {
    // Implementation
}

// PortUpdate represents changes to apply to a port
type PortUpdate struct {
    PortID       int
    Name         *string
    Speed        *PortSpeed
    IngressLimit *string
    EgressLimit  *string
    FlowControl  *bool
}
```

### Error Handling

```go
// pkg/netgear/errors.go
package netgear

import "fmt"

// Error types
type ErrorType string

const (
    ErrorTypeAuth      ErrorType = "authentication"
    ErrorTypeNetwork   ErrorType = "network"
    ErrorTypeParsing   ErrorType = "parsing"
    ErrorTypeModel     ErrorType = "model"
    ErrorTypeOperation ErrorType = "operation"
)

// Error represents a netgear client error
type Error struct {
    Type    ErrorType
    Message string
    Cause   error
}

func (e *Error) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s error: %s: %v", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
    return e.Cause
}

// Sentinel errors
var (
    ErrNotAuthenticated = &Error{Type: ErrorTypeAuth, Message: "not authenticated"}
    ErrSessionExpired   = &Error{Type: ErrorTypeAuth, Message: "session expired"}
    ErrModelNotSupported = &Error{Type: ErrorTypeModel, Message: "model not supported"}
)
```

### Token Management Interface

```go
// pkg/netgear/auth.go
package netgear

import "context"

// TokenManager handles token persistence
type TokenManager interface {
    // GetToken retrieves a stored token
    GetToken(ctx context.Context, address string) (token string, model Model, err error)
    
    // StoreToken saves a token
    StoreToken(ctx context.Context, address string, token string, model Model) error
    
    // DeleteToken removes a stored token
    DeleteToken(ctx context.Context, address string) error
}

// MemoryTokenManager stores tokens in memory
type MemoryTokenManager struct {
    tokens map[string]tokenData
    mu     sync.RWMutex
}

type tokenData struct {
    token string
    model Model
}

// FileTokenManager stores tokens in files (current behavior)
type FileTokenManager struct {
    dir string
}

func NewFileTokenManager(dir string) *FileTokenManager {
    if dir == "" {
        dir = filepath.Join(os.TempDir(), ".config", "ntgrrc")
    }
    return &FileTokenManager{dir: dir}
}
```

## CLI Refactoring

The CLI will be refactored to use the library:

```go
// cmd/ntgrrc/main.go
package main

import (
    "context"
    "fmt"
    "github.com/alecthomas/kong"
    "github.com/nitram509/ntgrrc/pkg/netgear"
)

type Context struct {
    Client *netgear.Client
}

var cli struct {
    // Global flags
    Address      string              `required:"" help:"switch address" short:"a"`
    OutputFormat string              `help:"output format" enum:"md,json" default:"md"`
    TokenDir     string              `help:"token directory" short:"d"`
    
    // Commands
    Login LoginCmd `cmd:"" help:"authenticate with switch"`
    POE   POECmd   `cmd:"" help:"manage POE settings"`
    Port  PortCmd  `cmd:"" help:"manage port settings"`
}

type LoginCmd struct {
    Password string `help:"password" short:"p"`
}

func (cmd *LoginCmd) Run(ctx *Context) error {
    password := cmd.Password
    if password == "" {
        password = promptPassword()
    }
    
    return ctx.Client.Login(context.Background(), password)
}

type POECmd struct {
    Status   POEStatusCmd   `cmd:"" help:"show POE status"`
    Settings POESettingsCmd `cmd:"" help:"show POE settings"`
    Set      POESetCmd      `cmd:"" help:"update POE settings"`
    Cycle    POECycleCmd    `cmd:"" help:"cycle POE ports"`
}

type POEStatusCmd struct{}

func (cmd *POEStatusCmd) Run(ctx *Context) error {
    statuses, err := ctx.Client.POE().GetStatus(context.Background())
    if err != nil {
        return err
    }
    
    formatter := getFormatter(cli.OutputFormat)
    formatter.FormatPOEStatus(statuses)
    return nil
}

// ... other command implementations

func main() {
    // Create token manager
    tokenMgr := netgear.NewFileTokenManager(cli.TokenDir)
    
    // Create client
    client, err := netgear.NewClient(
        cli.Address,
        netgear.WithTokenManager(tokenMgr),
    )
    if err != nil {
        panic(err)
    }
    
    // Parse and run command
    ctx := &Context{Client: client}
    kongCtx := kong.Parse(&cli)
    err = kongCtx.Run(ctx)
    kongCtx.FatalIfErrorf(err)
}
```

## Usage Examples

### Basic Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/nitram509/ntgrrc/pkg/netgear"
)

func main() {
    // Create client
    client, err := netgear.NewClient("192.168.1.10")
    if err != nil {
        log.Fatal(err)
    }
    
    // Login
    err = client.Login(context.Background(), "mypassword")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get POE status
    statuses, err := client.POE().GetStatus(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    for _, status := range statuses {
        fmt.Printf("Port %d: %s (%.2fW)\n", 
            status.PortID, status.Status, status.PowerW)
    }
    
    // Update POE settings
    err = client.POE().UpdatePort(context.Background(),
        netgear.POEPortUpdate{
            PortID:   1,
            Enabled:  &[]bool{true}[0],
            Priority: &[]netgear.POEPriority{netgear.POEPriorityHigh}[0],
        },
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Usage with Custom Token Manager

```go
package main

import (
    "context"
    "database/sql"
    
    "github.com/nitram509/ntgrrc/pkg/netgear"
)

// DatabaseTokenManager stores tokens in a database
type DatabaseTokenManager struct {
    db *sql.DB
}

func (m *DatabaseTokenManager) GetToken(ctx context.Context, address string) (string, netgear.Model, error) {
    var token string
    var model string
    err := m.db.QueryRowContext(ctx, 
        "SELECT token, model FROM tokens WHERE address = ?", address).
        Scan(&token, &model)
    return token, netgear.Model(model), err
}

func (m *DatabaseTokenManager) StoreToken(ctx context.Context, address, token string, model netgear.Model) error {
    _, err := m.db.ExecContext(ctx,
        "INSERT OR REPLACE INTO tokens (address, token, model) VALUES (?, ?, ?)",
        address, token, string(model))
    return err
}

func (m *DatabaseTokenManager) DeleteToken(ctx context.Context, address string) error {
    _, err := m.db.ExecContext(ctx, "DELETE FROM tokens WHERE address = ?", address)
    return err
}

func main() {
    db, _ := sql.Open("sqlite3", "tokens.db")
    tokenMgr := &DatabaseTokenManager{db: db}
    
    client, err := netgear.NewClient(
        "192.168.1.10",
        netgear.WithTokenManager(tokenMgr),
        netgear.WithTimeout(30*time.Second),
    )
    // ... use client
}
```

## Migration Strategy

### Phase 1: Create Library Package Structure
1. Create `pkg/netgear` package hierarchy
2. Define public interfaces and models
3. Implement error types

### Phase 2: Extract Core Logic
1. Move authentication logic to `auth.go`
2. Extract HTTP client to `internal/http.go`
3. Move HTML parsing to `internal/parser.go`
4. Create POE and Port managers

### Phase 3: Refactor CLI
1. Update command structures to use library
2. Move formatting logic to CLI layer
3. Remove global state dependencies

### Phase 4: Testing
1. Create unit tests for library components
2. Integration tests with mock HTTP responses
3. CLI compatibility tests

### Phase 5: Documentation
1. Generate godoc documentation
2. Create usage examples
3. Migration guide for existing users

## Benefits

1. **Reusability**: Other Go programs can import and use the library
2. **Testability**: Clean interfaces enable better unit testing
3. **Maintainability**: Clear separation of concerns
4. **Extensibility**: Easy to add new features without affecting CLI
5. **Type Safety**: Strongly typed API prevents runtime errors

## Compatibility Considerations

1. **CLI Compatibility**: The CLI interface remains unchanged
2. **Import Path**: Use Go modules for versioning
3. **Breaking Changes**: Follow semantic versioning
4. **Deprecation**: Mark old code patterns as deprecated before removal

## Future Enhancements

1. **Async Operations**: Add goroutine-safe operations
2. **Event Streaming**: WebSocket support for real-time updates
3. **Bulk Operations**: Batch multiple operations efficiently
4. **Caching**: Add optional response caching
5. **Middleware**: Support for interceptors and hooks