package multiplier

// GetMultiplier returns the multiplier for a given cascade number
// Base game: 1x, 2x, 3x, 5x... (increments by 1)
// Free spins: 2x, 4x, 6x, 10x... (increments by 2, starts at 2x)
func GetMultiplier(cascadeNumber int, isFreeSpin bool) int {
	if cascadeNumber < 1 {
		cascadeNumber = 1
	}

	if isFreeSpin {
		switch cascadeNumber {
		case 1:
			return 2
		case 2:
			return 4
		case 3:
			return 6
		default:
			return 10 // 4+
		}
	}

	switch cascadeNumber {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	default:
		return 5 // 4+
	}
}

// GetBaseMultiplier returns the starting multiplier for a mode
func GetBaseMultiplier(isFreeSpin bool) int {
	if isFreeSpin {
		return 2 // Free spins start at 2x
	}
	return 1 // Base game starts at 1x
}

// GetIncrement returns the multiplier increment per cascade
func GetIncrement(isFreeSpin bool) int {
	if isFreeSpin {
		return 2 // +2 per cascade in free spins
	}
	return 1 // +1 per cascade in base game
}

// CalculateMultiplierProgression returns the full multiplier progression for N cascades
func CalculateMultiplierProgression(numCascades int, isFreeSpin bool) []int {
	progression := make([]int, numCascades)
	for i := 0; i < numCascades; i++ {
		progression[i] = GetMultiplier(i+1, isFreeSpin)
	}
	return progression
}
