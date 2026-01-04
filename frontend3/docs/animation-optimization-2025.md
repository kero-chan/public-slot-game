# Animation Optimization (Nov 2025)

## Overview
Optimized timing parameters after migrating to 10-row backend grid (4 buffer + 6 visible rows).

## Changes Applied

### 1. Drop Animation (CASCADE)
**Before:**
```javascript
DROP_DURATION: 450ms
ease: 'power1.out'  // Linear deceleration
```

**After:**
```javascript
DROP_DURATION: 650ms  // +44% longer
ease: 'back.out(1.15)'  // Subtle bounce on landing
```

**Impact:**
- Smoother drops for 10-row grid (symbols travel further from buffer rows)
- Subtle bounce creates satisfying "thud" feel when tiles land
- More natural cascade animation

---

### 2. Cascade Wait Time
**Before:**
```javascript
CASCADE_WAIT: 1000ms
```

**After:**
```javascript
CASCADE_WAIT: 700ms  // -30% faster
```

**Impact:**
- Snappier gameplay flow between cascades
- Reduces wait time after symbols settle
- Players see next win/cascade sooner

---

### 3. Win Highlight Duration
**Before:**
```javascript
HIGHLIGHT_BEFORE_FLIP: 600ms
```

**After:**
```javascript
HIGHLIGHT_BEFORE_FLIP: 450ms  // -25% faster
```

**Impact:**
- Faster win-to-cascade transition
- Less waiting before cascade starts
- Still enough time to see winning tiles

---

### 4. Disappear Wait
**Before:**
```javascript
DISAPPEAR_WAIT: 50ms
```

**After:**
```javascript
DISAPPEAR_WAIT: 120ms  // +140% longer
```

**Impact:**
- Better win feedback - tiles don't vanish instantly
- Players have time to register the win
- More satisfying win animation

---

## Total Impact

### Win Cycle Timeline Comparison

**Before (Total: ~2.4 seconds):**
```
Highlight: 600ms
Flip: 300ms
Disappear: 50ms
Drop: 450ms
Wait: 1000ms
────────────────
Total: 2400ms
```

**After (Total: ~2.22 seconds - 7.5% faster):**
```
Highlight: 450ms  (-150ms)
Flip: 300ms
Disappear: 120ms  (+70ms)
Drop: 650ms  (+200ms)
Wait: 700ms  (-300ms)
────────────────
Total: 2220ms (-180ms)
```

### Overall Game Feel
- **~8% faster gameplay** while feeling smoother
- **Better feedback** on wins (longer disappear wait)
- **More natural cascades** (bounce easing + longer duration)
- **Snappier flow** (reduced highlight and cascade waits)

## Technical Notes

### Bounce Easing Calculation
```javascript
ease: 'back.out(1.15)'
```
- Overshoots target by ~15% then settles back
- Creates subtle "impact" feel without being cartoonish
- Optimized for tile-based grid (minimal overshoot to avoid visual overlap)

### Drop Duration Rationale
With 10-row grid:
- Buffer rows (0-3) to visible rows (4-9) = ~4-6 tile heights
- Previous 450ms felt rushed for this distance
- 650ms provides ~100ms per tile height (natural falling speed)
- Bounce easing adds ~50ms perceived duration

### Cascade Wait Optimization
- Previous 1000ms included safety margin for animation cleanup
- GSAP's tween system is reliable, reduced safety margin
- 700ms still provides comfortable buffer for:
  - Drop completion
  - Symbol texture swaps
  - Grid state synchronization

## Files Modified

1. `frontend/src/stores/timingStore.js`
   - Lines 14, 20, 26, 34

2. `frontend/src/composables/slotMachine/reels/dropAnimation.js`
   - Lines 48-49

## Testing Recommendations

1. Test with multiple consecutive cascades (5+ wins in a row)
2. Verify bounce doesn't cause visual overlap between tiles
3. Check timing feels consistent across different device framerates
4. Test in free spin mode (2x multipliers) for cascade speed
5. Verify anticipation mode slowdown isn't affected

## Rollback Instructions

If animation feels too fast/bouncy, revert to previous values:
```javascript
// timingStore.js
HIGHLIGHT_BEFORE_FLIP: 600
DISAPPEAR_WAIT: 50
CASCADE_WAIT: 1000
DROP_DURATION: 450

// dropAnimation.js
ease: 'power1.out'
```
