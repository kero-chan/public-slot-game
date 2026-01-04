/**
 * Timer Manager - Centralized timer management with session-based invalidation
 *
 * Prevents stale callbacks from executing after overlay is hidden or re-shown
 * by using session IDs to track validity of scheduled callbacks.
 */
export class TimerManager {
  private sessionId = 0
  private timeouts: ReturnType<typeof setTimeout>[] = []
  private intervals: ReturnType<typeof setInterval>[] = []

  /**
   * Start a new session - invalidates all previous timers
   * Call this at the start of show()
   */
  newSession(): number {
    this.clearAll()
    this.sessionId++
    return this.sessionId
  }

  /**
   * Get current session ID
   */
  getSessionId(): number {
    return this.sessionId
  }

  /**
   * Schedule a timeout that only executes if session is still valid
   */
  setTimeout(callback: () => void, delay: number, sessionId?: number): ReturnType<typeof setTimeout> {
    const targetSession = sessionId ?? this.sessionId
    const timeout = setTimeout(() => {
      if (targetSession === this.sessionId) {
        callback()
      }
      this.removeTimeout(timeout)
    }, delay)
    this.timeouts.push(timeout)
    return timeout
  }

  /**
   * Schedule an interval that only executes if session is still valid
   */
  setInterval(callback: () => void, delay: number, sessionId?: number): ReturnType<typeof setInterval> {
    const targetSession = sessionId ?? this.sessionId
    const interval = setInterval(() => {
      if (targetSession === this.sessionId) {
        callback()
      } else {
        this.clearInterval(interval)
      }
    }, delay)
    this.intervals.push(interval)
    return interval
  }

  /**
   * Clear a specific timeout
   */
  clearTimeout(timeout: ReturnType<typeof setTimeout>): void {
    clearTimeout(timeout)
    this.removeTimeout(timeout)
  }

  /**
   * Clear a specific interval
   */
  clearInterval(interval: ReturnType<typeof setInterval>): void {
    clearInterval(interval)
    this.removeInterval(interval)
  }

  /**
   * Clear all timers
   */
  clearAll(): void {
    this.timeouts.forEach(t => clearTimeout(t))
    this.timeouts = []
    this.intervals.forEach(i => clearInterval(i))
    this.intervals = []
  }

  private removeTimeout(timeout: ReturnType<typeof setTimeout>): void {
    const index = this.timeouts.indexOf(timeout)
    if (index > -1) {
      this.timeouts.splice(index, 1)
    }
  }

  private removeInterval(interval: ReturnType<typeof setInterval>): void {
    const index = this.intervals.indexOf(interval)
    if (index > -1) {
      this.intervals.splice(index, 1)
    }
  }
}

/**
 * Create a timer manager instance
 */
export function createTimerManager(): TimerManager {
  return new TimerManager()
}
