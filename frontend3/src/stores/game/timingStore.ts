/**
 * Timing Store
 * Central configuration for all game timing constants
 * Single source of truth for all durations, delays, and waits
 */

import { defineStore } from 'pinia'

/**
 * Timing store state interface
 */
export interface TimingState {
  // ========== WIN CYCLE TIMING ==========
  // The complete win cycle: Highlight → Flip → Disappear → Cascade → Wait

  /** Show highlight effect before flip starts (reduced for snappier flow) */
  HIGHLIGHT_BEFORE_FLIP: number

  /** Tiles flip from visible to hidden */
  FLIP_DURATION: number

  /** Brief pause after flip for better win feedback */
  DISAPPEAR_WAIT: number

  /** Brief pause to show wild transformation (reduced to prevent double-drop feeling) */
  GOLD_WAIT: number

  /** Duration of gold-to-wild transformation animation */
  GOLD_TRANSFORM_DURATION: number

  /** Wait after cascade before checking for next win (reduced for snappier gameplay) */
  CASCADE_WAIT: number

  // ========== ANIMATION DURATIONS ==========

  /** How long highlight animation runs (can be stopped early) */
  HIGHLIGHT_ANIMATION_DURATION: number

  /** Tiles falling down during cascade with natural gravity physics */
  DROP_DURATION: number

  /** Keep dropped symbols stable to prevent flickering */
  DROP_GRACE_PERIOD: number

  /** Window to force sprite resets after cascade */
  CASCADE_RESET_WINDOW: number

  /** Max time to wait for cascade animations (safety timeout) */
  CASCADE_MAX_WAIT: number

  /** Bonus tile "bump" animation (up and down) */
  BUMP_DURATION: number

  // ========== SPIN TIMING ==========

  /** Base duration for first reel (normal speed) */
  SPIN_BASE_DURATION_NORMAL: number

  /** Base duration for first reel (fast speed) */
  SPIN_BASE_DURATION_FAST: number

  /** Delay between each reel (normal speed) */
  SPIN_REEL_STAGGER_NORMAL: number

  /** Delay between each reel (fast speed) */
  SPIN_REEL_STAGGER_FAST: number

  // ========== ANTICIPATION MODE TIMING ==========

  /** Exact time each column takes to slow down and stop during anticipation (normal speed) */
  ANTICIPATION_SLOWDOWN_PER_COLUMN_NORMAL: number

  /** Exact time each column takes to slow down and stop during anticipation (fast speed) */
  ANTICIPATION_SLOWDOWN_PER_COLUMN_FAST: number

  // ========== JACKPOT TIMING ==========

  /** Pause after last column stops to let player observe result before bonus tiles start popping */
  JACKPOT_PAUSE_BEFORE_POP: number
}

/**
 * Timing store - manages all game timing constants
 */
export const useTimingStore = defineStore('timing', {
  state: (): TimingState => ({
    // ========== WIN CYCLE TIMING ==========
    // The complete win cycle: Highlight → Flip → Disappear → Cascade → Wait

    // Phase 1: Highlight winning tiles
    HIGHLIGHT_BEFORE_FLIP: 450,  // ms - Show highlight effect before flip starts (reduced for snappier flow)

    // Phase 2: Flip animation
    FLIP_DURATION: 600,           // ms - Tiles flip from visible to hidden (slower for visibility)

    // Phase 3: Disappear wait
    DISAPPEAR_WAIT: 0,            // ms - Brief pause after flip for better win feedback

    // Phase 4: Gold transformation (if applicable)
    GOLD_WAIT: 0,                 // ms - Brief pause to show wild transformation (reduced to prevent double-drop feeling)
    GOLD_TRANSFORM_DURATION: 600, // ms - Duration of gold-to-wild transformation animation (shimmer + burst + transform)

    // Phase 5: Cascade wait
    CASCADE_WAIT: 700,            // ms - Wait after cascade before checking for next win (reduced for snappier gameplay)

    // ========== ANIMATION DURATIONS ==========

    // Highlight effect (glow/pulse animation)
    HIGHLIGHT_ANIMATION_DURATION: 800,   // ms - How long highlight animation runs (can be stopped early) - Extreme speed

    // Drop/cascade animations
    DROP_DURATION: 300,           // ms - Tiles falling down during cascade with natural gravity physics
    DROP_GRACE_PERIOD: 6000,      // ms - Keep dropped symbols stable to prevent flickering
    CASCADE_RESET_WINDOW: 100,    // ms - Window to force sprite resets after cascade - Extreme speed
    CASCADE_MAX_WAIT: 5000,       // ms - Max time to wait for cascade animations (safety timeout)

    // Tile animations
    BUMP_DURATION: 250,           // ms - Bonus tile "bump" animation (up and down) - Extreme speed

    // ========== SPIN TIMING ==========
    SPIN_BASE_DURATION_NORMAL: 1500,     // ms - Base duration for first reel (normal speed)
    SPIN_BASE_DURATION_FAST: 250,        // ms - Base duration for first reel (fast speed)
    SPIN_REEL_STAGGER_NORMAL: 200,       // ms - Delay between each reel (normal speed)
    SPIN_REEL_STAGGER_FAST: 30,          // ms - Delay between each reel (fast speed)

    // ========== ANTICIPATION MODE TIMING ==========
    ANTICIPATION_SLOWDOWN_PER_COLUMN_NORMAL: 4000,  // ms - Anticipation timing (normal speed)
    ANTICIPATION_SLOWDOWN_PER_COLUMN_FAST: 800,     // ms - Anticipation timing (fast speed)

    // ========== JACKPOT TIMING (always uses normal speed) ==========
    JACKPOT_PAUSE_BEFORE_POP: 1000  // ms - Pause after last column stops to let player observe result before bonus tiles start popping
  }),

  getters: {
    /**
     * Total time from highlight start to flip complete
     * Used by sparkle effects and other animations that need to sync with flip
     */
    totalHighlightToFlipComplete(): number {
      return this.HIGHLIGHT_BEFORE_FLIP + this.FLIP_DURATION
    },

    /**
     * Total time for one complete win cycle (without cascade animation time)
     * Highlight → Flip → Disappear → Cascade → Wait
     */
    totalWinCycleDuration(): number {
      return this.HIGHLIGHT_BEFORE_FLIP +
             this.FLIP_DURATION +
             this.DISAPPEAR_WAIT +
             this.GOLD_WAIT +
             this.DROP_DURATION +  // Assuming average drop time
             this.CASCADE_WAIT
    },

    /**
     * When flip animation starts (relative to highlight start)
     */
    flipStartTime(): number {
      return this.HIGHLIGHT_BEFORE_FLIP
    },

    /**
     * When flip animation ends (relative to highlight start)
     */
    flipEndTime(): number {
      return this.HIGHLIGHT_BEFORE_FLIP + this.FLIP_DURATION
    }
  },

  actions: {

    /**
     * Update a timing value at runtime (useful for debugging/tuning)
     * @param key - The timing constant to update
     * @param value - New value in milliseconds
     */
    updateTiming(key: keyof TimingState, value: number): void {
      if (key in this.$state) {
        this.$state[key] = value
      }
    },

    /**
     * Reset all timings to defaults (useful for testing)
     */
    resetToDefaults(): void {
      this.$reset()
    }
  }
})
