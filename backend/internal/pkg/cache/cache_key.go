package cache

import (
	"fmt"

	"github.com/google/uuid"
)

func (c *Cache) DefaultReelStripConfig(gameMode string) string {
	return c.setKey("defaultReelStripConfig:%s", gameMode)
}

func (c *Cache) ReelStripConfigById(id uuid.UUID) string {
	return c.setKey("ReelStripConfigId:%s", id.String())
}

func (c *Cache) ReelStripConfigSetKey(id uuid.UUID) string {
	return c.setKey("ReelStripConfigSet:%s", id.String())
}

func (c *Cache) ReelStripsDefaultKey(gameMode string) string {
	return c.setKey("reelStripsDefault:%s", gameMode)
}

func (c *Cache) ReelStripsByConfigIdKey(id uuid.UUID) string {
	return c.setKey("reelStripsByConfigId:%s", id.String())
}

func (c *Cache) PlayerAssignmentKey(playerID uuid.UUID) string {
	return c.setKey("playerAssignment:%s", playerID.String())
}

func (c *Cache) setKey(format string, a ...any) string {
	originKey := fmt.Sprintf(format, a...)

	return fmt.Sprintf("%s:%s:%s", c.config.App.Name, c.config.App.Env, originKey)
}
