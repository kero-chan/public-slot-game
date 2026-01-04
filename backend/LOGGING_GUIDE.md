# Logging Best Practices Guide

This guide explains how to use the logging system with **traceID** and **client IP** for better observability and request tracking.

## Table of Contents
- [Overview](#overview)
- [TraceID and Client IP](#traceid-and-client-ip)
- [Logging in Handlers](#logging-in-handlers)
- [Logging in Services](#logging-in-services)
- [SQL Query Logging](#sql-query-logging)
- [Log Levels](#log-levels)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

The application uses **zerolog** for structured logging with automatic **traceID** and **client IP** injection:

- **TraceID**: Unique identifier for each HTTP request (UUID)
- **Client IP**: IP address of the client making the request
- **Automatic SQL logging**: All database queries include traceID and client IP

### Architecture

```
HTTP Request
    ↓
TraceMiddleware (generates traceID, extracts client IP)
    ↓
RequestLogger (logs all requests with traceID + IP)
    ↓
Handler (uses log.WithTrace(c))
    ↓
Service (receives context with trace info)
    ↓
Repository (GORM logs SQL with trace info)
```

## TraceID and Client IP

### How TraceID is Generated

1. Client can send `X-Trace-ID` header (reuses existing traceID)
2. If not provided, server generates UUID automatically
3. TraceID is returned in response header: `X-Trace-ID`

### How Client IP is Extracted

- Uses Fiber's `c.IP()` method
- Respects `X-Forwarded-For` and `X-Real-IP` headers
- Handles proxies and load balancers correctly

### Example Log Output

```json
{
  "level": "info",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_ip": "192.168.1.100",
  "method": "POST",
  "path": "/api/spin",
  "status": 200,
  "duration": 45,
  "time": "2025-11-19T10:30:00Z",
  "message": "HTTP request"
}
```

## Logging in Handlers

### ✅ Correct Way (With Trace)

Always create a traced logger at the beginning of handler functions:

```go
func (h *SpinHandler) ExecuteSpin(c *fiber.Ctx) error {
    // Create traced logger (includes traceID and clientIP automatically)
    log := h.logger.WithTrace(c)

    log.Info().Msg("Executing spin request")

    // ... handler logic

    if err != nil {
        log.Error().Err(err).Msg("Spin execution failed")
        return c.Status(500).JSON(...)
    }

    log.Info().
        Str("player_id", playerID.String()).
        Float64("bet_amount", req.BetAmount).
        Msg("Spin executed successfully")

    return c.JSON(response)
}
```

### ❌ Wrong Way (Without Trace)

Don't use the base logger directly in handlers:

```go
func (h *SpinHandler) ExecuteSpin(c *fiber.Ctx) error {
    // DON'T DO THIS - missing traceID and clientIP
    h.logger.Info().Msg("Executing spin")  // ❌

    // ... handler logic
}
```

## Logging in Services

Services should use context to pass trace information:

### Method 1: Use Logger with Context

```go
func (s *SpinService) ExecuteSpin(ctx context.Context, playerID uuid.UUID, betAmount float64) (*Spin, error) {
    // Create traced logger from context
    log := s.logger.WithTraceContext(ctx)

    log.Info().
        Str("player_id", playerID.String()).
        Float64("bet_amount", betAmount).
        Msg("Starting spin execution")

    // ... service logic

    if err != nil {
        log.Error().Err(err).Msg("Failed to execute spin")
        return nil, err
    }

    return spin, nil
}
```

### Method 2: Pass Context to Repository

```go
func (s *SpinService) ExecuteSpin(ctx context.Context, playerID uuid.UUID, betAmount float64) (*Spin, error) {
    // Repository will automatically log SQL with traceID and clientIP
    spin, err := s.spinRepo.Create(ctx, newSpin)
    if err != nil {
        s.logger.WithTraceContext(ctx).Error().Err(err).Msg("Failed to save spin")
        return nil, err
    }

    return spin, nil
}
```

### Passing Context from Handler to Service

In handlers, enrich the context with trace info:

```go
import "github.com/slotmachine/backend/internal/pkg/ctxutil"

func (h *SpinHandler) ExecuteSpin(c *fiber.Ctx) error {
    log := h.logger.WithTrace(c)

    // Add trace info to context for services
    ctx := ctxutil.WithTraceInfo(c.Context(), c)

    // Call service with enriched context
    spin, err := h.spinService.ExecuteSpin(ctx, playerID, betAmount)
    if err != nil {
        log.Error().Err(err).Msg("Service call failed")
        return c.Status(500).JSON(...)
    }

    return c.JSON(spin)
}
```

## SQL Query Logging

GORM automatically logs all SQL queries with traceID and client IP when context is passed:

### Example Service/Repository Call

```go
// Service
func (s *PlayerService) GetBalance(ctx context.Context, playerID uuid.UUID) (float64, error) {
    // Pass context to repository - SQL will be logged with trace info
    player, err := s.playerRepo.GetByID(ctx, playerID)
    if err != nil {
        return 0, err
    }
    return player.Balance, nil
}

// Repository
func (r *PlayerGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error) {
    var p player.Player
    // GORM will log this query with traceID and clientIP from context
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
        return nil, err
    }
    return &p, nil
}
```

### SQL Log Output

```json
{
  "level": "debug",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_ip": "192.168.1.100",
  "elapsed": 2.5,
  "rows": 1,
  "sql": "SELECT * FROM \"players\" WHERE id = $1 LIMIT 1",
  "time": "2025-11-19T10:30:00Z",
  "message": "SQL query"
}
```

### Slow Query Logging

Queries taking longer than 200ms are automatically logged as warnings:

```json
{
  "level": "warn",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_ip": "192.168.1.100",
  "elapsed": 450.2,
  "rows": 1000,
  "sql": "SELECT * FROM \"spins\" WHERE player_id = $1",
  "threshold": "200ms",
  "time": "2025-11-19T10:30:00Z",
  "message": "SLOW SQL >= 200ms"
}
```

## Log Levels

### Debug
Development details, SQL queries, detailed flow:
```go
log.Debug().Str("player_id", id).Msg("Fetching player data")
```

### Info
Important business events:
```go
log.Info().
    Str("player_id", playerID).
    Float64("bet_amount", betAmount).
    Msg("Spin executed successfully")
```

### Warn
Warning conditions, slow queries, deprecated usage:
```go
log.Warn().
    Dur("elapsed", duration).
    Msg("Slow database query detected")
```

### Error
Errors that need attention:
```go
log.Error().
    Err(err).
    Str("player_id", playerID).
    Msg("Failed to execute spin")
```

### Fatal
Critical errors that require immediate shutdown:
```go
log.Fatal().Err(err).Msg("Failed to connect to database")
```

## Best Practices

### ✅ DO

1. **Always use traced logger in handlers**:
   ```go
   log := h.logger.WithTrace(c)
   ```

2. **Pass context to all service/repository calls**:
   ```go
   player, err := s.playerRepo.GetByID(ctx, playerID)
   ```

3. **Add context fields for important data**:
   ```go
   log.Info().
       Str("player_id", playerID).
       Float64("amount", amount).
       Msg("Payment processed")
   ```

4. **Log errors with context**:
   ```go
   log.Error().
       Err(err).
       Str("operation", "spin_execution").
       Msg("Operation failed")
   ```

5. **Use appropriate log levels**:
   ```go
   log.Debug() // Development details
   log.Info()  // Business events
   log.Warn()  // Warnings
   log.Error() // Errors
   ```

### ❌ DON'T

1. **Don't use base logger in handlers**:
   ```go
   h.logger.Info().Msg("...") // ❌ Missing trace info
   ```

2. **Don't forget to pass context**:
   ```go
   player, err := s.playerRepo.GetByID(playerID) // ❌ No context
   ```

3. **Don't log sensitive data**:
   ```go
   log.Info().Str("password", pwd).Msg("...") // ❌ Security risk
   ```

4. **Don't log in tight loops**:
   ```go
   for _, item := range items {
       log.Debug().Msg("Processing...") // ❌ Too many logs
   }
   ```

5. **Don't use string formatting**:
   ```go
   log.Info().Msg(fmt.Sprintf("Player: %s", id)) // ❌ Use structured fields
   log.Info().Str("player_id", id).Msg("Player action") // ✅ Better
   ```

## Examples

### Complete Handler Example

```go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/slotmachine/backend/internal/pkg/ctxutil"
)

func (h *SpinHandler) ExecuteSpin(c *fiber.Ctx) error {
    // 1. Create traced logger
    log := h.logger.WithTrace(c)

    // 2. Parse request
    var req SpinRequest
    if err := c.BodyParser(&req); err != nil {
        log.Warn().Err(err).Msg("Invalid request body")
        return c.Status(400).JSON(ErrorResponse{
            Error: "invalid_request",
            Message: "Invalid request body",
        })
    }

    // 3. Log incoming request
    log.Info().
        Float64("bet_amount", req.BetAmount).
        Msg("Spin request received")

    // 4. Get player ID from auth
    playerID := c.Locals("user_id").(string)

    // 5. Enrich context with trace info
    ctx := ctxutil.WithTraceInfo(c.Context(), c)

    // 6. Call service (SQL logs will include traceID and IP)
    spin, err := h.spinService.ExecuteSpin(ctx, playerID, req.BetAmount)
    if err != nil {
        log.Error().
            Err(err).
            Str("player_id", playerID).
            Float64("bet_amount", req.BetAmount).
            Msg("Spin execution failed")
        return c.Status(500).JSON(ErrorResponse{
            Error: "spin_failed",
            Message: "Failed to execute spin",
        })
    }

    // 7. Log success
    log.Info().
        Str("player_id", playerID).
        Float64("bet_amount", req.BetAmount).
        Float64("win_amount", spin.WinAmount).
        Msg("Spin executed successfully")

    return c.JSON(SpinResponse{
        SpinID:    spin.ID,
        WinAmount: spin.WinAmount,
        Result:    spin.Result,
    })
}
```

### Complete Service Example

```go
package service

import (
    "context"
    "github.com/google/uuid"
)

func (s *SpinService) ExecuteSpin(ctx context.Context, playerID uuid.UUID, betAmount float64) (*Spin, error) {
    // 1. Create traced logger from context
    log := s.logger.WithTraceContext(ctx)

    log.Info().
        Str("player_id", playerID.String()).
        Float64("bet_amount", betAmount).
        Msg("Executing spin")

    // 2. Get player (SQL logged with trace info)
    player, err := s.playerRepo.GetByID(ctx, playerID)
    if err != nil {
        log.Error().Err(err).Msg("Failed to get player")
        return nil, err
    }

    // 3. Validate balance
    if player.Balance < betAmount {
        log.Warn().
            Float64("balance", player.Balance).
            Float64("bet_amount", betAmount).
            Msg("Insufficient balance")
        return nil, ErrInsufficientBalance
    }

    // 4. Execute game logic
    result := s.gameEngine.Spin(betAmount)

    // 5. Save spin (SQL logged with trace info)
    spin := &Spin{
        PlayerID:  playerID,
        BetAmount: betAmount,
        WinAmount: result.WinAmount,
        Result:    result,
    }

    if err := s.spinRepo.Create(ctx, spin); err != nil {
        log.Error().Err(err).Msg("Failed to save spin")
        return nil, err
    }

    log.Info().
        Str("spin_id", spin.ID.String()).
        Float64("win_amount", spin.WinAmount).
        Msg("Spin completed successfully")

    return spin, nil
}
```

## Searching Logs

### Find all requests for a specific traceID
```bash
cat logs.json | jq 'select(.trace_id == "550e8400-e29b-41d4-a716-446655440000")'
```

### Find all requests from a specific IP
```bash
cat logs.json | jq 'select(.client_ip == "192.168.1.100")'
```

### Find all slow SQL queries
```bash
cat logs.json | jq 'select(.message | contains("SLOW SQL"))'
```

### Find all errors for a specific trace
```bash
cat logs.json | jq 'select(.trace_id == "550e8400-..." and .level == "error")'
```

## Configuration

### Log Level (via .env)
```bash
LOG_LEVEL=debug  # debug, info, warn, error
LOG_FORMAT=json  # json, pretty, console
```

### Slow Query Threshold
Default: 200ms (configured in `internal/db/gorm.go`)

```go
customLogger := NewGormLogger(log, 200*time.Millisecond, gormLogLevel)
```

## Summary

✅ **Key Points**:
1. Every HTTP request gets a unique traceID
2. Client IP is automatically extracted
3. Use `log.WithTrace(c)` in handlers
4. Pass context to all service/repository calls
5. SQL queries automatically include traceID and IP
6. Slow queries (>200ms) are logged as warnings
7. All logs are structured JSON with trace info

This system makes it easy to:
- 🔍 Track requests end-to-end
- 🐛 Debug issues by traceID
- 📊 Analyze traffic by client IP
- ⚡ Identify slow SQL queries
- 🔒 Audit security events
