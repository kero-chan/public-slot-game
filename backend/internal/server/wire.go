package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ProviderSet is the Wire provider set for server
var ProviderSet = wire.NewSet(
	ProvideFiberApp,
)

// ProvideFiberApp creates a new Fiber application
func ProvideFiberApp(cfg *config.Config, log *logger.Logger) *fiber.App {
	return NewFiberApp(cfg, log)
}
