# 🚀 Go Learning Project

> golang-standards/project-layout

## Cấu trúc

```
src/
├── api/                     # API layer
│   ├── handler/             # Xử lý yêu cầu HTTP (handler)
│   ├── middleware/           # API middleware
│   └── router.go            # Router definition
├── assets/                  # Static assets
├── build/                   # Build & CI configs
├── cmd/
│   ├── app/                 # Entry point — start app
│   └── cli/                 # CLI commands
├── configs/                 # Config files
├── deployments/             # Docker Compose, k8s, terraform
├── docs/                    # Documentation
├── internal/                # Private code
│   ├── model/               # Domain models / DB models
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic
│   └── util/                # Internal utilities
├── migrations/              # Database migrations
├── pkg/                     # Public shared libraries
├── scripts/                 # Build, install scripts
├── test/                    # Integration & E2E tests
├── third_party/             # External tools
├── tools/                   # Supporting tools
├── vendor/                  # Vendored dependencies
└── web/                     # Frontend (nếu có)
```

## Quick Start

```bash
# Run
go run cmd/app/main.go

# Build
go build -o bin/server ./cmd/app

# Test
go test ./...
```
