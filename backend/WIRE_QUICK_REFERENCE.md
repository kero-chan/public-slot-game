# Wire Quick Reference Card

## Common Commands

```bash
# Generate Wire code (automatic on build)
make wire

# Check Wire configuration
make wire-check

# Build (auto-generates Wire code)
make build

# Run development server
make dev

# Install Wire tool
go install github.com/google/wire/cmd/wire@latest
```

## Project Structure

```
backend/
├── cmd/server/
│   ├── wire.go          # Main injector definition
│   ├── wire_gen.go      # Auto-generated (DO NOT EDIT)
│   └── main.go          # Uses InitializeApplication()
│
├── scripts/seed_reelstrips/
│   ├── wire.go          # Seed injector definition
│   ├── wire_gen.go      # Auto-generated
│   └── main.go          # Uses InitializeSeedApplication()
│
└── internal/
    ├── config/
    │   └── wire.go      # Config provider
    ├── pkg/logger/
    │   └── wire.go      # Logger provider
    ├── db/
    │   └── wire.go      # Database provider
    ├── infra/repository/
    │   └── wire.go      # Repository providers
    ├── service/
    │   └── wire.go      # Service providers
    ├── api/handler/
    │   └── wire.go      # Handler providers
    ├── game/engine/
    │   └── wire.go      # Game engine provider
    └── server/
        └── wire.go      # Fiber app provider
```

## Adding New Dependencies

### 1. Create Constructor Function

```go
// internal/service/email_service.go
package service

func NewEmailService(log *logger.Logger) *EmailService {
    return &EmailService{logger: log}
}
```

### 2. Add to Provider Set

```go
// internal/service/wire.go
var ProviderSet = wire.NewSet(
    NewPlayerService,
    NewEmailService,  // Add here
)
```

### 3. Regenerate Wire

```bash
make wire
```

That's it! Wire figures out the rest.

## Provider Set Patterns

### Simple Provider (no dependencies)
```go
func ProvideGameEngine() *GameEngine {
    return NewGameEngine()
}
```

### Provider with Dependencies
```go
func ProvideLogger(cfg *config.Config) *Logger {
    return New(cfg.Logging.Level, cfg.Logging.Format)
}
```

### Provider with Error
```go
func ProvideDatabase(cfg *config.Config, log *Logger) (*gorm.DB, error) {
    return NewGormDB(cfg, log)
}
```

### Provider Set
```go
var ProviderSet = wire.NewSet(
    NewPlayerService,
    NewSessionService,
    NewSpinService,
)
```

## Injector Pattern

```go
//go:build wireinject
// +build wireinject

package main

import "github.com/google/wire"

type Application struct {
    Config  *config.Config
    Logger  *logger.Logger
    Handler *handler.Handler
}

func InitializeApplication() (*Application, error) {
    wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        handler.ProviderSet,
        wire.Struct(new(Application), "*"),
    )
    return &Application{}, nil
}
```

## Usage in Main

```go
func main() {
    app, err := InitializeApplication()
    if err != nil {
        log.Fatal(err)
    }

    // Use app.Config, app.Logger, app.Handler
}
```

## Common Errors & Solutions

### Error: "unused provider"
**Problem**: Provider in set but not used
**Solution**: Remove from provider set or use it

### Error: "no provider found"
**Problem**: Missing provider for type
**Solution**: Add provider function to appropriate wire.go

### Error: "multiple providers"
**Problem**: Two functions return same type
**Solution**: Keep only one provider per type

### Error: "cycle detected"
**Problem**: Circular dependency (A needs B, B needs A)
**Solution**: Refactor to break cycle

## Build Tags Explained

**wire.go** (injector definition):
```go
//go:build wireinject
// +build wireinject
```
- Only compiled when running `wire` command
- Contains Wire.Build() calls

**wire_gen.go** (generated code):
```go
//go:build !wireinject
// +build !wireinject
```
- Only compiled for normal builds
- Contains actual wiring code
- Auto-generated, don't edit

## Dependency Layers

```
Application Layer (main.go)
    ↓
Handler Layer (HTTP/gRPC handlers)
    ↓
Service Layer (business logic)
    ↓
Repository Layer (data access)
    ↓
Infrastructure (DB, Cache, Logger, Config)
```

## Wire vs Manual DI

### Manual (Old Way)
```go
cfg := config.Load()
log := logger.New(cfg)
db := database.Connect(cfg, log)
repo := repository.New(db)
service := service.New(repo, log)
handler := handler.New(service, log)
```

### Wire (New Way)
```go
app, err := InitializeApplication()
// All dependencies wired automatically!
```

## Best Practices

✅ **DO**
- One provider per type
- Group providers by layer
- Keep provider sets organized
- Run `make wire` before commit
- Add comments to providers
- Use interfaces in dependencies

❌ **DON'T**
- Edit wire_gen.go manually
- Have multiple providers for same type
- Create circular dependencies
- Forget to regenerate after changes
- Mix business logic in providers

## Testing with Wire

```go
//go:build wireinject

func InitializeTestApplication() (*Application, error) {
    wire.Build(
        ProvideTestConfig,      // Mock config
        ProvideMockLogger,      // Mock logger
        ProvideMockDB,          // Mock database
        // ... other mock providers
        wire.Struct(new(Application), "*"),
    )
    return &Application{}, nil
}
```

## Makefile Targets

| Command | Description |
|---------|-------------|
| `make wire` | Generate Wire code |
| `make wire-check` | Validate Wire config |
| `make build` | Build (auto-generates Wire) |
| `make dev` | Run dev server |
| `make install-tools` | Install Wire + other tools |

## File Naming Convention

| File | Purpose |
|------|---------|
| `wire.go` | Provider sets & injector definitions |
| `wire_gen.go` | Auto-generated wiring code |
| `*_test.go` | Unit tests |
| `wire_test.go` | Test injectors (optional) |

## Documentation

- Full Guide: [WIRE_GUIDE.md](./WIRE_GUIDE.md)
- Implementation: [WIRE_IMPLEMENTATION_SUMMARY.md](./WIRE_IMPLEMENTATION_SUMMARY.md)
- Official Docs: https://github.com/google/wire

## Support

If you encounter issues:
1. Check error message
2. Run `make wire-check`
3. Verify provider sets
4. Check for circular dependencies
5. See [WIRE_GUIDE.md](./WIRE_GUIDE.md) troubleshooting section
