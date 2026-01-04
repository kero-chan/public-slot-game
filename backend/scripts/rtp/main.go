package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- Paytable ---
type SymbolPay struct {
	Symbol string
	Pay3   int
	Pay4   int
	Pay5   int
}

type SymbolPosition struct {
	IsTransformToGold bool
	Reel              int
	Row               int
}

// --- Spin Result ---
type SpinResult struct {
	Win       float64
	FreeSpins int
}

const (
	rows               = 4
	numReels           = 5
	reelLength         = 1000
	targetRTP          = 0.96
	tolerance          = 0.005
	simSpins           = 1_000_000 // giảm để test nhanh, tăng để chính xác hơn
	maxIter            = 20
	betAmount  float64 = 10
)

var paytable = map[string]SymbolPay{
	"fa":        {"fa", 10, 25, 50},
	"zhong":     {"zhong", 8, 20, 40},
	"bai":       {"bai", 6, 15, 30},
	"bawan":     {"bawan", 5, 10, 15},
	"wusuo":     {"wusuo", 3, 5, 12},
	"wutong":    {"wutong", 3, 5, 12},
	"liangsuo":  {"liangsuo", 2, 4, 10},
	"liangtong": {"liangtong", 1, 3, 6},
}

var symbolCountsMap = map[int]map[string]int{
	1: {
		"bai":       16,
		"bawan":     34,
		"fa":        16,
		"liangsuo":  249,
		"liangtong": 330,
		"scatter":   15,
		"wusuo":     238,
		"wutong":    86,
		"zhong":     16,
	},
	2: {
		"bai":            164,
		"bai_gold":       12,
		"bawan":          156,
		"bawan_gold":     15,
		"fa":             115,
		"fa_gold":        15,
		"liangsuo":       67,
		"liangsuo_gold":  12,
		"liangtong":      53,
		"liangtong_gold": 8,
		"scatter":        4,
		"wusuo":          97,
		"wusuo_gold":     10,
		"wutong":         121,
		"wutong_gold":    11,
		"zhong":          126,
		"zhong_gold":     14,
	},
	3: {
		"bai":            43,
		"bai_gold":       13,
		"bawan":          69,
		"bawan_gold":     19,
		"fa":             16,
		"fa_gold":        5,
		"liangsuo":       163,
		"liangsuo_gold":  14,
		"liangtong":      202,
		"liangtong_gold": 20,
		"scatter":        24,
		"wusuo":          212,
		"wusuo_gold":     19,
		"wutong":         130,
		"wutong_gold":    13,
		"zhong":          23,
		"zhong_gold":     15,
	},
	4: {
		"bai":            140,
		"bai_gold":       15,
		"bawan":          121,
		"bawan_gold":     15,
		"fa":             87,
		"fa_gold":        18,
		"liangsuo":       95,
		"liangsuo_gold":  14,
		"liangtong":      77,
		"liangtong_gold": 13,
		"scatter":        14,
		"wusuo":          125,
		"wusuo_gold":     14,
		"wutong":         123,
		"wutong_gold":    12,
		"zhong":          105,
		"zhong_gold":     12,
	},
	5: {
		"bai":       73,
		"bawan":     103,
		"fa":        34,
		"liangsuo":  198,
		"liangtong": 223,
		"scatter":   5,
		"wusuo":     194,
		"wutong":    116,
		"zhong":     54,
	},
}

// check high pay symbol for tuner
func highPaySymbol(sym string) bool {
	// treat gold variants as not high-pay for tuning moves
	if strings.HasSuffix(sym, "_gold") {
		return false
	}
	switch sym {
	case "liangtong", "liangsuo", "wusuo", "scatter":
		return false
	default:
		return true
	}
}

func generateReels(counts map[int]map[string]int) [][]string {
	lowPaySymbols := []string{"liangtong", "liangsuo", "wusuo"}
	reels := make([][]string, numReels)

	for r := range reels {
		strip := []string{}
		total := 0

		// Build strip từ symbolCountsMap
		for sym, count := range counts[r+1] {
			for i := 0; i < count; i++ {
				strip = append(strip, sym)
			}
			total += count
		}

		// Fill phần còn lại bằng low-pay
		for i := total; i < reelLength; i++ {
			strip = append(strip, lowPaySymbols[i%len(lowPaySymbols)])
		}

		// Shuffle
		rand.Shuffle(len(strip), func(i, j int) { strip[i], strip[j] = strip[j], strip[i] })

		// CẮT strip về đúng kích thước
		if len(strip) > reelLength {
			strip = strip[:reelLength]
		}

		// —— IMPORTANT ——
		// Cập nhật lại symbolCountsMap theo strip thực tế
		newCounts := map[string]int{}
		for _, sym := range strip {
			newCounts[sym]++
		}
		counts[r+1] = newCounts

		// Gán vào reels
		reels[r] = strip
	}

	return reels
}

// --- RTP Simulation ---
func simulateRTP(reels [][]string) float64 {
	totalWin := 0.0
	totalBet := float64(simSpins) * betAmount
	for s := 0; s < simSpins; s++ {
		reelStops := map[int]int{}
		for r := 0; r < numReels; r++ {
			reelStops[r] = rand.Intn(len(reels[r]))
		}
		grid := make([][]string, numReels)
		for r := range grid {
			grid[r] = []string{}
		}
		spinResult := &SpinResult{Win: 0, FreeSpins: 0}
		spinWays(spinResult, reels, reelStops, paytable, grid, 1, betAmount, false)
		totalWin += spinResult.Win
	}
	return totalWin / totalBet
}

func printSymbolDistribution(symbolCountsMap map[int]map[string]int) {
	fmt.Println("=== Symbol Distribution (%) ===")

	for reel, symbols := range symbolCountsMap {
		total := 0
		for _, c := range symbols {
			total += c
		}

		fmt.Printf("Reel %d (total=%d):\n", reel, total)
		for sym, c := range symbols {
			pct := float64(c) / float64(total) * 100
			fmt.Printf("  %-10s : %6.2f%%   (count=%d)\n", sym, pct, c)
		}
		fmt.Println()
	}
}

func writeReelsToCSV(reels [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row
	header := []string{"Position"}
	for i := 1; i <= numReels; i++ {
		header = append(header, "Reel"+strconv.Itoa(i))
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write reel data row by row
	for pos := 0; pos < reelLength; pos++ {
		row := []string{strconv.Itoa(pos)}
		for r := 0; r < numReels; r++ {
			if pos < len(reels[r]) {
				row = append(row, reels[r][pos])
			} else {
				row = append(row, "")
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", pos, err)
		}
	}

	return nil
}

func autoTune() {
	for iter := 0; iter < maxIter; iter++ {
		startTime := time.Now()
		reels := generateReels(symbolCountsMap)
		rtpEst := simulateRTP(reels)
		finishDuration := time.Since(startTime)
		fmt.Printf("run speed: %0.2f spin/s; ", float64(simSpins)/float64(finishDuration.Seconds()))
		fmt.Printf("Iter %d: RTP=%.5f\n", iter, rtpEst)
		if math.Abs(rtpEst-targetRTP) < tolerance {
			fmt.Println("Target achieved")

			// Write reels to CSV file
			filename := fmt.Sprintf("reels_rtp_%.5f_%s.csv", rtpEst, time.Now().Format("20060102_150405"))
			if err := writeReelsToCSV(reels, filename); err != nil {
				fmt.Printf("Error writing CSV file: %v\n", err)
			} else {
				fmt.Printf("Reels successfully written to %s\n", filename)
			}

			break
		}

		// adjust counts: step size 1 for high-pay, in reversed direction if overshoot
		for r := 1; r <= numReels; r++ {
			for sym := range symbolCountsMap[r] {
				// do not touch gold variants here, or we can tune separately
				if highPaySymbol(sym) {
					if rtpEst > targetRTP {
						if r%2 == 1 {
							if symbolCountsMap[r][sym] > 15 {
								symbolCountsMap[r][sym]--
							}
						} else {
							symbolCountsMap[r][sym]++
						}
					} else {
						if r%2 == 1 {
							symbolCountsMap[r][sym]++
						} else {
							if symbolCountsMap[r][sym] > 15 {
								symbolCountsMap[r][sym]--
							}
						}
					}
				}
			}
		}
	}

	out, _ := json.Marshal(symbolCountsMap)
	printSymbolDistribution(symbolCountsMap)
	fmt.Println("Final Output: ", string(out))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	autoTune()

	fmt.Println("Final symbol counts:", symbolCountsMap)
}

// --- Spin function with corrected ways to win ---
func spinWays(result *SpinResult, reels [][]string, reelStops map[int]int, paytable map[string]SymbolPay, grid [][]string, cascade int, betAmount float64, isFreeSpin bool) {
	if cascade > 3 {
		cascade = 5
	}
	// Random stop
	for r := range numReels {
		reel := reels[r]
		newGridReel := []string{}
		for _, symbol := range grid[r] {
			if symbol != "" {
				newGridReel = append(newGridReel, symbol)
			}
		}
		symbolCount := len(newGridReel)
		for range rows - symbolCount {
			reelStops[r] = reelStops[r] - 1
			if reelStops[r] < 0 {
				reelStops[r] += len(reel)
			}
			newGridReel = append([]string{reel[reelStops[r]]}, newGridReel...)
		}
		grid[r] = newGridReel
	}

	removeLists := []SymbolPosition{}
	showedSymbolList := map[string]map[int][]SymbolPosition{}

	for r := range numReels {
		for row := range rows {
			symbolSplit := grid[r][row]
			symbols := strings.Split(symbolSplit, "_")
			symbol := symbols[0]
			if _, ok := showedSymbolList[symbol]; !ok {
				showedSymbolList[symbol] = map[int][]SymbolPosition{
					r: {SymbolPosition{Reel: r, Row: row, IsTransformToGold: len(symbols) > 1}},
				}
			} else {
				showedSymbolList[symbol][r] = append(showedSymbolList[symbol][r], SymbolPosition{Reel: r, Row: row, IsTransformToGold: len(symbols) > 1})
			}
		}
	}

	var goldReel map[int][]SymbolPosition
	if _, ok := showedSymbolList["gold"]; ok {
		goldReel = showedSymbolList["gold"]
	} else {
		goldReel = map[int][]SymbolPosition{}
	}

	for symbol, reelMap := range showedSymbolList {
		if symbol == "scatter" {
			if len(reelMap) >= 3 {
				result.FreeSpins += 12 + (len(reelMap)-3)*2
				for _, removePosition := range reelMap {
					removeLists = append(removeLists, removePosition...)
				}
			}
		} else if symbol != "gold" {
			if _, ok := reelMap[0]; !ok {
				continue
			}

			// symbol is on reel 1
			consec := 0
			winWays := len(reelMap[0]) * (len(reelMap[1]) + len(goldReel[1])) * (len(reelMap[2]) + len(goldReel[2]))
			if winWays > 0 {
				consec = 3
				removeLists = append(removeLists, reelMap[0]...)
				removeLists = append(removeLists, reelMap[1]...)
				removeLists = append(removeLists, goldReel[1]...)
				removeLists = append(removeLists, reelMap[2]...)
				removeLists = append(removeLists, goldReel[2]...)
				if (len(reelMap[3]) + len(goldReel[3])) > 0 {
					consec = 4
					winWays = winWays * (len(reelMap[3]) + len(goldReel[3]))
					removeLists = append(removeLists, reelMap[3]...)
					removeLists = append(removeLists, goldReel[3]...)

					if (len(reelMap[4]) + len(goldReel[4])) > 0 {
						consec = 5
						winWays = winWays * len(reelMap[4])
						removeLists = append(removeLists, reelMap[4]...)
						removeLists = append(removeLists, goldReel[4]...)
					}
				}
			}

			if consec >= 3 {
				if sp, ok := paytable[symbol]; ok {
					pay := 0.0
					switch consec {
					case 3:
						pay = float64(sp.Pay3)
					case 4:
						pay = float64(sp.Pay4)
					case 5:
						pay = float64(sp.Pay5)
					}
					if isFreeSpin {
						cascade *= 2
					}

					winAmount := pay * float64(winWays) * float64(cascade) * betAmount / 20
					result.Win += winAmount
				}
			}
		}
	}

	if len(removeLists) > 0 {
		for _, removePosition := range removeLists {
			if removePosition.IsTransformToGold {
				grid[removePosition.Reel][removePosition.Row] = "gold"
			} else {
				grid[removePosition.Reel][removePosition.Row] = ""
			}
		}

		spinWays(result, reels, reelStops, paytable, grid, cascade+1, betAmount, false)
	} else {
		if result.FreeSpins > 0 {
			result.FreeSpins--
			spinWays(result, reels, reelStops, paytable, grid, cascade, betAmount, true)
		} else {
			return
		}
	}
}
