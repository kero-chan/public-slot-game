package tuningfreespin

import "github.com/slotmachine/backend/cmd/rtp-tuning/tuning"

func DefaultTopologies() [5]tuning.ReelTopology {
	return [5]tuning.ReelTopology{
		// Reel 1: Activator - triggers wins
		// Per base_rtp_tuning.md: Low 60-70%, Mid 25-30%, High <10%
		// Final calibrated: Low 0.62-0.67, Mid 1.12-1.15, High 0.55-0.65
		{
			Role: tuning.RoleActivator,
			SymbolDensity: map[string]float64{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     1,
				"wutong":    1,
				"liangsuo":  1,
				"liangtong": 1,
				"bonus":     2.64,
			},
			MinSpacing: map[string]int{
				"fa":    4,
				"zhong": 4,
				"bai":   4,
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
		},

		// Reel 2: Connector - continues patterns
		{
			Role: tuning.RoleConnector,
			SymbolDensity: map[string]float64{
				"fa":        2.55,
				"zhong":     3.08,
				"bai":       3.51,
				"bawan":     3.9,
				"wusuo":     1.0,
				"wutong":    1.0,
				"liangsuo":  1.38,
				"liangtong": 1.36,
				"bonus":     2.58,
			},
			MinSpacing: map[string]int{
				"bonus": 10,
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			GoldConfig: &tuning.GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.05, // 10% of paying symbols will be gold
				MinGoldSpacing: 2,    // At least 5 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 3: Core - main win driver
		{
			Role: tuning.RoleCore,
			SymbolDensity: map[string]float64{
				"fa":        1.0,
				"zhong":     1.18,
				"bai":       1.46,
				"bawan":     1.64,
				"wusuo":     5.09,
				"wutong":    5.15,
				"liangsuo":  7.06,
				"liangtong": 7.11,
				"bonus":     2.6,
			},
			MinSpacing: map[string]int{
				"fa":    4,
				"zhong": 4,
				"bai":   4,
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			GoldConfig: &tuning.GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.06, // 12% of paying symbols will be gold (core reel has slightly more)
				MinGoldSpacing: 2,    // At least 4 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 4: Amplifier - extends wins to 4-of-kind
		{
			Role: tuning.RoleAmplifier,
			SymbolDensity: map[string]float64{
				"fa":        2.69,
				"zhong":     3.08,
				"bai":       3.52,
				"bawan":     3.91,
				"wusuo":     1.0,
				"wutong":    1.0,
				"liangsuo":  1.38,
				"liangtong": 1.37,
				"bonus":     2.24,
			},
			MinSpacing: map[string]int{
				"bonus": 10,
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			GoldConfig: &tuning.GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.07, // 8% of paying symbols will be gold (amplifier reel has less)
				MinGoldSpacing: 2,    // At least 6 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 5: Rare Spike - controls 5-of-kind (rare big wins)
		{
			Role: tuning.RoleRareSpike,
			SymbolDensity: map[string]float64{
				"fa":        1.0,
				"zhong":     1.18,
				"bai":       1.46,
				"bawan":     1.64,
				"wusuo":     5.09,
				"wutong":    5.15,
				"liangsuo":  7.06,
				"liangtong": 7.11,
				"bonus":     2.0,
			},
			MinSpacing: map[string]int{
				"fa":    4,
				"zhong": 4,
				"bai":   4,
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
		},
	}
}
