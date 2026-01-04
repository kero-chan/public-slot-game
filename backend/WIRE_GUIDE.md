# Google Wire Dependency Injection Guide

This project uses [Google Wire](https://github.com/google/wire) for compile-time dependency injection. Wire automatically generates code to wire up your application's dependencies, making the code more maintainable and testable.

## Table of Contents
- [What is Wire?](#what-is-wire)
- [Architecture Overview](#architecture-overview)
- [Provider Sets](#provider-sets)
- [How It Works](#how-it-works)
- [Development Workflow](#development-workflow)
- [Adding New Dependencies](#adding-new-dependencies)
- [Troubleshooting](#troubleshooting)

## What is Wire?

Wire is a code generation tool that automates connecting components using dependency injection. Unlike runtime dependency injection frameworks, Wire generates code at compile-time, providing:

- **Compile-time safety**: Catches wiring errors before runtime
- **No reflection**: Better performance and easier debugging
- **Explicit dependencies**: Clear dependency graph
- **Type-safe**: Full type checking by the Go compiler

## Architecture Overview

The application is organized into layers, each with its own provider set:

```
┌─────────────────────────────────────┐
│         Application Layer           │
│  (cmd/server/wire.go)               │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Handler Layer                │
│  (internal/api/handler/wire.go)     │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Service Layer                │
│  (internal/service/wire.go)         │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Repository Layer               │
│  (internal/infra/repository/wire.go)│
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│    Infrastructure (DB, Logger, etc) │
│  (internal/db/wire.go, etc)         │
└─────────────────────────────────────┘
```

## Provider Sets

Each layer defines a `ProviderSet` in a `wire.go` file:

### 1. Config Layer (`internal/config/wire.go`)
```go
var ProviderSet = wire.NewSet(
    Load, // Loads configuration from environment
)
```

### 2. Logger Layer (`internal/pkg/logger/wire.go`)
```go
var ProviderSet = wire.NewSet(
    ProvideLogger, // Creates logger from config
)
```

### 3. Database Layer (`internal/db/wire.go`)
```go
var ProviderSet = wire.NewSet(
    ProvideDatabase, // Creates GORM database connection
)
```

### 4. Repository Layer (`internal/infra/repository/wire.go`)
```go
var ProviderSet = wire.NewSet(
    NewPlayerGormRepository,
    NewSessionGormRepository,
    NewSpinGormRepository,
    NewFreeSpinsGormRepository,
    NewReelStripGormRepository,
)
```

### 5. Service Layer (`internal/service/wire.go`)
```go
var ProviderSet = wire.NewSet(
    NewPlayerService,
    NewSessionService,
    NewSpinService,
    NewFreeSpinsService,
    NewReelStripService,
)
```

### 6. Handler Layer (`internal/api/handler/wire.go`)
```go
var ProviderSet = wire.NewSet(
    NewAuthHandler,
    NewPlayerHandler,
    NewSessionHandler,
    NewSpinHandler,
    NewFreeSpinsHandler,
)
```

### 7. Game Engine Layer (`internal/game/engine/wire.go`)
```go
var ProviderSet = wire.NewSet(
    ProvideGameEngine, // Creates game engine instance
)
```

### 8. Server Layer (`internal/server/wire.go`)
```go
var ProviderSet = wire.NewSet(
    ProvideFiberApp, // Creates Fiber web application
)
```

## How It Works

### 1. Wire Injector Definition (`cmd/server/wire.go`)

The injector file defines what dependencies to wire:

```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    // ... imports
)

type Application struct {
    Config           *config.Config
    Logger           *logger.Logger
    App              *fiber.App
    AuthHandler      *handler.AuthHandler
    PlayerHandler    *handler.PlayerHandler
    SessionHandler   *handler.SessionHandler
    SpinHandler      *handler.SpinHandler
    FreeSpinsHandler *handler.FreeSpinsHandler
}

func InitializeApplication() (*Application, error) {
    wire.Build(
        config.ProviderSet,
        logger.ProviderSet,
        db.ProviderSet,
        engine.ProviderSet,
        repository.ProviderSet,
        service.ProviderSet,
        handler.ProviderSet,
        server.ProviderSet,
        wire.Struct(new(Application), "*"),
    )
    return &Application{}, nil
}
```

### 2. Wire Code Generation

When you run `wire`, it generates `wire_gen.go`:

```go
//go:build !wireinject
// +build !wireinject

package main

func InitializeApplication() (*Application, error) {
    configConfig, err := config.Load()
    if err != nil {
        return nil, err
    }
    loggerLogger := logger.ProvideLogger(configConfig)
    app := server.ProvideFiberApp(configConfig, loggerLogger)
    gormDB, err := db.ProvideDatabase(configConfig, loggerLogger)
    if err != nil {
        return nil, err
    }
    playerRepository := repository.NewPlayerGormRepository(gormDB)
    // ... all dependencies wired automatically

    application := &Application{
        Config:           configConfig,
        Logger:           loggerLogger,
        App:              app,
        AuthHandler:      authHandler,
        PlayerHandler:    playerHandler,
        SessionHandler:   sessionHandler,
        SpinHandler:      spinHandler,
        FreeSpinsHandler: freeSpinsHandler,
    }
    return application, nil
}
```

### 3. Using in Main (`cmd/server/main.go`)

```go
func main() {
    application, err := InitializeApplication()
    if err != nil {
        fmt.Printf("Failed to initialize application: %v\n", err)
        os.Exit(1)
    }

    log := application.Logger
    cfg := application.Config

    // Setup routes and start server
    server.SetupRoutes(
        application.App,
        cfg,
        log,
        application.AuthHandler,
        application.PlayerHandler,
        application.SessionHandler,
        application.SpinHandler,
        application.FreeSpinsHandler,
    )

    // ... server startup code
}
```

## Development Workflow

### Generate Wire Code

Wire code is automatically generated when you build:

```bash
make build   # Generates Wire code and builds
```

Or manually:

```bash
make wire    # Generate Wire code only
```

### Check Wire Configuration

Verify your Wire setup without generating code:

```bash
make wire-check
```

### Install Wire Tool

Wire is automatically installed when you run `make install-tools`:

```bash
make install-tools
```

Or install manually:

```bash
go install github.com/google/wire/cmd/wire@latest
```

## Adding New Dependencies

### Example: Adding a New Service

1. **Create the service constructor** in `internal/service/`:

```go
// email_service.go
package service

import (
    "github.com/slotmachine/backend/internal/pkg/logger"
)

type EmailService struct {
    logger *logger.Logger
}

func NewEmailService(log *logger.Logger) *EmailService {
    return &EmailService{
        logger: log,
    }
}
```

2. **Add to provider set** in `internal/service/wire.go`:

```go
var ProviderSet = wire.NewSet(
    NewPlayerService,
    NewSessionService,
    NewSpinService,
    NewFreeSpinsService,
    NewReelStripService,
    NewEmailService,  // Add here
)
```

3. **Add to Application struct** in `cmd/server/wire.go` (if needed):

```go
type Application struct {
    Config           *config.Config
    Logger           *logger.Logger
    App              *fiber.App
    EmailService     *service.EmailService  // Add here
    // ... other fields
}
```

4. **Regenerate Wire code**:

```bash
make wire
```

Wire will automatically figure out the dependency chain!

### Example: Adding a New Repository

1. **Create repository** in `internal/infra/repository/`:

```go
// email_gorm.go
package repository

import (
    "github.com/slotmachine/backend/domain/email"
    "gorm.io/gorm"
)

type EmailGormRepository struct {
    db *gorm.DB
}

func NewEmailGormRepository(db *gorm.DB) email.Repository {
    return &EmailGormRepository{db: db}
}
```

2. **Add to provider set** in `internal/infra/repository/wire.go`:

```go
var ProviderSet = wire.NewSet(
    NewPlayerGormRepository,
    NewSessionGormRepository,
    NewSpinGormRepository,
    NewFreeSpinsGormRepository,
    NewReelStripGormRepository,
    NewEmailGormRepository,  // Add here
)
```

3. **Update service constructor** to use it:

```go
func NewEmailService(repo email.Repository, log *logger.Logger) *EmailService {
    return &EmailService{
        repo:   repo,
        logger: log,
    }
}
```

4. **Regenerate**:

```bash
make wire
```

## Troubleshooting

### Error: "unused provider"

This means you've added a provider to a set but it's not used anywhere.

**Solution**: Either use it in a constructor or remove it from the provider set.

### Error: "no provider found for X"

Wire can't figure out how to create type X.

**Solution**: Add a provider function for X to the appropriate provider set.

### Error: "multiple providers for X"

More than one provider returns the same type.

**Solution**: Ensure only one provider returns each type, or use specific binding.

### Error: "cycle detected"

Your dependencies form a circular reference.

**Solution**: Refactor to break the cycle. Consider using interfaces or restructuring.

### Wire not regenerating

Wire code might be cached.

**Solution**:
```bash
rm cmd/server/wire_gen.go
make wire
```

### Build tag issues

If you see errors about build tags:

**Solution**: Ensure `wire.go` has:
```go
//go:build wireinject
// +build wireinject
```

And `wire_gen.go` has:
```go
//go:build !wireinject
// +build !wireinject
```

## Best Practices

1. **One provider per type**: Each type should have exactly one provider function
2. **Use interfaces**: Depend on interfaces, not concrete types
3. **Keep provider sets organized**: Group by layer (repository, service, handler)
4. **Document providers**: Add comments explaining what each provider does
5. **Run wire before committing**: Always regenerate Wire code before committing
6. **Don't edit wire_gen.go**: This file is auto-generated, never edit it manually
7. **Test your wiring**: Use `make wire-check` to validate configuration

## Files Overview

| File | Purpose |
|------|---------|
| `cmd/server/wire.go` | Main application injector definition |
| `cmd/server/wire_gen.go` | Auto-generated wiring code (DO NOT EDIT) |
| `internal/*/wire.go` | Provider sets for each layer |
| `scripts/seed_reelstrips/wire.go` | Injector for seed script |
| `scripts/seed_reelstrips/wire_gen.go` | Auto-generated seed script wiring |

## Additional Resources

- [Wire Documentation](https://github.com/google/wire/blob/main/docs/guide.md)
- [Wire Best Practices](https://github.com/google/wire/blob/main/docs/best-practices.md)
- [Wire FAQ](https://github.com/google/wire/blob/main/docs/faq.md)

## Summary

Wire provides:
- ✅ Compile-time dependency injection
- ✅ Type-safe wiring
- ✅ Clear dependency graphs
- ✅ Easy testing and mocking
- ✅ No runtime overhead
- ✅ Automatic code generation

The build process automatically handles Wire generation, so you can focus on writing business logic while Wire handles the wiring!
