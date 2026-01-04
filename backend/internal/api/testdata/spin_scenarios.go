package testdata

import (
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/internal/api/dto"
)

// Symbol IDs for reference:
// 0: wild, 1: bonus, 2: fa, 3: zhong, 4: bai, 5: bawan
// 6: wusuo, 7: wutong, 8: liangsuo, 9: liangtong

// SpinScenarios provides test spin responses for various scenarios
// Usage: testdata.SpinScenarios.BonusCascadeTrigger()
var SpinScenarios = &spinScenarios{}

type spinScenarios struct{}

// BonusCascadeTrigger returns a response where:
// - Initial grid has 2 bonus tiles
// - 4 'fa' tiles create a winning combination
// - After cascade, 3rd bonus tile appears
// - free_spins_triggered = true
func (s *spinScenarios) BonusCascadeTrigger() *dto.SpinResponse {
	return getBonusCascadeTriggerResponse(uuid.New())
}

// SimpleBonusTrigger returns a response where:
// - 3 bonus tiles land directly
// - No cascade wins
// - free_spins_triggered = true
func (s *spinScenarios) SimpleBonusTrigger() *dto.SpinResponse {
	return getSimpleBonusTriggerResponse(uuid.New())
}

// BigWin returns a response where:
// - 5 'fa' tiles across create a big win
// - Single cascade
func (s *spinScenarios) BigWin() *dto.SpinResponse {
	return getBigWinResponse(uuid.New())
}

// LastFreeSpinNoCascade returns a response where:
// - This is the last free spin (free_spins_remaining = 0)
// - No wins, no cascade
// - Should show final jackpot result
func (s *spinScenarios) LastFreeSpinNoCascade() *dto.SpinResponse {
	return getLastFreeSpinNoCascadeResponse(uuid.New())
}

// LastFreeSpinWithCascade returns a response where:
// - This is the last free spin (free_spins_remaining = 0)
// - Has winning combination with cascade
// - Should show final jackpot result after cascade completes
func (s *spinScenarios) LastFreeSpinWithCascade() *dto.SpinResponse {
	return getLastFreeSpinWithCascadeResponse(uuid.New())
}

// getBonusCascadeTriggerResponse creates response for bonus cascade trigger scenario
// Scenario:
// - Initial grid has 2 bonus tiles at [0,5] and [2,7]
// - 4 'fa' tiles at row 6 create a winning combination
// - After cascade, 3rd bonus tile appears at [1,6]
// - free_spins_triggered = true (backend correctly detects 3 bonus after cascade)
func getBonusCascadeTriggerResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       1.0,
		BalanceBefore:   10000,
		BalanceAfterBet: 9999,
		NewBalance:      10049,
		// Initial grid: 5 columns x 10 rows
		// Rows 0-4 are buffer, rows 5-8 are visible (win check area)
		Grid: [][]int{
			{9, 9, 9, 9, 9, 1, 2, 9, 9, 9}, // col 0: bonus at row 5, fa at row 6
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 1: fa at row 6
			{9, 9, 9, 9, 9, 9, 2, 1, 9, 9}, // col 2: fa at row 6, bonus at row 7
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 3: fa at row 6
			{9, 9, 9, 9, 9, 9, 5, 9, 9, 9}, // col 4: bawan at row 6 (no win)
		},
		Cascades: []dto.CascadeInfo{
			{
				CascadeNumber: 1,
				Multiplier:    1,
				Wins: []dto.WinInfo{
					{
						Symbol:    2, // fa
						Count:     4,
						Ways:      1,
						Payout:    50,
						WinAmount: 50,
						Positions: []dto.Position{
							{Reel: 0, Row: 6},
							{Reel: 1, Row: 6},
							{Reel: 2, Row: 6},
							{Reel: 3, Row: 6},
						},
						WinIntensity: "medium",
					},
				},
				TotalCascadeWin: 50,
				// After cascade: fa tiles removed, new tiles drop
				// 3rd bonus appears at [1,6]
				GridAfter: [][]int{
					{9, 9, 9, 9, 9, 1, 4, 9, 9, 9}, // col 0: bonus stays, bai dropped
					{9, 9, 9, 9, 9, 9, 1, 9, 9, 9}, // col 1: NEW BONUS at row 6!
					{9, 9, 9, 9, 9, 9, 5, 1, 9, 9}, // col 2: bawan dropped, bonus stays
					{9, 9, 9, 9, 9, 9, 3, 9, 9, 9}, // col 3: zhong dropped
					{9, 9, 9, 9, 9, 9, 5, 9, 9, 9}, // col 4: unchanged
				},
			},
		},
		SpinTotalWin:            50,
		ScatterCount:            3, // 3 bonus tiles after cascade
		IsFreeSpin:              false,
		FreeSpinsTriggered:      true, // Should trigger bonus mode!
		FreeSpinsRetriggered:    false,
		FreeSpinsSessionID:      uuid.New().String(),
		FreeSpinsRemainingSpins: 10,
		FreeSessionTotalWin:     0,
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

// getSimpleBonusTriggerResponse creates response for simple bonus trigger scenario
// 3 bonus tiles land directly, triggers bonus mode
func getSimpleBonusTriggerResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       1.0,
		BalanceBefore:   10000,
		BalanceAfterBet: 9999,
		NewBalance:      9999,
		Grid: [][]int{
			{9, 9, 9, 9, 9, 1, 9, 9, 9, 9}, // col 0: bonus at row 5
			{9, 9, 9, 9, 9, 9, 9, 9, 9, 9}, // col 1
			{9, 9, 9, 9, 9, 9, 1, 9, 9, 9}, // col 2: bonus at row 6
			{9, 9, 9, 9, 9, 9, 9, 9, 9, 9}, // col 3
			{9, 9, 9, 9, 9, 9, 9, 1, 9, 9}, // col 4: bonus at row 7
		},
		Cascades:                []dto.CascadeInfo{},
		SpinTotalWin:            0,
		ScatterCount:            3,
		IsFreeSpin:              false,
		FreeSpinsTriggered:      true,
		FreeSpinsRetriggered:    false,
		FreeSpinsSessionID:      uuid.New().String(),
		FreeSpinsRemainingSpins: 10,
		FreeSessionTotalWin:     0,
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

// getBigWinResponse creates response for big win scenario
func getBigWinResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       1.0,
		BalanceBefore:   10000,
		BalanceAfterBet: 9999,
		NewBalance:      10349,
		// 5 fa tiles across row 6
		Grid: [][]int{
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9},
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9},
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9},
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9},
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9},
		},
		Cascades: []dto.CascadeInfo{
			{
				CascadeNumber: 1,
				Multiplier:    1,
				Wins: []dto.WinInfo{
					{
						Symbol:    2,
						Count:     5,
						Ways:      1,
						Payout:    350,
						WinAmount: 350,
						Positions: []dto.Position{
							{Reel: 0, Row: 6},
							{Reel: 1, Row: 6},
							{Reel: 2, Row: 6},
							{Reel: 3, Row: 6},
							{Reel: 4, Row: 6},
						},
						WinIntensity: "big",
					},
				},
				TotalCascadeWin: 350,
				GridAfter: [][]int{
					{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
					{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
					{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
					{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
					{9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
				},
			},
		},
		SpinTotalWin:            350,
		ScatterCount:            0,
		IsFreeSpin:              false,
		FreeSpinsTriggered:      false,
		FreeSpinsRetriggered:    false,
		FreeSpinsRemainingSpins: 0,
		FreeSessionTotalWin:     0,
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

// getLastFreeSpinNoCascadeResponse creates response for last free spin with no cascade
// Scenario:
// - This is the last free spin (free_spins_remaining = 0)
// - No winning combinations, no cascade
// - IsFreeSpin = true (we're in free spin mode)
// - FreeSessionTotalWin shows accumulated wins from the session
func getLastFreeSpinNoCascadeResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       0, // Free spins don't deduct balance
		BalanceBefore:   10500,
		BalanceAfterBet: 10500, // No bet deducted
		NewBalance:      10500, // No win this spin
		// Random grid with no winning combinations
		Grid: [][]int{
			{9, 9, 9, 9, 9, 3, 5, 8, 9, 9}, // col 0: zhong, bawan, liangsuo
			{9, 9, 9, 9, 9, 7, 4, 6, 9, 9}, // col 1: wutong, bai, wusuo
			{9, 9, 9, 9, 9, 8, 3, 7, 9, 9}, // col 2: liangsuo, zhong, wutong
			{9, 9, 9, 9, 9, 5, 9, 4, 9, 9}, // col 3: bawan, liangtong, bai
			{9, 9, 9, 9, 9, 6, 8, 5, 9, 9}, // col 4: wusuo, liangsuo, bawan
		},
		Cascades:                []dto.CascadeInfo{}, // No cascades
		SpinTotalWin:            0,                   // No win this spin
		ScatterCount:            0,                   // No bonus tiles
		IsFreeSpin:              true,                // We're in free spin mode
		FreeSpinsTriggered:      false,               // Not triggering new free spins
		FreeSpinsRetriggered:    false,
		FreeSpinsSessionID:      sessionID.String(),
		FreeSpinsRemainingSpins: 0,   // This is the LAST spin
		FreeSessionTotalWin:     500, // Accumulated wins from the session
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

// FreeSpinRetrigger returns a response where:
// - This is a free spin that triggers a retrigger
// - 3 bonus tiles land during free spins
// - FreeSpinsRetriggered = true
// - FreeSpinsAdditional = 5 (extra spins awarded)
// - Use this to test the retrigger overlay multiple times
// - Every free spin will trigger retrigger, allowing you to test 2nd, 3rd, etc. retriggers
func (s *spinScenarios) FreeSpinRetrigger() *dto.SpinResponse {
	return getFreeSpinRetriggerResponse(uuid.New())
}

// getLastFreeSpinWithCascadeResponse creates response for last free spin with cascade
// Scenario:
// - This is the last free spin (free_spins_remaining = 0)
// - 4 'fa' tiles create a winning combination (50 win)
// - After cascade, no more wins
// - IsFreeSpin = true (we're in free spin mode)
// - FreeSessionTotalWin = 550 (previous 500 + this spin's 50)
func getLastFreeSpinWithCascadeResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       0, // Free spins don't deduct balance
		BalanceBefore:   10500,
		BalanceAfterBet: 10500, // No bet deducted
		NewBalance:      10550, // Won 50 this spin
		// 4 fa tiles across row 6
		Grid: [][]int{
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 0: fa at row 6
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 1: fa at row 6
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 2: fa at row 6
			{9, 9, 9, 9, 9, 9, 2, 9, 9, 9}, // col 3: fa at row 6
			{9, 9, 9, 9, 9, 9, 5, 9, 9, 9}, // col 4: bawan at row 6 (breaks the win)
		},
		Cascades: []dto.CascadeInfo{
			{
				CascadeNumber: 1,
				Multiplier:    1,
				Wins: []dto.WinInfo{
					{
						Symbol:    2, // fa
						Count:     4,
						Ways:      1,
						Payout:    50,
						WinAmount: 50,
						Positions: []dto.Position{
							{Reel: 0, Row: 6},
							{Reel: 1, Row: 6},
							{Reel: 2, Row: 6},
							{Reel: 3, Row: 6},
						},
						WinIntensity: "medium",
					},
				},
				TotalCascadeWin: 50,
				// After cascade: fa tiles removed, random tiles drop (no more wins)
				GridAfter: [][]int{
					{9, 9, 9, 9, 9, 9, 3, 9, 9, 9}, // col 0: zhong dropped
					{9, 9, 9, 9, 9, 9, 7, 9, 9, 9}, // col 1: wutong dropped
					{9, 9, 9, 9, 9, 9, 8, 9, 9, 9}, // col 2: liangsuo dropped
					{9, 9, 9, 9, 9, 9, 4, 9, 9, 9}, // col 3: bai dropped
					{9, 9, 9, 9, 9, 9, 5, 9, 9, 9}, // col 4: unchanged
				},
			},
		},
		SpinTotalWin:            50,   // Won 50 this spin
		ScatterCount:            0,    // No bonus tiles
		IsFreeSpin:              true, // We're in free spin mode
		FreeSpinsTriggered:      false,
		FreeSpinsRetriggered:    false,
		FreeSpinsSessionID:      sessionID.String(),
		FreeSpinsRemainingSpins: 0,   // This is the LAST spin
		FreeSessionTotalWin:     550, // Previous 500 + this spin's 50
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

// getFreeSpinRetriggerResponse creates response for free spin retrigger scenario
// Scenario:
// - During free spins, 3 bonus tiles land
// - FreeSpinsRetriggered = true
// - FreeSpinsAdditional = 5
// - This is the FIRST retrigger
func getFreeSpinRetriggerResponse(sessionID uuid.UUID) *dto.SpinResponse {
	return &dto.SpinResponse{
		SpinID:          uuid.New().String(),
		SessionID:       sessionID.String(),
		BetAmount:       0, // Free spins don't deduct balance
		BalanceBefore:   10500,
		BalanceAfterBet: 10500,
		NewBalance:      10500,
		// Grid with 3 bonus tiles (symbol 1) in visible area (rows 5-8)
		Grid: [][]int{
			{9, 9, 9, 9, 9, 1, 9, 9, 9, 9}, // col 0: bonus at row 5
			{9, 9, 9, 9, 9, 9, 3, 9, 9, 9}, // col 1: zhong at row 6
			{9, 9, 9, 9, 9, 9, 1, 9, 9, 9}, // col 2: bonus at row 6
			{9, 9, 9, 9, 9, 9, 9, 5, 9, 9}, // col 3: bawan at row 7
			{9, 9, 9, 9, 9, 9, 9, 1, 9, 9}, // col 4: bonus at row 7
		},
		Cascades:                []dto.CascadeInfo{}, // No winning combinations
		SpinTotalWin:            0,
		ScatterCount:            3, // 3 bonus tiles
		IsFreeSpin:              true,
		FreeSpinsTriggered:      false, // Not initial trigger
		FreeSpinsRetriggered:    true,  // This is a RETRIGGER
		FreeSpinsAdditional:     5,     // 5 extra spins awarded
		FreeSpinsSessionID:      sessionID.String(),
		FreeSpinsRemainingSpins: 8, // Had 3 remaining, now 3 + 5 = 8
		FreeSessionTotalWin:     200,
		Timestamp:               time.Now().Format(time.RFC3339),
	}
}

