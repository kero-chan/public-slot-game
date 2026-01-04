package tuning

// SymbolWeights represents weight configuration for a single reel
type SymbolWeights map[string]int

// ReelWeightsSet represents weights for all 5 reels
type ReelWeightsSet struct {
	Reel1 SymbolWeights `json:"reel1"`
	Reel2 SymbolWeights `json:"reel2"`
	Reel3 SymbolWeights `json:"reel3"`
	Reel4 SymbolWeights `json:"reel4"`
	Reel5 SymbolWeights `json:"reel5"`
}

// GetReel returns weights for a specific reel (0-indexed)
func (r *ReelWeightsSet) GetReel(index int) SymbolWeights {
	switch index {
	case 0:
		return r.Reel1
	case 1:
		return r.Reel2
	case 2:
		return r.Reel3
	case 3:
		return r.Reel4
	case 4:
		return r.Reel5
	default:
		return nil
	}
}

// SetReel sets weights for a specific reel (0-indexed)
func (r *ReelWeightsSet) SetReel(index int, weights SymbolWeights) {
	switch index {
	case 0:
		r.Reel1 = weights
	case 1:
		r.Reel2 = weights
	case 2:
		r.Reel3 = weights
	case 3:
		r.Reel4 = weights
	case 4:
		r.Reel5 = weights
	}
}

// Clone creates a deep copy of ReelWeightsSet
func (r *ReelWeightsSet) Clone() *ReelWeightsSet {
	clone := &ReelWeightsSet{
		Reel1: make(SymbolWeights),
		Reel2: make(SymbolWeights),
		Reel3: make(SymbolWeights),
		Reel4: make(SymbolWeights),
		Reel5: make(SymbolWeights),
	}
	for k, v := range r.Reel1 {
		clone.Reel1[k] = v
	}
	for k, v := range r.Reel2 {
		clone.Reel2[k] = v
	}
	for k, v := range r.Reel3 {
		clone.Reel3[k] = v
	}
	for k, v := range r.Reel4 {
		clone.Reel4[k] = v
	}
	for k, v := range r.Reel5 {
		clone.Reel5[k] = v
	}
	return clone
}

// BlockPattern defines a symbol block pattern for strip generation
type BlockPattern struct {
	Symbols []string `json:"symbols"` // e.g., ["low", "low", "high"]
	Weight  int      `json:"weight"`  // Weight for selection
}

// StandardBlockPatterns returns PG-style block patterns
func StandardBlockPatterns() []BlockPattern {
	return []BlockPattern{
		// Low symbol clusters
		{Symbols: []string{"low", "low", "low"}, Weight: 30},
		{Symbols: []string{"low", "low", "high"}, Weight: 20},
		{Symbols: []string{"low", "high", "low"}, Weight: 15},

		// High symbol patterns (less frequent)
		{Symbols: []string{"high", "low", "low"}, Weight: 15},
		{Symbols: []string{"high", "high", "low"}, Weight: 8},
		{Symbols: []string{"low", "high", "high"}, Weight: 5},

		// Special patterns
		{Symbols: []string{"low", "scatter", "low"}, Weight: 3},
		{Symbols: []string{"low", "low", "wild"}, Weight: 2},
		{Symbols: []string{"wild", "low", "low"}, Weight: 2},
	}
}
