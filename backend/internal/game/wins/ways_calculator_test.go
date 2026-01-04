package wins

import (
	"testing"

	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
)

func TestCalculateWays_ThreeOfAKind(t *testing.T) {
	// Create a 5x10 grid with 3 matching symbols on first 3 reels
	// Rows 0-4: buffer rows (not checked for wins)
	// Rows 5-8: fully visible rows (checked for wins) - "fa" appears here in reels 0-2
	// Row 9: bottom partial row (not checked for wins)
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	// Should find "fa" with 3 consecutive reels (1 * 1 * 1 = 1 way)
	if len(wins) == 0 {
		t.Fatal("Expected at least one win, got none")
	}

	foundFa := false
	for _, win := range wins {
		if win.Symbol == symbols.SymbolFa {
			foundFa = true
			if win.Count != 3 {
				t.Errorf("Expected count 3 for 'fa', got %d", win.Count)
			}
			if win.Ways != 1 {
				t.Errorf("Expected 1 way for 'fa', got %d", win.Ways)
			}
		}
	}

	if !foundFa {
		t.Error("Expected to find 'fa' win")
	}
}

func TestCalculateWays_MultipleWaysPerSymbol(t *testing.T) {
	// Create a 5x10 grid where a symbol appears multiple times per reel
	// "fa" appears 2 times in rows 5-8 on reels 0-2 = 2*2*2 = 8 ways
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	// Should find "fa" with 2*2*2 = 8 ways
	foundFa := false
	for _, win := range wins {
		if win.Symbol == symbols.SymbolFa {
			foundFa = true
			if win.Count != 3 {
				t.Errorf("Expected count 3 for 'fa', got %d", win.Count)
			}
			if win.Ways != 8 {
				t.Errorf("Expected 8 ways for 'fa', got %d", win.Ways)
			}
		}
	}

	if !foundFa {
		t.Error("Expected to find 'fa' win with 8 ways")
	}
}

func TestCalculateWays_FiveOfAKind(t *testing.T) {
	// Create a 5x10 grid with all 5 reels having "zhong" in the visible rows
	grid := reels.Grid{
		{"cai", "fu", "shu", "fa", "liangtong", "zhong", "fa", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "fa", "liangtong", "zhong", "cai", "fu", "shu", "fa"},
		{"cai", "fu", "shu", "fa", "liangtong", "zhong", "liangtong", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "fa", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "fa", "liangtong", "zhong", "cai", "fu", "shu", "fa"},
	}

	wins := CalculateWays(grid)

	// Should find "zhong" with 5 matching reels
	foundZhong := false
	for _, win := range wins {
		if win.Symbol == symbols.SymbolZhong {
			foundZhong = true
			if win.Count != 5 {
				t.Errorf("Expected count 5 for 'zhong', got %d", win.Count)
			}
			if win.Ways != 1 {
				t.Errorf("Expected 1 way for 'zhong', got %d", win.Ways)
			}
		}
	}

	if !foundZhong {
		t.Error("Expected to find 'zhong' 5-of-a-kind win")
	}
}

func TestCalculateWays_NoWin(t *testing.T) {
	// Create a 5x10 grid with no consecutive matches in visible rows (5-8)
	// Ensure reel 0 and reel 1 have different symbols to break the chain
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "bawan", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "liangsuo"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "fa", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "cai", "liangtong", "fa", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fu", "cai", "liangtong", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	if len(wins) != 0 {
		t.Errorf("Expected no wins, got %d wins: %+v", len(wins), wins)
	}
}

func TestCalculateWays_WildSubstitution(t *testing.T) {
	// Create a 5x10 grid with wild symbol that should substitute
	// Reel 0: "fa", Reel 1: "wild", Reel 2: "fa" = 3-of-a-kind with wild substitution
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "wild", "cai", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	// Wild should substitute for "fa" creating 3-of-a-kind
	foundFa := false
	for _, win := range wins {
		if win.Symbol == symbols.SymbolFa {
			foundFa = true
			if win.Count < 3 {
				t.Errorf("Expected count >= 3 for 'fa' with wild, got %d", win.Count)
			}
		}
	}

	if !foundFa {
		t.Error("Expected to find 'fa' win with wild substitution")
	}
}

func TestCalculateWays_SkipsBonus(t *testing.T) {
	// Create a 5x10 grid with bonus (scatter) symbols
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bonus", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "cai", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "liangtong", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	// Bonus symbols should not create wins (they're scatters)
	for _, win := range wins {
		if win.Symbol == symbols.SymbolBonus {
			t.Errorf("Bonus symbol should not create ways wins, found: %+v", win)
		}
	}
}

func TestCalculateWays_GoldVariants(t *testing.T) {
	// Create a 5x10 grid with gold variants
	// Gold variants should match with regular symbols
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "zhong", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "liangtong", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	wins := CalculateWays(grid)

	// Gold variants should match with regular symbols
	foundFa := false
	for _, win := range wins {
		if win.Symbol == symbols.SymbolFa {
			foundFa = true
			if win.Count < 3 {
				t.Errorf("Expected count >= 3 for 'fa' (including gold), got %d", win.Count)
			}
		}
	}

	if !foundFa {
		t.Error("Expected to find 'fa' win including gold variants")
	}
}

func TestGetWinningSymbols(t *testing.T) {
	t.Run("should return empty slice for grid with no wins", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "bawan", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "liangsuo"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "cai", "liangtong", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fu", "cai", "liangtong", "shu", "zhong"},
		}
		result := GetWinningSymbols(grid)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d symbols", len(result))
		}
	})

	t.Run("should return winning symbols from grid", func(t *testing.T) {
		// Grid with "fa" appearing on first 3 reels
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}
		result := GetWinningSymbols(grid)

		// Should have at least one winning symbol
		if len(result) == 0 {
			t.Error("Expected at least one winning symbol")
		}

		// Should contain "fa"
		found := false
		for _, sym := range result {
			if sym == symbols.SymbolFa {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'fa' symbol in winning symbols")
		}
	})
}

func TestHasAnyWins(t *testing.T) {
	t.Run("should return false for grid with no wins", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "bawan", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "liangsuo"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "cai", "liangtong", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fu", "cai", "liangtong", "shu", "zhong"},
		}
		result := HasAnyWins(grid)
		if result {
			t.Error("Expected false for grid with no wins")
		}
	})

	t.Run("should return true for grid with wins", func(t *testing.T) {
		// Grid with "fa" appearing on first 3 reels
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}
		result := HasAnyWins(grid)
		if !result {
			t.Error("Expected true for grid with wins")
		}
	})

	t.Run("should return true for grid with multiple wins", func(t *testing.T) {
		// Grid with multiple winning symbols
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}
		result := HasAnyWins(grid)
		if !result {
			t.Error("Expected true for grid with multiple wins")
		}
	})
}

func BenchmarkCalculateWays(b *testing.B) {
	// Create a typical 5x10 grid
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateWays(grid)
	}
}
