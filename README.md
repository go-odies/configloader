# Go Config Loader

A Go library for loading and merging application configuration from:

- JSON files
- Environment variables
- Command-line arguments

It is designed for apps that need deterministic config resolution across local development, containerized environments, and production.

## Features

- Load a base JSON config file
- Override values with environment variables
- Override values again with CLI flags/arguments
- Keep a predictable and explicit precedence model

## Precedence

When the same key exists in multiple sources, the winner is:

1. Command-line arguments
2. Environment variables
3. JSON file values

This means later sources override earlier ones.

## Typical Use Cases

- Service configuration for local and cloud environments
- Twelve-factor style env overrides
- Runtime tuning with CLI flags for one-off executions

## Package Layout

```text
.
├── pkg/
│   └── configloader/
├── go.mod
├── README.md
├── CHANGELOG.md
├── LICENSE
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Usage

The library loads configuration in precedence order:

1. Config file values
2. Environment variables
3. Command-line arguments

```go
package main

import (
	"fmt"

	"github.com/go-odies/configloader"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func main() {
	cfg, err := configloader.Load[Config](
		configloader.WithFilePath("./config.json"),
		configloader.WithEnvPrefix("APP"),
		configloader.WithEnvSeparator("__"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s:%d\n", cfg.Host, cfg.Port)
}
```

This version supports nested keys such as `--database.host=example.com` and environment names such as `APP_DATABASE__HOST`.

## Quick Start

1. Clone the repository and open it in the dev container or local Go environment.
2. Run the standard validation commands:

```bash
make test
make vet
make build
```

## Development Setup

### VS Code + Dev Container

1. Open the project in VS Code.
2. Run: Dev Containers: Reopen in Container.
3. Open a terminal and run make help to view available tasks.

### From Host With Docker

```bash
make setup
make docker-shell
```

Stop services:

```bash
make down
```

Destroy services, images, and volumes:

```bash
make teardown
```

## Versioning And Releases

- Follow Semantic Versioning: MAJOR.MINOR.PATCH
- Document all notable changes in [CHANGELOG.md](CHANGELOG.md)
- Tag releases in Git (for example: v0.1.0)

## Make Targets

Common targets:

- make test
- make build
- make fmt
- make vet
- make install-deps
- make help

Container helper targets:

- make docker-test
- make docker-build
- make docker-fmt
- make docker-vet
- make docker-shell

## License

MIT. See [LICENSE](LICENSE).
