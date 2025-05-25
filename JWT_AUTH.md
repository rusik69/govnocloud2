# JWT Authentication for GovnoCloud2 Client

This document describes the JWT (JSON Web Token) authentication implementation for the GovnoCloud2 client.

## Overview

The GovnoCloud2 client has been updated to use JWT authentication instead of basic authentication headers. This provides better security and follows modern authentication practices.

## Changes Made

### 1. New Authentication Module (`pkg/client/auth.go`)

- Added `JWTClaims` struct to define token claims
- Implemented `GenerateToken()` method to create JWT tokens
- Implemented `SetAuthHeader()` method to add JWT tokens to HTTP requests

### 2. Updated Client Methods

All client methods in the following files have been updated to use JWT authentication:

- `pkg/client/vms.go` - Virtual machine operations
- `pkg/client/users.go` - User management operations
- `pkg/client/containers.go` - Container operations
- `pkg/client/nodes.go` - Node management operations

### 3. Token Configuration

- **Signing Method**: HMAC SHA-256 (HS256)
- **Expiration**: 24 hours from token generation
- **Secret Key**: Uses the client password as the signing key
- **Claims**: Includes username and standard JWT claims (issued at, expires at, not before)

## Usage

The JWT authentication is transparent to the user. Simply create a client instance and use it normally:

```go
package main

import (
    "github.com/rusik69/govnocloud2/pkg/client"
)

func main() {
    // Create client - JWT tokens will be generated automatically
    c := client.NewClient("localhost", "6969", "admin", "password")
    
    // Use client methods normally - JWT auth happens automatically
    vms, err := c.ListVMs("default")
    if err != nil {
        // Handle error
    }
    
    // All other client methods work the same way
    users, err := c.ListUsers()
    nodes, err := c.ListNodes()
    // etc.
}
```

## Security Features

1. **Token Expiration**: Tokens automatically expire after 24 hours
2. **Secure Signing**: Uses HMAC SHA-256 for token signing
3. **Standard Claims**: Includes standard JWT claims for validation
4. **Automatic Generation**: New tokens are generated for each request

## Testing

Run the JWT authentication tests:

```bash
go test ./pkg/client/auth_test.go ./pkg/client/auth.go ./pkg/client/vms.go -v
```

## Example

See `examples/jwt_auth_example.go` for a complete example of using the JWT-authenticated client.

## Migration from Basic Auth

If you were previously using the client with basic authentication, no code changes are required. The client interface remains the same, but now uses JWT tokens internally instead of basic auth headers.

### Before (Basic Auth)
```
User: admin
Password: password
```

### After (JWT)
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Token Structure

The JWT tokens contain the following claims:

```json
{
  "username": "admin",
  "iss": "",
  "sub": "",
  "aud": null,
  "exp": 1640995200,
  "nbf": 1640908800,
  "iat": 1640908800
}
```

- `username`: The authenticated user's username
- `exp`: Token expiration time (Unix timestamp)
- `iat`: Token issued at time (Unix timestamp)
- `nbf`: Token not valid before time (Unix timestamp) 