# GovnoCloud2

A cloud management platform with HTTP Basic Authentication.

## Authentication

GovnoCloud2 uses standard HTTP Basic Authentication for all API requests. See [BASIC_AUTH.md](BASIC_AUTH.md) for detailed documentation.

## Quick Start

```go
package main

import (
    "github.com/rusik69/govnocloud2/pkg/client"
)

func main() {
    // Create client with basic authentication
    c := client.NewClient("localhost", "6969", "admin", "password")
    
    // Use client methods
    vms, err := c.ListVMs("default")
    // Handle error and use vms...
}
```

## Examples

See the `examples/` directory for usage examples:
- `examples/basic_auth_example.go` - Basic authentication example

## Documentation

- [BASIC_AUTH.md](BASIC_AUTH.md) - Basic HTTP Authentication documentation
