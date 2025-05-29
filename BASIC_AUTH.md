# Basic HTTP Authentication for GovnoCloud2 Client

This document describes the basic HTTP authentication implementation for the GovnoCloud2 client.

## Overview

The GovnoCloud2 client uses standard HTTP Basic Authentication for all API requests. This provides simple and reliable authentication using username and password credentials.

## Authentication Method

- **Method**: HTTP Basic Authentication (RFC 7617)
- **Header**: `Authorization: Basic <base64-encoded-credentials>`
- **Encoding**: Base64 encoding of `username:password`

## Usage

The basic authentication is transparent to the user. Simply create a client instance with your credentials:

```go
package main

import (
    "github.com/rusik69/govnocloud2/pkg/client"
)

func main() {
    // Create client with basic authentication
    c := client.NewClient("localhost", "6969", "admin", "password")
    
    // Use client methods - basic auth happens automatically
    vms, err := c.ListVMs("default")
    if err != nil {
        // Handle error
    }
    
    // All other client methods work the same way
    users, err := c.ListUsers()
    nodes, err := c.ListNodes()
    containers, err := c.ListContainers("default")
    namespaces, err := c.ListNamespaces()
    // etc.
}
```

## Client Methods with Basic Auth

All client methods automatically include basic authentication:

### Virtual Machine Operations (`pkg/client/vms.go`)
- `CreateVM(name, image, size, namespace string) error`
- `ListVMs(namespace string) ([]string, error)`
- `GetVM(name, namespace string) (*types.VM, error)`
- `DeleteVM(name, namespace string) error`
- `WaitVM(name, namespace string) error`
- `StartVM(name, namespace string) error`
- `StopVM(name, namespace string) error`
- `RestartVM(name, namespace string) error`

### User Management Operations (`pkg/client/users.go`)
- `CreateUser(username, password string) error`
- `ListUsers() ([]string, error)`
- `GetUser(username string) (*types.User, error)`
- `DeleteUser(username string) error`
- `UpdateUser(username, password string) error`
- `CreateUserKey(username, key string) error`
- `ListUserKeys(username string) ([]string, error)`

### Container Operations (`pkg/client/containers.go`)
- `CreateContainer(name, image, namespace string) error`
- `ListContainers(namespace string) ([]string, error)`
- `GetContainer(name, namespace string) (*types.Container, error)`
- `DeleteContainer(name, namespace string) error`

### Node Management Operations (`pkg/client/nodes.go`)
- `ListNodes() ([]string, error)`
- `CreateNode(name, host, user, key string) error`
- `GetNode(name string) (*types.Node, error)`
- `DeleteNode(name string) error`
- `UpdateNode(name, host, user, key string) error`
- `PingNode(name string) error`

### Namespace Operations (`pkg/client/namespaces.go`)
- `CreateNamespace(name string) error`
- `DeleteNamespace(name string) error`
- `ListNamespaces() ([]string, error)`
- `GetNamespace(name string) (string, error)`

## Security Features

1. **Standard Protocol**: Uses well-established HTTP Basic Authentication
2. **Simple Implementation**: No token management or expiration handling needed
3. **Transparent**: Authentication happens automatically on every request
4. **Compatible**: Works with standard HTTP authentication mechanisms

## Example

See `examples/basic_auth_example.go` for a complete example of using the basic auth client.

## Migration from JWT Auth

If you were previously using JWT authentication, no code changes are required. The client interface remains exactly the same, but now uses HTTP Basic Authentication instead of JWT tokens.

### Before (JWT)
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### After (Basic Auth)
```
Authorization: Basic YWRtaW46cGFzc3dvcmQ=
```

## Testing

Run the basic authentication tests:

```bash
go test ./pkg/client/... -v
```

Note: Tests require a running GovnoCloud2 server instance.

## Configuration

The client is configured with:

- **Host**: Server hostname or IP address
- **Port**: Server port number
- **Username**: Authentication username
- **Password**: Authentication password

```go
client := client.NewClient("localhost", "6969", "admin", "password")
``` 