package storage

import (
	"fmt"

	"github.com/google/wire"
	"github.com/slotmachine/backend/internal/config"
)

// ProviderSet is the Wire provider set for storage
var ProviderSet = wire.NewSet(
	ProvideStorage,
)

// ProvideStorage provides the appropriate storage implementation based on config
func ProvideStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Provider {
	case "gcs":
		return NewGCSStorage(&cfg.Storage)
	case "minio", "":
		return NewMinIOStorage(&cfg.Storage)
	default:
		return nil, fmt.Errorf("unknown storage provider: %s", cfg.Storage.Provider)
	}
}
