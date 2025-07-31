# ntgrrc Design Documentation

## Overview

ntgrrc (Netgear Remote Control) is a command-line interface (CLI) tool designed to manage Netgear managed plus switches in the 300 series. Since Netgear does not provide a REST API for these switches, ntgrrc employs web scraping techniques to interact with the switch's web interface, enabling configuration management and status monitoring.

## Architecture

### Core Components

The application follows a command-based architecture with the following key components:

```mermaid
graph TB
    subgraph "CLI Layer"
        Main[main.go<br/>Entry Point]
        Commands[Command Structures<br/>LoginCommand, PoeCommand, etc.]
    end
    
    subgraph "Business Logic"
        Auth[Authentication<br/>login.go]
        POE[POE Management<br/>poe_*.go]
        Port[Port Management<br/>port_*.go]
        Model[Model Detection<br/>netgear_model.go]
    end
    
    subgraph "Infrastructure Layer"
        HTTP[HTTP Client<br/>http.go]
        Token[Token Management<br/>token.go]
        Parser[HTML Parsing<br/>goquery]
        Format[Formatters<br/>formatter_*.go]
    end
    
    subgraph "External"
        Switch[Netgear Switch<br/>Web Interface]
    end
    
    Main --> Commands
    Commands --> Auth
    Commands --> POE
    Commands --> Port
    Auth --> Model
    Auth --> Token
    POE --> HTTP
    Port --> HTTP
    HTTP --> Switch
    POE --> Parser
    Port --> Parser
    POE --> Format
    Port --> Format
```

### Component Descriptions

#### 1. CLI Layer
- **main.go**: Application entry point using Kong library for CLI parsing
- **Command Structures**: Implements command pattern for different operations (login, poe, port, etc.)

#### 2. Authentication & Security
- **login.go**: Handles authentication flow including:
  - Password prompting (with hidden input)
  - Password encryption using MD5 with seed value
  - Session token management
- **token.go**: Manages persistent session tokens for authenticated requests

#### 3. Model Management
- **netgear_model.go**: Detects switch models and provides model-specific behavior
- Supports models: GS305EP(P), GS308EP(P), GS316EP(P)
- Different models have different authentication mechanisms and API endpoints

#### 4. Feature Modules
- **POE Management**: Controls Power over Ethernet settings
  - Status monitoring (poe_status.go)
  - Settings management (poe_settings.go)
  - Port configuration (poe_set_port.go)
  - Power cycling (poe_cycle.go)
- **Port Management**: Controls switch port settings
  - Port settings (port_settings.go)
  - Port configuration (port_set.go)

#### 5. Infrastructure
- **http.go**: HTTP client wrapper with authentication cookie management
- **Formatters**: Output formatting in Markdown (default) or JSON
- **Utilities**: Helper functions for parsing and data manipulation

## Authentication Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Auth
    participant HTTP
    participant Switch
    participant Token
    
    User->>CLI: ntgrrc login --address=switch_ip
    CLI->>Auth: Run login command
    Auth->>User: Prompt for password
    User->>Auth: Enter password
    Auth->>HTTP: Detect switch model
    HTTP->>Switch: GET /
    Switch->>HTTP: HTML response
    HTTP->>Auth: Model detected
    Auth->>HTTP: Get seed value
    HTTP->>Switch: GET /login.cgi or /wmi/login
    Switch->>HTTP: HTML with seed value
    Auth->>Auth: Encrypt password with seed
    Auth->>HTTP: POST login request
    HTTP->>Switch: POST /login.cgi with encrypted password
    Switch->>HTTP: Response with session cookie
    HTTP->>Auth: Extract token
    Auth->>Token: Store token to file
    Token->>User: Login successful
```

## Command Execution Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Command
    participant Token
    participant HTTP
    participant Switch
    participant Parser
    participant Formatter
    
    User->>CLI: ntgrrc poe status --address=switch_ip
    CLI->>Command: Execute POE status command
    Command->>Token: Read stored token
    Token->>Command: Return token & model
    Command->>HTTP: Request status page
    HTTP->>HTTP: Add authentication headers
    HTTP->>Switch: GET /getPoePortStatus.cgi
    Switch->>HTTP: HTML response
    HTTP->>Command: Return HTML
    Command->>Parser: Parse HTML with goquery
    Parser->>Command: Extract status data
    Command->>Formatter: Format output
    Formatter->>User: Display table (Markdown/JSON)
```

## Data Flow

### Model-Specific Behavior

The application handles two main model families differently:

#### GS30x Series (305/308)
- Uses SID cookie for authentication
- Endpoints: `/login.cgi`, `/getPoePortStatus.cgi`, etc.
- Session token in cookie header

#### GS316 Series
- Uses Gambit token in URL parameters
- Endpoints: `/wmi/login`, `/iss/specific/poePortStatus.html`, etc.
- Token embedded in query string

### HTML Parsing Strategy

The tool uses goquery (jQuery-like library for Go) to parse HTML responses:

1. **Status Pages**: Extract data from specific HTML elements
2. **Settings Pages**: Parse form values and configuration
3. **Model Detection**: Analyze page titles and content

### Output Formatting

The formatter system provides flexibility in output:

```mermaid
graph LR
    Data[Raw Data<br/>Arrays] --> Switch{Format?}
    Switch -->|Markdown| MD[Markdown Table<br/>Human Readable]
    Switch -->|JSON| JSON[JSON Object<br/>Machine Parseable]
```

## Security Considerations

1. **Password Handling**:
   - Passwords are never stored on disk
   - Input is hidden during entry using terminal control
   - Encryption uses switch-provided seed value

2. **Token Management**:
   - Tokens stored in temp directory with predictable naming
   - File-based storage allows parallel sessions
   - No token expiration handling (relies on switch timeout)

3. **Network Communication**:
   - HTTP only (no HTTPS support)
   - No certificate validation
   - Vulnerable to MITM attacks on local network

## Error Handling

The application handles several error scenarios:

1. **Network Errors**: Connection failures, timeouts
2. **Authentication Errors**: Invalid credentials, expired sessions
3. **Parsing Errors**: Unexpected HTML structure
4. **Model Detection Errors**: Unknown switch models

## Extensibility

### Adding New Commands

1. Create command structure with Kong tags
2. Implement Run method
3. Add to CLI struct in main.go
4. Create supporting business logic

### Supporting New Models

1. Add model constant in netgear_model.go
2. Implement model detection logic
3. Add model-specific endpoints
4. Update parsing logic for HTML differences

### Adding Output Formats

1. Define new OutputFormat constant
2. Implement formatter function
3. Add case to format switch statements
4. Update CLI help text

## Limitations

1. **Web Scraping Dependency**: Vulnerable to web interface changes
2. **HTTP Only**: No secure communication option
3. **Limited Feature Coverage**: Only implements subset of switch capabilities
4. **Model-Specific Code**: Requires updates for new firmware/models
5. **No Concurrent Modifications**: Single session per switch

## Module Interaction Diagram

```mermaid
graph TB
    subgraph "Command Processing"
        direction TB
        CLI[CLI Parser<br/>Kong] --> CMD{Command<br/>Router}
        CMD --> Login[Login<br/>Command]
        CMD --> POE[POE<br/>Commands]
        CMD --> Port[Port<br/>Commands]
        CMD --> Debug[Debug<br/>Command]
    end
    
    subgraph "Core Services"
        direction LR
        Auth[Authentication<br/>Service]
        Model[Model<br/>Detection]
        Token[Token<br/>Storage]
        HTTP[HTTP<br/>Client]
    end
    
    subgraph "Data Processing"
        direction TB
        Parser[HTML<br/>Parser]
        Mapper[Value<br/>Mappers]
        Format[Output<br/>Formatters]
    end
    
    Login --> Auth
    Login --> Model
    Auth --> Token
    POE --> HTTP
    Port --> HTTP
    HTTP --> Parser
    Parser --> Mapper
    Mapper --> Format
    
    style CLI fill:#f9f,stroke:#333,stroke-width:2px
    style Auth fill:#bbf,stroke:#333,stroke-width:2px
    style HTTP fill:#bbf,stroke:#333,stroke-width:2px
    style Format fill:#bfb,stroke:#333,stroke-width:2px
```

## POE Configuration Workflow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant PoeSet
    participant ValueMapper
    participant HTTP
    participant Switch
    participant Formatter
    
    User->>CLI: poe set --port=1,2 --power=enable
    CLI->>PoeSet: Parse command arguments
    PoeSet->>PoeSet: Validate port numbers
    PoeSet->>HTTP: Request current settings
    HTTP->>Switch: GET PoEPortConfig.cgi
    Switch->>HTTP: HTML with current config
    HTTP->>PoeSet: Return HTML
    PoeSet->>PoeSet: Parse current settings
    PoeSet->>ValueMapper: Map CLI values to form values
    ValueMapper->>PoeSet: Return encoded values
    PoeSet->>HTTP: Build form data
    HTTP->>Switch: POST with new settings
    Switch->>HTTP: Updated configuration
    HTTP->>PoeSet: Confirmation response
    PoeSet->>Formatter: Format results
    Formatter->>User: Display updated settings
```

## Port Settings Update Flow

```mermaid
sequenceDiagram
    participant User
    participant PortSet
    participant HTTP
    participant Switch
    participant Parser
    
    User->>PortSet: port set -p 1 -n "Camera" -s "100M full"
    PortSet->>HTTP: Get current port config
    HTTP->>Switch: Request port settings page
    Switch->>HTTP: HTML response
    HTTP->>PortSet: Current settings
    PortSet->>Parser: Extract form data
    Parser->>PortSet: Current values
    PortSet->>PortSet: Merge new values
    PortSet->>HTTP: Submit updated config
    HTTP->>Switch: POST new settings
    Switch->>HTTP: Success response
    HTTP->>PortSet: Confirmation
    PortSet->>User: Display results
```

## State Management

```mermaid
stateDiagram-v2
    [*] --> Unauthenticated
    Unauthenticated --> Authenticating: login command
    Authenticating --> Authenticated: successful login
    Authenticating --> Unauthenticated: failed login
    Authenticated --> Executing: run command
    Executing --> Authenticated: command complete
    Authenticated --> Unauthenticated: session timeout
    Executing --> ReAuth: session expired
    ReAuth --> Authenticated: auto re-login
    ReAuth --> Unauthenticated: re-login failed
```

## Future Considerations

1. **HTTPS Support**: Add TLS communication options
2. **Configuration Files**: Support for connection profiles
3. **Batch Operations**: Execute multiple commands in sequence
4. **Error Recovery**: Automatic re-authentication on session timeout
5. **Extended Model Support**: Add support for more switch models
6. **API Abstraction**: Create abstraction layer for easier model support