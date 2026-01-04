package dto

// CreateGameRequest is the request body for creating a game
type CreateGameRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	DevURL      *string `json:"dev_url"`
	ProdURL     *string `json:"prod_url"`
	IsActive    bool    `json:"is_active"`
}

// UpdateGameRequest is the request body for updating a game
type UpdateGameRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	DevURL      *string `json:"dev_url"`
	ProdURL     *string `json:"prod_url"`
	IsActive    *bool   `json:"is_active"`
}

// CreateAssetRequest is the request body for creating an asset
type CreateAssetRequest struct {
	Name            string         `json:"name"`
	Description     *string        `json:"description"`
	ObjectName      string         `json:"object_name"`
	SpritesheetJSON map[string]any `json:"spritesheet_json"`
	Images          map[string]any `json:"images"`
	Audios          map[string]any `json:"audios"`
	Videos          map[string]any `json:"videos"`
	IsActive        bool           `json:"is_active"`
}

// UpdateAssetRequest is the request body for updating an asset
type UpdateAssetRequest struct {
	Name            *string        `json:"name"`
	Description     *string        `json:"description"`
	ObjectName      *string        `json:"object_name"`
	SpritesheetJSON map[string]any `json:"spritesheet_json"`
	Images          map[string]any `json:"images"`
	Audios          map[string]any `json:"audios"`
	Videos          map[string]any `json:"videos"`
	IsActive        *bool          `json:"is_active"`
}

// CreateGameConfigRequest is the request body for creating a game config
type CreateGameConfigRequest struct {
	GameID   string `json:"game_id"`
	AssetID  string `json:"asset_id"`
	IsActive bool   `json:"is_active"`
}
