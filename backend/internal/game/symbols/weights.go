package symbols

import (
	"math"
	"strings"
)

// ReelWeights defines symbol weights for a reel (number of occurrences)
type ReelWeights map[string]int

var ReelStripLengths = map[int]int{
	1: 500,
	2: 500,
	3: 500,
	4: 500,
	5: 500,
}

var goldRates = map[int]float64{
	1: 0.0,
	2: 0.0,
	3: 0.0,
	4: 0.0,
	5: 0.0,
}

var BaseGameWeightRate = map[string]float64{
	// bonus symbol rate
	"bonus": 2, // rate 10% trigger free spins

	// High-value symbols
	"fa":    4.0,
	"zhong": 5.0,
	"bai":   7.0,
	"bawan": 9.0,

	// Low-value symbols
	"wusuo":     12.0,
	"wutong":    12.0,
	"liangsuo":  23.0,
	"liangtong": 23.0,
}

var FreeSpinWeightRate = map[string]float64{
	// bonus symbol rate
	"bonus": 2, // rate 10% trigger free spins

	// High-value symbols
	"fa":    8.0,
	"zhong": 9.0,
	"bai":   11.0,
	"bawan": 12.0,

	// Low-value symbols
	"wusuo":     13.0,
	"wutong":    14.0,
	"liangsuo":  16.0,
	"liangtong": 17.0,
}

// TrialWeightRate defines symbol weights for TRIAL mode with HUGE RTP
// This gives trial players a much better winning experience
var TrialWeightRate = map[string]float64{
	// bonus symbol rate - MUCH HIGHER for more free spins triggers
	"bonus": 6.0, // 2x base rate = more free spins!

	// wild symbol rate - HIGHER for more wins
	"gold": 15.0,

	// pay symbol rates - HEAVILY favor high-value symbols
	// High-value symbols (BOOSTED)
	"fa":    10.0, // 2.5x base rate
	"zhong": 12.0, // 2.4x base rate
	"bai":   14.0, // 2x base rate
	"bawan": 14.0, // 1.5x base rate

	// Low-value symbols (REDUCED)
	"wusuo":     10.0,
	"wutong":    10.0,
	"liangsuo":  14.0,
	"liangtong": 14.0,
}

// TrialFreeSpinWeightRate defines symbol weights for TRIAL free spins mode
// Even more generous than regular trial mode
var TrialFreeSpinWeightRate = map[string]float64{
	// bonus symbol rate - for retriggers
	"bonus": 4.0,

	// wild symbol rate - VERY HIGH for huge wins
	"gold": 20.0,

	// pay symbol rates - MEGA favor high-value symbols
	// High-value symbols (MEGA BOOSTED)
	"fa":    14.0,
	"zhong": 14.0,
	"bai":   14.0,
	"bawan": 14.0,

	// Low-value symbols (HEAVILY REDUCED)
	"wusuo":     8.0,
	"wutong":    8.0,
	"liangsuo":  10.0,
	"liangtong": 10.0,
}

// BaseGameWeights defines symbol weights for base game reels
var BaseGameWeights = struct {
	Reel1 ReelWeights // Outer reel (left)
	Reel2 ReelWeights // Middle reel with golden variants
	Reel3 ReelWeights // Center reel with golden variants
	Reel4 ReelWeights // Middle reel with golden variants
	Reel5 ReelWeights // Outer reel (right)
}{
	// // Reel 1 & 5 (Outer reels - no golden variants)
	Reel1: defaultReelWeights(false, 1), // Total: 1000
	// Reels 2, 3, 4 (Middle reels - with golden variants)
	Reel2: defaultReelWeights(false, 2), // Total: 1000
	Reel3: defaultReelWeights(false, 3), // Total: 1000
	Reel4: defaultReelWeights(false, 4), // Total: 1000
	// Reel 5 same as Reel 1 (outer reel)
	Reel5: defaultReelWeights(false, 5), // Total: 1000
}

// FreeSpinsWeights defines symbol weights for free spins mode
// Free spins have higher-value symbol frequencies for better wins
var FreeSpinsWeights = struct {
	Reel1 ReelWeights
	Reel2 ReelWeights
	Reel3 ReelWeights
	Reel4 ReelWeights
	Reel5 ReelWeights
}{
	// Free spins: increase high-value symbols, reduce low-value symbols
	Reel1: defaultReelWeights(true, 1), // Total: 1000

	Reel2: defaultReelWeights(true, 2), // Total: 1000

	Reel3: defaultReelWeights(true, 3), // Total: 1000

	Reel4: defaultReelWeights(true, 4), // Total: 1000

	Reel5: defaultReelWeights(true, 5), // Total: 1000
}

func GetBaseGameWinRate(symbol string) float64 {
	return BaseGameWeightRate[symbol]
}

func GetFreeSpinsWinRate(symbol string) float64 {
	return FreeSpinWeightRate[symbol]
}

// GetBaseGameWeights returns weights for a specific reel in base game
func GetBaseGameWeights(reelIndex int) ReelWeights {
	switch reelIndex {
	case 0:
		return BaseGameWeights.Reel1
	case 1:
		return BaseGameWeights.Reel2
	case 2:
		return BaseGameWeights.Reel3
	case 3:
		return BaseGameWeights.Reel4
	case 4:
		return BaseGameWeights.Reel5
	default:
		return BaseGameWeights.Reel1
	}
}

// GetFreeSpinsWeights returns weights for a specific reel in free spins
func GetFreeSpinsWeights(reelIndex int) ReelWeights {
	switch reelIndex {
	case 0:
		return FreeSpinsWeights.Reel1
	case 1:
		return FreeSpinsWeights.Reel2
	case 2:
		return FreeSpinsWeights.Reel3
	case 3:
		return FreeSpinsWeights.Reel4
	case 4:
		return FreeSpinsWeights.Reel5
	default:
		return FreeSpinsWeights.Reel1
	}
}

func getWeightNumber(isFreeMode bool, reelNumber int, symbol string) int {
	rates := BaseGameWeightRate
	if isFreeMode {
		rates = FreeSpinWeightRate
	}
	reelStripLength := ReelStripLengths[reelNumber]
	goldRate := goldRates[reelNumber]
	bonusCount := int(math.Round(rates["bonus"] * float64(reelStripLength) / 100))
	normalSymbol := reelStripLength - bonusCount
	symbolStr := strings.Split(symbol, "_")[0]
	if symbol == "bonus" {
		return bonusCount
	} else {
		weight := int(math.Round(rates[symbolStr] * float64(normalSymbol) / 100))
		goldWeight := int(math.Round(float64(weight) * (goldRate / 100)))
		if strings.HasSuffix(symbol, "_gold") {
			return goldWeight
		} else {
			return weight - goldWeight
		}
	}
}

func defaultReelWeights(isFreeMode bool, reelNumber int) ReelWeights {
	return ReelWeights{
		"bonus":          getWeightNumber(isFreeMode, reelNumber, "bonus"),
		"bai":            getWeightNumber(isFreeMode, reelNumber, "bai"),
		"bai_gold":       getWeightNumber(isFreeMode, reelNumber, "bai_gold"),
		"bawan":          getWeightNumber(isFreeMode, reelNumber, "bawan"),
		"bawan_gold":     getWeightNumber(isFreeMode, reelNumber, "bawan_gold"),
		"fa":             getWeightNumber(isFreeMode, reelNumber, "fa"),
		"fa_gold":        getWeightNumber(isFreeMode, reelNumber, "fa_gold"),
		"liangsuo":       getWeightNumber(isFreeMode, reelNumber, "liangsuo"),
		"liangsuo_gold":  getWeightNumber(isFreeMode, reelNumber, "liangsuo_gold"),
		"liangtong":      getWeightNumber(isFreeMode, reelNumber, "liangtong"),
		"liangtong_gold": getWeightNumber(isFreeMode, reelNumber, "liangtong_gold"),
		"wusuo":          getWeightNumber(isFreeMode, reelNumber, "wusuo"),
		"wusuo_gold":     getWeightNumber(isFreeMode, reelNumber, "wusuo_gold"),
		"wutong":         getWeightNumber(isFreeMode, reelNumber, "wutong"),
		"wutong_gold":    getWeightNumber(isFreeMode, reelNumber, "wutong_gold"),
		"zhong":          getWeightNumber(isFreeMode, reelNumber, "zhong"),
		"zhong_gold":     getWeightNumber(isFreeMode, reelNumber, "zhong_gold"),
	}
}

// Trial mode weight functions for HUGE RTP

// getTrialWeightNumber calculates weight number for trial mode symbols
func getTrialWeightNumber(isFreeMode bool, reelNumber int, symbol string) int {
	rates := TrialWeightRate
	if isFreeMode {
		rates = TrialFreeSpinWeightRate
	}
	bonusCount := int(math.Round(rates["bonus"] * float64(ReelStripLengths[reelNumber]) / 100))
	normalSymbol := ReelStripLengths[reelNumber] - bonusCount
	symbolStr := strings.Split(symbol, "_")[0]
	if symbol == "bonus" {
		return bonusCount
	} else {
		if reelNumber == 1 || reelNumber == 5 {
			if strings.HasSuffix(symbol, "_gold") {
				return 0
			} else {
				return int(math.Round(rates[symbolStr] * float64(normalSymbol) / 100))
			}
		} else {
			weight := int(math.Round(rates[symbolStr] * float64(normalSymbol) / 100))
			goldWeight := int(math.Round(float64(weight) * (rates["gold"] / 100)))
			if strings.HasSuffix(symbol, "_gold") {
				return goldWeight
			} else {
				return weight - goldWeight
			}
		}
	}
}

// trialReelWeights generates trial-specific reel weights with HUGE RTP
func trialReelWeights(isFreeMode bool, reelNumber int) ReelWeights {
	return ReelWeights{
		"bonus":          getTrialWeightNumber(isFreeMode, reelNumber, "bonus"),
		"bai":            getTrialWeightNumber(isFreeMode, reelNumber, "bai"),
		"bai_gold":       getTrialWeightNumber(isFreeMode, reelNumber, "bai_gold"),
		"bawan":          getTrialWeightNumber(isFreeMode, reelNumber, "bawan"),
		"bawan_gold":     getTrialWeightNumber(isFreeMode, reelNumber, "bawan_gold"),
		"fa":             getTrialWeightNumber(isFreeMode, reelNumber, "fa"),
		"fa_gold":        getTrialWeightNumber(isFreeMode, reelNumber, "fa_gold"),
		"liangsuo":       getTrialWeightNumber(isFreeMode, reelNumber, "liangsuo"),
		"liangsuo_gold":  getTrialWeightNumber(isFreeMode, reelNumber, "liangsuo_gold"),
		"liangtong":      getTrialWeightNumber(isFreeMode, reelNumber, "liangtong"),
		"liangtong_gold": getTrialWeightNumber(isFreeMode, reelNumber, "liangtong_gold"),
		"wusuo":          getTrialWeightNumber(isFreeMode, reelNumber, "wusuo"),
		"wusuo_gold":     getTrialWeightNumber(isFreeMode, reelNumber, "wusuo_gold"),
		"wutong":         getTrialWeightNumber(isFreeMode, reelNumber, "wutong"),
		"wutong_gold":    getTrialWeightNumber(isFreeMode, reelNumber, "wutong_gold"),
		"zhong":          getTrialWeightNumber(isFreeMode, reelNumber, "zhong"),
		"zhong_gold":     getTrialWeightNumber(isFreeMode, reelNumber, "zhong_gold"),
	}
}

// GetTrialWeights returns trial-specific weights for a reel with HUGE RTP
func GetTrialWeights(reelIndex int) ReelWeights {
	return trialReelWeights(false, reelIndex+1)
}

// GetTrialFreeSpinsWeights returns trial-specific weights for free spins with HUGE RTP
func GetTrialFreeSpinsWeights(reelIndex int) ReelWeights {
	return trialReelWeights(true, reelIndex+1)
}
