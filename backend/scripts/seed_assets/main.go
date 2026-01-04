package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/internal/infra/storage"
)

// Fixed UUIDs for predictability (UUID v4 format)
var (
	// Game 1: Mahjong Ways (without videos)
	MahjongWays1ID = uuid.MustParse("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
	BambooAsset1ID = uuid.MustParse("b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e")
	GameConfig1ID  = uuid.MustParse("c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f")

	// Game 2: Mahjong Ways 2 (with videos)
	MahjongWays2ID = uuid.MustParse("d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a")
	BambooAsset2ID = uuid.MustParse("e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b")
	GameConfig2ID  = uuid.MustParse("f6a7b8c9-d0e1-4f2a-3b4c-5d6e7f8a9b0c")
)

// BambooSpritesheetJSON contains the spritesheet coordinates for bamboo theme
const BambooSpritesheetJSON = `{"background_reel.jpg":{"frame":{"x":0,"y":0,"w":755,"h":882}},"background_coins_top.png":{"frame":{"x":-755,"y":0,"w":757,"h":249}},"background_wood.png":{"frame":{"x":-755,"y":-249,"w":756,"h":98}},"background_top_icons.png":{"frame":{"x":-755,"y":-347,"w":756,"h":123}},"background_coins.png":{"frame":{"x":0,"y":-882,"w":756,"h":585}},"background_bottom.png":{"frame":{"x":-755,"y":-470,"w":756,"h":117}},"background_marquee_purple.png":{"frame":{"x":-755,"y":-587,"w":736,"h":137}},"background_marquee_green.png":{"frame":{"x":-755,"y":-724,"w":736,"h":102}},"background_marquee_red.png":{"frame":{"x":-756,"y":-882,"w":715,"h":89}},"active_status_bg.png":{"frame":{"x":-755,"y":-826,"w":117,"h":55}},"glyph_random_notification_4.png":{"frame":{"x":0,"y":0,"w":935,"h":67}},"glyph_random_notification_2.png":{"frame":{"x":0,"y":-67,"w":922,"h":69}},"glyph_random_notification_3.png":{"frame":{"x":0,"y":-136,"w":727,"h":82}},"glyph_spin_last.png":{"frame":{"x":0,"y":-218,"w":500,"h":108}},"glyph_1024_ways.png":{"frame":{"x":0,"y":-326,"w":438,"h":61}},"glyph_random_notification_1.png":{"frame":{"x":-500,"y":-218,"w":433,"h":60}},"glyph_jackpot_result_title.png":{"frame":{"x":0,"y":-387,"w":338,"h":178}},"glyph_free_spin_left.png":{"frame":{"x":-338,"y":-387,"w":300,"h":143}},"glyph_got_free_spins.png":{"frame":{"x":-727,"y":-136,"w":201,"h":29}},"glyph_total_win.png":{"frame":{"x":-638,"y":-387,"w":194,"h":70}},"glyph_x10_active.png":{"frame":{"x":0,"y":-565,"w":149,"h":76}},"glyph_win.png":{"frame":{"x":-638,"y":-457,"w":142,"h":70}},"glyph_4_gold_large.png":{"frame":{"x":0,"y":-641,"w":115,"h":140}},"glyph_9_gold_large.png":{"frame":{"x":-115,"y":-641,"w":108,"h":137}},"glyph_8_gold_large.png":{"frame":{"x":-223,"y":-641,"w":110,"h":137}},"glyph_6_gold_large.png":{"frame":{"x":-333,"y":-641,"w":108,"h":137}},"glyph_5_gold_large.png":{"frame":{"x":-441,"y":-641,"w":111,"h":137}},"glyph_0_gold_large.png":{"frame":{"x":-552,"y":-641,"w":113,"h":137}},"glyph_3_gold_large.png":{"frame":{"x":-665,"y":-641,"w":110,"h":136}},"glyph_dot_gold_large.png":{"frame":{"x":-775,"y":-641,"w":30,"h":135}},"glyph_comma_gold_large.png":{"frame":{"x":-805,"y":-641,"w":30,"h":135}},"glyph_2_gold_large.png":{"frame":{"x":0,"y":-781,"w":112,"h":135}},"glyph_1_gold_large.png":{"frame":{"x":-835,"y":-641,"w":77,"h":134}},"glyph_7_gold_large.png":{"frame":{"x":-112,"y":-781,"w":101,"h":132}},"glyph_x10_default.png":{"frame":{"x":-780,"y":-457,"w":131,"h":67}},"glyph_x4_active.png":{"frame":{"x":-213,"y":-781,"w":112,"h":79}},"glyph_x2_active.png":{"frame":{"x":-325,"y":-781,"w":112,"h":78}},"glyph_x6_active.png":{"frame":{"x":-935,"y":0,"w":109,"h":80}},"glyph_x5_active.png":{"frame":{"x":-935,"y":-80,"w":108,"h":79}},"glyph_x3_active.png":{"frame":{"x":-935,"y":-159,"w":108,"h":78}},"glyph_free_spin.png":{"frame":{"x":-935,"y":-237,"w":103,"h":29}},"glyph_x5_default.png":{"frame":{"x":-935,"y":-266,"w":102,"h":69}},"glyph_x4_default.png":{"frame":{"x":-935,"y":-335,"w":98,"h":68}},"glyph_x2_default.png":{"frame":{"x":-935,"y":-403,"w":98,"h":66}},"glyph_x2_auto_default.png":{"frame":{"x":-935,"y":-469,"w":98,"h":73}},"glyph_x6_default.png":{"frame":{"x":-935,"y":-542,"w":97,"h":71}},"glyph_x3_default.png":{"frame":{"x":-935,"y":-613,"w":96,"h":69}},"glyph_x1_active.png":{"frame":{"x":-935,"y":-682,"w":94,"h":75}},"glyph_x1_default.png":{"frame":{"x":-935,"y":-757,"w":93,"h":65}},"glyph_free_spin_start.png":{"frame":{"x":-935,"y":-822,"w":91,"h":48}},"glyph_congrats.png":{"frame":{"x":-935,"y":-870,"w":91,"h":41}},"glyph_paytable.png":{"frame":{"x":-727,"y":-165,"w":71,"h":23}},"glyph_dot_gold_small.png":{"frame":{"x":-438,"y":-326,"w":20,"h":60}},"glyph_comma_gold_small.png":{"frame":{"x":-458,"y":-326,"w":20,"h":60}},"glyph_9_gold_small.png":{"frame":{"x":-478,"y":-326,"w":46,"h":60}},"glyph_8_gold_small.png":{"frame":{"x":-524,"y":-326,"w":46,"h":60}},"glyph_7_gold_small.png":{"frame":{"x":-570,"y":-326,"w":46,"h":60}},"glyph_6_gold_small.png":{"frame":{"x":-616,"y":-326,"w":46,"h":60}},"glyph_5_gold_small.png":{"frame":{"x":-662,"y":-326,"w":46,"h":60}},"glyph_4_gold_small.png":{"frame":{"x":-708,"y":-326,"w":46,"h":60}},"glyph_3_gold_small.png":{"frame":{"x":-754,"y":-326,"w":46,"h":60}},"glyph_2_gold_small.png":{"frame":{"x":-800,"y":-326,"w":46,"h":60}},"glyph_1_gold_small.png":{"frame":{"x":-846,"y":-326,"w":34,"h":60}},"glyph_0_gold_small.png":{"frame":{"x":-880,"y":-326,"w":46,"h":60}},"glyph_x_gold_small.png":{"frame":{"x":-500,"y":-278,"w":48,"h":48}},"glyph_volume.png":{"frame":{"x":-798,"y":-165,"w":47,"h":23}},"glyph_history.png":{"frame":{"x":-845,"y":-165,"w":47,"h":23}},"glyph_rules.png":{"frame":{"x":-727,"y":-188,"w":46,"h":23}},"glyph_exit.png":{"frame":{"x":-773,"y":-188,"w":45,"h":23}},"glyph_close.png":{"frame":{"x":-818,"y":-188,"w":45,"h":23}},"icon_spin_circle_bg.png":{"frame":{"x":-200,"y":0,"w":180,"h":180}},"icon_spin_arrows_normal_blur.png":{"frame":{"x":0,"y":-200,"w":116,"h":116}},"icon_spin_arrows_normal.png":{"frame":{"x":-116,"y":-200,"w":116,"h":116}},"icon_spin_arrows_disabled.png":{"frame":{"x":-232,"y":-200,"w":116,"h":116}},"icon_plus.png":{"frame":{"x":-380,"y":0,"w":108,"h":110}},"icon_minus.png":{"frame":{"x":-380,"y":-110,"w":108,"h":110}},"icon_history.png":{"frame":{"x":-380,"y":-220,"w":102,"h":87}},"icon_turbo_bg.png":{"frame":{"x":0,"y":-316,"w":90,"h":92}},"icon_auto_spin_bg.png":{"frame":{"x":-90,"y":-316,"w":90,"h":91}},"icon_exit.png":{"frame":{"x":-180,"y":-316,"w":89,"h":82}},"icon_close.png":{"frame":{"x":-488,"y":0,"w":87,"h":89}},"icon_volume_on.png":{"frame":{"x":-488,"y":-89,"w":87,"h":71}},"icon_volume_off.png":{"frame":{"x":-488,"y":-160,"w":86,"h":74}},"icon_rules.png":{"frame":{"x":-488,"y":-234,"w":82,"h":85}},"icon_paytable.png":{"frame":{"x":-488,"y":-319,"w":82,"h":80}},"icon_auto_spin_arrow.png":{"frame":{"x":-269,"y":-316,"w":63,"h":66}},"icon_turbo.png":{"frame":{"x":-332,"y":-316,"w":43,"h":58}},"icon_win_amount.png":{"frame":{"x":-375,"y":-316,"w":57,"h":46}},"icon_menu.png":{"frame":{"x":-432,"y":-316,"w":48,"h":38}},"icon_wallet.png":{"frame":{"x":0,"y":-408,"w":46,"h":44}},"icon_coin.png":{"frame":{"x":-46,"y":-408,"w":44,"h":44}},"icon_auto_spin.png":{"frame":{"x":-348,"y":-200,"w":31,"h":36}},"icon_menu_mute.png":{"frame":{"x":-348,"y":-236,"w":26,"h":30}},"tile_zhong_gold.png":{"frame":{"x":0,"y":0,"w":480,"h":600}},"tile_zhong.png":{"frame":{"x":-480,"y":0,"w":480,"h":600}},"tile_wutong_gold.png":{"frame":{"x":-960,"y":0,"w":480,"h":600}},"tile_wutong.png":{"frame":{"x":0,"y":-600,"w":480,"h":600}},"tile_wusuo_gold.png":{"frame":{"x":-480,"y":-600,"w":480,"h":600}},"tile_wusuo.png":{"frame":{"x":-960,"y":-600,"w":480,"h":600}},"tile_liangtong_gold.png":{"frame":{"x":-1440,"y":0,"w":480,"h":600}},"tile_liangtong.png":{"frame":{"x":-1440,"y":-600,"w":480,"h":600}},"tile_liangsuo_gold.png":{"frame":{"x":0,"y":-1200,"w":480,"h":600}},"tile_liangsuo.png":{"frame":{"x":-480,"y":-1200,"w":480,"h":600}},"tile_wild.png":{"frame":{"x":-960,"y":-1200,"w":480,"h":600}},"tile_fa_gold.png":{"frame":{"x":-1440,"y":-1200,"w":480,"h":600}},"tile_fa.png":{"frame":{"x":-1920,"y":0,"w":480,"h":600}},"tile_bonus.png":{"frame":{"x":-1920,"y":-600,"w":480,"h":600}},"tile_bawan_gold.png":{"frame":{"x":-1920,"y":-1200,"w":480,"h":600}},"tile_bawan.png":{"frame":{"x":0,"y":-1800,"w":480,"h":600}},"tile_bai_gold.png":{"frame":{"x":-480,"y":-1800,"w":480,"h":600}},"tile_bai.png":{"frame":{"x":-960,"y":-1800,"w":480,"h":600}},"free_spins_overlay_bg.png":{"frame":{"x":0,"y":0,"w":756,"h":1051}},"win_small.png":{"frame":{"x":-756,"y":0,"w":820,"h":200}},"win_mega.png":{"frame":{"x":-756,"y":-200,"w":820,"h":200}},"win_medium.png":{"frame":{"x":-756,"y":-400,"w":820,"h":200}},"win_jackpot.png":{"frame":{"x":-756,"y":-600,"w":820,"h":200}},"win_grand.png":{"frame":{"x":-756,"y":-800,"w":820,"h":200}},"win_big.png":{"frame":{"x":0,"y":-1051,"w":820,"h":200}},"win_gold.png":{"frame":{"x":-820,"y":-1051,"w":141,"h":82}}}`

// BambooImagesJSON contains the image path mappings for bamboo theme
const BambooImagesJSON = `{
	"backgrounds": "images/backgrounds.png",
	"glyphs": "images/glyphs.png",
	"icons": "images/icons.png",
	"tiles": "images/tiles.png",
	"winAnnouncements": "images/winAnnouncements.png",
	"backgroundMain": "images/background/background.jpg",
	"backgroundStart": "images/background/background_start_game.jpg",
	"startBtn": "images/startScreen/start_btn.png",
	"preparing": "images/startScreen/preparing.png",
	"preparingSound": "images/startScreen/preparing_sound_system.png",
	"loadingResources": "images/startScreen/loading_resources.png",
	"loadingComplete": "images/startScreen/loading_resources_complete.png",
	"iconExit": "images/icons/icon_exit.png",
	"iconHistory": "images/icons/icon_history.png"
}`

// BambooAudiosJSON contains the audio path mappings for bamboo theme
const BambooAudiosJSON = `{
	"background_music": "audios/background_music.mp3",
	"background_music_jackpot": "audios/background_music_jackpot.mp3",
	"game_start": "audios/game_start.mp3",
	"jackpot_finalize": "audios/jackpot_finalize.m4a",
	"consecutive_wins_2x": "audios/consecutive_wins/2x.mp3",
	"consecutive_wins_3x": "audios/consecutive_wins/3x.mp3",
	"consecutive_wins_4x": "audios/consecutive_wins/4x.mp3",
	"consecutive_wins_5x": "audios/consecutive_wins/5x.mp3",
	"consecutive_wins_6x": "audios/consecutive_wins/6x.mp3",
	"consecutive_wins_10x": "audios/consecutive_wins/10x.mp3",
	"win_bai": "audios/wins/bai.mp3",
	"win_zhong": "audios/wins/zhong.mp3",
	"win_fa": "audios/wins/fa.mp3",
	"win_liangsuo": "audios/wins/liangsuo.mp3",
	"win_liangtong": "audios/wins/liangtong.mp3",
	"win_wusuo": "audios/wins/wusuo.mp3",
	"win_wutong": "audios/wins/wutong.mp3",
	"win_bawan": "audios/wins/bawan.mp3",
	"win_jackpot": "audios/wins/jackpot.mp3",
	"winning_announcement": "audios/wins/winning_announcement.mp3",
	"winning_highlight": "audios/wins/winning_highlight.mp3",
	"background_noises": [
		"audios/background_noise/noise_1.mp3",
		"audios/background_noise/noise_2.mp3",
		"audios/background_noise/noise_3.mp3",
		"audios/background_noise/noise_4.mp3",
		"audios/background_noise/noise_5.mp3",
		"audios/background_noise/noise_6.mp3",
		"audios/background_noise/noise_7.mp3",
		"audios/background_noise/noise_8.mp3",
		"audios/background_noise/noise_9.mp3",
		"audios/background_noise/noise_10.mp3",
		"audios/background_noise/noise_11.mp3"
	],
	"lot": "audios/effect/lot.m4a",
	"reel_spin": "audios/effect/reel_spin.m4a",
	"reel_spin_stop": "audios/effect/reel_spin_stop.m4a",
	"reach_bonus": "audios/effect/reach_bonus.m4a",
	"jackpot_start": "audios/effect/jackpot_start.mp3",
	"increase_bet": "audios/effect/increase_bet.mp3",
	"decrease_bet": "audios/effect/decrease_bet.mp3",
	"generic_ui": "audios/effect/generic_ui.mp3",
	"start_button": "audios/effect/start_button.mp3",
	"card_transition": "audios/effect/card_transition.mp3",
	"tile_break": "audios/effect/tile_break.mp3",
	"line_win": "audios/effect/line_win_sound.mp3"
}`

// BambooVideosJSON contains the video path mappings for bamboo theme
const BambooVideosJSON = `{
	"idle_loop": "videos/idle-loop.mp4",
	"win_small": "videos/win-small.mp4",
	"win_medium": "videos/win-medium.mp4",
	"win_big": "videos/win-big.mp4",
	"win_mega": "videos/win-mega.mp4",
	"win_jackpot": "videos/win-jackpot.mp4"
}`

func main() {
	// Initialize application with Wire
	application, err := InitializeSeedApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	log := application.Logger
	repo := application.GameRepository
	stor := application.Storage

	ctx := context.Background()

	log.Info().Msg("Starting game assets seed")
	startTime := time.Now()

	// Seed Game 1: Mahjong Ways (without videos)
	if err := seedGame1Assets(ctx, repo, stor); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed Game 1 assets")
	}

	// Seed Game 2: Mahjong Ways 2 (with videos)
	if err := seedGame2Assets(ctx, repo, stor); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed Game 2 assets")
	}

	duration := time.Since(startTime)

	log.Info().
		Dur("duration", duration).
		Msg("Game assets seed completed successfully")

	fmt.Println("\n=== Game Assets Seed Summary ===")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println("Games:")
	fmt.Println("  Game 1 (without videos):")
	fmt.Printf("    - Game: Mahjong Ways (ID: %s)\n", MahjongWays1ID)
	fmt.Printf("    - Theme: Bamboo Theme (ID: %s)\n", BambooAsset1ID)
	fmt.Printf("    - Config ID: %s\n", GameConfig1ID)
	fmt.Println("  Game 2 (with videos):")
	fmt.Printf("    - Game: Mahjong Ways 2 (ID: %s)\n", MahjongWays2ID)
	fmt.Printf("    - Theme: Bamboo Theme V2 (ID: %s)\n", BambooAsset2ID)
	fmt.Printf("    - Config ID: %s\n", GameConfig2ID)
	fmt.Println("================================")
}

// isAssetNotFoundError checks if the error is an asset not found error
func isAssetNotFoundError(err error) bool {
	return errors.Is(err, game.ErrAssetNotFound)
}

// isGameNotFoundError checks if the error is a game not found error
func isGameNotFoundError(err error) bool {
	return errors.Is(err, game.ErrGameNotFound)
}

// isGameConfigNotFoundError checks if the error is a game config not found error
func isGameConfigNotFoundError(err error) bool {
	return errors.Is(err, game.ErrGameConfigNotFound)
}

// seedGame1Assets seeds Game 1: Mahjong Ways with images and audios only (no videos)
func seedGame1Assets(ctx context.Context, repo game.Repository, stor storage.Storage) error {
	fmt.Println("\n--- Seeding Game 1: Mahjong Ways (without videos) ---")

	objectName1 := "bamboo-theme"

	// Check and create Asset 1
	_, err := repo.GetAssetByID(ctx, BambooAsset1ID)
	if err != nil {
		if isAssetNotFoundError(err) {
			// Create folder on storage
			if err := stor.CreateFolder(ctx, objectName1); err != nil {
				fmt.Printf("⚠️  Warning: Failed to create storage folder: %v\n", err)
			}

			// Get base URL from storage
			baseURL := stor.GetBaseURL(objectName1)

			description1 := "Mahjong bamboo theme with traditional tiles (no video background)"
			asset1 := &game.Asset{
				ID:              BambooAsset1ID,
				Name:            "Bamboo Theme",
				Description:     &description1,
				ObjectName:      objectName1,
				BaseURL:         baseURL,
				SpritesheetJSON: json.RawMessage(BambooSpritesheetJSON),
				Images:          json.RawMessage(BambooImagesJSON),
				Audios:          json.RawMessage(BambooAudiosJSON),
				Videos:          json.RawMessage(`{}`),
				IsActive:        true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := repo.CreateAsset(ctx, asset1); err != nil {
				return fmt.Errorf("failed to create asset 1: %w", err)
			}
			fmt.Printf("✅ Created asset: %s (ID: %s, ObjectName: %s, BaseURL: %s)\n", asset1.Name, asset1.ID, asset1.ObjectName, asset1.BaseURL)
		} else {
			return fmt.Errorf("failed to check asset 1: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Asset already exists: Bamboo Theme (ID: %s)\n", BambooAsset1ID)
	}

	// Check and create Game 1
	_, err = repo.GetGameByID(ctx, MahjongWays1ID)
	if err != nil {
		if isGameNotFoundError(err) {
			gameDesc1 := "Classic mahjong-themed slot game with 1024 ways to win"
			game1 := &game.Game{
				ID:          MahjongWays1ID,
				Name:        "Mahjong Ways",
				Description: &gameDesc1,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := repo.CreateGame(ctx, game1); err != nil {
				return fmt.Errorf("failed to create game 1: %w", err)
			}
			fmt.Printf("✅ Created game: %s (ID: %s)\n", game1.Name, game1.ID)
		} else {
			return fmt.Errorf("failed to check game 1: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Game already exists: Mahjong Ways (ID: %s)\n", MahjongWays1ID)
	}

	// Check and create GameConfig 1
	_, err = repo.GetGameConfigByID(ctx, GameConfig1ID)
	if err != nil {
		if isGameConfigNotFoundError(err) {
			config1 := &game.GameConfig{
				ID:        GameConfig1ID,
				GameID:    MahjongWays1ID,
				AssetID:   BambooAsset1ID,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := repo.CreateGameConfig(ctx, config1); err != nil {
				return fmt.Errorf("failed to create game config 1: %w", err)
			}
			fmt.Printf("✅ Created game config: Game %s -> Asset %s\n", MahjongWays1ID, BambooAsset1ID)
		} else {
			return fmt.Errorf("failed to check game config 1: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Game config already exists (ID: %s)\n", GameConfig1ID)
	}

	return nil
}

// seedGame2Assets seeds Game 2: Mahjong Ways 2 with images, audios, and videos
func seedGame2Assets(ctx context.Context, repo game.Repository, stor storage.Storage) error {
	fmt.Println("\n--- Seeding Game 2: Mahjong Ways 2 (with videos) ---")

	objectName2 := "kungfu-theme"

	// Check and create Asset 2
	_, err := repo.GetAssetByID(ctx, BambooAsset2ID)
	if err != nil {
		if isAssetNotFoundError(err) {
			// Create folder on storage
			if err := stor.CreateFolder(ctx, objectName2); err != nil {
				fmt.Printf("⚠️  Warning: Failed to create storage folder: %v\n", err)
			}

			// Get base URL from storage
			baseURL := stor.GetBaseURL(objectName2)

			description2 := "Mahjong kungfu theme with traditional tiles and video backgrounds"
			asset2 := &game.Asset{
				ID:              BambooAsset2ID,
				Name:            "Kungfu Theme",
				Description:     &description2,
				ObjectName:      objectName2,
				BaseURL:         baseURL,
				SpritesheetJSON: json.RawMessage(BambooSpritesheetJSON),
				Images:          json.RawMessage(BambooImagesJSON),
				Audios:          json.RawMessage(BambooAudiosJSON),
				Videos:          json.RawMessage(BambooVideosJSON),
				IsActive:        true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := repo.CreateAsset(ctx, asset2); err != nil {
				return fmt.Errorf("failed to create asset 2: %w", err)
			}
			fmt.Printf("✅ Created asset: %s (ID: %s, ObjectName: %s, BaseURL: %s)\n", asset2.Name, asset2.ID, asset2.ObjectName, asset2.BaseURL)
		} else {
			return fmt.Errorf("failed to check asset 2: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Asset already exists: Bamboo Theme V2 (ID: %s)\n", BambooAsset2ID)
	}

	// Check and create Game 2
	_, err = repo.GetGameByID(ctx, MahjongWays2ID)
	if err != nil {
		if isGameNotFoundError(err) {
			gameDesc2 := "Enhanced mahjong-themed slot game with 1024 ways to win and video backgrounds"
			game2 := &game.Game{
				ID:          MahjongWays2ID,
				Name:        "Mahjong Ways 2",
				Description: &gameDesc2,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := repo.CreateGame(ctx, game2); err != nil {
				return fmt.Errorf("failed to create game 2: %w", err)
			}
			fmt.Printf("✅ Created game: %s (ID: %s)\n", game2.Name, game2.ID)
		} else {
			return fmt.Errorf("failed to check game 2: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Game already exists: Mahjong Ways 2 (ID: %s)\n", MahjongWays2ID)
	}

	// Check and create GameConfig 2
	_, err = repo.GetGameConfigByID(ctx, GameConfig2ID)
	if err != nil {
		if isGameConfigNotFoundError(err) {
			config2 := &game.GameConfig{
				ID:        GameConfig2ID,
				GameID:    MahjongWays2ID,
				AssetID:   BambooAsset2ID,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := repo.CreateGameConfig(ctx, config2); err != nil {
				return fmt.Errorf("failed to create game config 2: %w", err)
			}
			fmt.Printf("✅ Created game config: Game %s -> Asset %s\n", MahjongWays2ID, BambooAsset2ID)
		} else {
			return fmt.Errorf("failed to check game config 2: %w", err)
		}
	} else {
		fmt.Printf("⏭️  Game config already exists (ID: %s)\n", GameConfig2ID)
	}

	return nil
}
