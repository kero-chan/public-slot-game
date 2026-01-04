# Reel Strip Management System

## Overview

The new reel strip management system provides **deterministic, version-controlled reel strip assignment** instead of random selection. This enables:

- ✅ Release version management for reel strips
- ✅ Player-specific reel strip assignments (A/B testing, VIP configurations)
- ✅ No random selection - deterministic behavior
- ✅ Easy rollback and version control
- ✅ Separate configurations for base game and free spins

## Architecture

### 1. **ReelStrips Table** (existing)
Stores individual reel strip data (one strip per reel).

```sql
reel_strips (
  id, game_mode, reel_number, version,
  strip_data, checksum, strip_length, is_active
)
```

### 2. **ReelStripConfigs Table** (new)
Groups 5 reel strips into a named configuration.

```sql
reel_strip_configs (
  id, name, game_mode, description,
  reel_0_strip_id, reel_1_strip_id, reel_2_strip_id, reel_3_strip_id, reel_4_strip_id,
  version, target_rtp, is_active, is_default
)
```

**Example configurations:**
- `v1.0-standard` - Standard RTP (96.5%)
- `v1.0-high-rtp` - High RTP for VIP players (97.5%)
- `v2.0-default` - New version with adjusted volatility

### 3. **PlayerReelStripAssignments Table** (new)
Assigns specific configurations to specific players.

```sql
player_reel_strip_assignments (
  id, player_id, game_mode,
  base_game_config_id, free_spins_config_id,
  assigned_at, assigned_by, reason, expires_at
)
```

###4. **Players Table** (updated)
Added optional direct config references.

```sql
players (
  ...,
  base_game_config_id, -- Optional: Direct assignment
  free_spins_config_id -- Optional: Direct assignment
)
```

## How It Works

### Configuration Selection Priority

When a player spins, the system determines which reel strip configuration to use based on this priority:

1. **Player Assignment** (highest priority)
   - Check `player_reel_strip_assignments` table
   - If found and active → use that config

2. **Player Direct Assignment**
   - Check `players.base_game_config_id` or `free_spins_config_id`
   - If set → use that config

3. **Default Configuration**
   - Get config where `is_default = true` for the game mode
   - If found → use that config

4. **Fallback** (only if database is empty or error)
   - Generate reel strips on-the-fly (legacy behavior)

### No Random Selection

**Before (Old System):**
```sql
-- Random selection - different result each time
SELECT * FROM reel_strips WHERE game_mode = 'base_game' AND reel_number = 0 ORDER BY RANDOM() LIMIT 1;
```

**After (New System):**
```sql
-- Deterministic selection based on configuration
SELECT * FROM reel_strips WHERE id = (
  SELECT reel_0_strip_id FROM reel_strip_configs WHERE id = $config_id
);
```

## Usage Examples

### 1. Creating a Reel Strip Configuration

```go
// Step 1: Generate 5 reel strips
service.GenerateAndSaveStrips(ctx, "base_game", 1, 1) // Creates 5 strips

// Step 2: Create a configuration referencing these strips
config, err := service.CreateConfig(
    ctx,
    "v1.0-standard",           // name
    "base_game",               // game_mode
    "1.0",                     // version
    "Standard RTP 96.5%",      // description
    [5]uuid.UUID{              // reel strip IDs
        reel0ID, reel1ID, reel2ID, reel3ID, reel4ID,
    },
    96.50,                     // target RTP
)

// Step 3: Set as default
service.SetDefaultConfig(ctx, config.ID, "base_game")
```

### 2. Assigning Configuration to a Player

```go
// Assign "high RTP" config to VIP player
err := service.AssignConfigToPlayer(
    ctx,
    playerID,
    highRTPConfigID,
    "base_game",
    "VIP Player - High RTP Tier",
)
```

### 3. A/B Testing Example

```go
// Create two configs: A and B
configA := createConfig("v1.0-group-a", baseGameMode, ...)
configB := createConfig("v1.0-group-b", baseGameMode, ...)

// Assign 50% of players to each group
for _, player := range newPlayers {
    if player.ID.String()[0] < '8' { // Simple 50/50 split
        service.AssignConfigToPlayer(ctx, player.ID, configA.ID, "base_game", "A/B Test Group A")
    } else {
        service.AssignConfigToPlayer(ctx, player.ID, configB.ID, "base_game", "A/B Test Group B")
    }
}
```

### 4. Version Release Management

```go
// Release v2.0
newConfig := service.CreateConfig(ctx, "v2.0-default", "base_game", "2.0", ...)

// Gradually roll out
// Option 1: Set as default for new players
service.SetDefaultConfig(ctx, newConfig.ID, "base_game")

// Option 2: Migrate existing players gradually
migratePlayersToNewConfig(ctx, playerIDs, newConfig.ID)

// Option 3: Rollback if needed
service.SetDefaultConfig(ctx, oldConfig.ID, "base_game")
```

## Service Interface

### Main Methods

```go
// Get reel set for a player (handles all priority logic)
GetReelSetForPlayer(ctx, playerID, gameMode) (*ReelStripConfigSet, error)

// Configuration management
CreateConfig(ctx, name, gameMode, version, description, reelStripIDs, targetRTP) (*ReelStripConfig, error)
SetDefaultConfig(ctx, configID, gameMode) error
ActivateConfig(ctx, configID) error
DeactivateConfig(ctx, configID) error

// Player assignment management
AssignConfigToPlayer(ctx, playerID, configID, gameMode, reason) error
GetPlayerAssignment(ctx, playerID, gameMode) (*PlayerReelStripAssignment, error)
RemovePlayerAssignment(ctx, playerID, gameMode) error
```

## Migration Guide

### Running the Migration

```bash
# Apply migration
make migrate-up

# Or using migrate CLI
migrate -path ./migrations -database "postgres://..." up
```

### Seeding Initial Configuration

```bash
# 1. Generate reel strips
go run ./scripts/seed_reelstrips

# 2. Create default configuration (example script needed)
go run ./scripts/create_default_config
```

## Benefits

### 1. **Deterministic Behavior**
- Same player always gets same reel strips (unless explicitly reassigned)
- Reproducible spins for debugging
- Consistent player experience

### 2. **Version Control**
- Track reel strip changes over time
- Easy rollback to previous versions
- A/B test different configurations

### 3. **Player Segmentation**
- VIP players → high RTP config
- New players → standard config
- Problem gamblers → lower volatility config

### 4. **Compliance & Auditing**
- Clear audit trail of configuration changes
- Know exactly which player used which reel strips
- Regulatory compliance for RTP claims

### 5. **Flexible Testing**
- Test new configurations on small player groups
- Gradual rollout of changes
- Real-time configuration switching (no code deployment)

## Database Schema Diagram

```
┌─────────────────┐
│  reel_strips    │
│─────────────────│
│ id (PK)         │
│ game_mode       │
│ reel_number     │
│ strip_data      │
└─────────────────┘
         ▲
         │ (references)
         │
┌──────────────────────┐
│ reel_strip_configs   │
│──────────────────────│
│ id (PK)              │
│ name (unique)        │
│ reel_0_strip_id (FK) │──┐
│ reel_1_strip_id (FK) │  │
│ reel_2_strip_id (FK) │  │ References 5 strips
│ reel_3_strip_id (FK) │  │
│ reel_4_strip_id (FK) │──┘
│ is_default           │
└──────────────────────┘
         ▲
         │ (references)
         │
┌────────────────────────────────┐
│ player_reel_strip_assignments  │
│────────────────────────────────│
│ id (PK)                        │
│ player_id (FK)                 │
│ base_game_config_id (FK)       │
│ free_spins_config_id (FK)      │
│ reason                         │
└────────────────────────────────┘
         │
         │ (references)
         ▼
┌─────────────────┐
│    players      │
│─────────────────│
│ id (PK)         │
│ username        │
│ base_game_config_id (FK)  │ (optional direct assignment)
│ free_spins_config_id (FK) │ (optional direct assignment)
└─────────────────┘
```

## Next Steps

1. ✅ Run migration to create new tables
2. ⬜ Implement repository methods for new tables
3. ⬜ Implement service methods
4. ⬜ Update game engine to use player-specific configs
5. ⬜ Create admin API endpoints for config management
6. ⬜ Create script to seed initial default configuration
7. ⬜ Update documentation and tests

## FAQ

**Q: What happens if a player has no assignment?**
A: They use the default configuration (where `is_default = true`).

**Q: Can I have different configs for base game vs free spins?**
A: Yes! Each player can have separate `base_game_config_id` and `free_spins_config_id`.

**Q: How do I roll back to an old version?**
A: Simply set the old configuration as default: `SetDefaultConfig(oldConfigID, gameMode)`.

**Q: Can I temporarily assign a config to a player?**
A: Yes! Use the `expires_at` field in player assignments.

**Q: Is the old random selection still available?**
A: Yes, as a fallback, but it's deprecated. The system will warn when it's used.
