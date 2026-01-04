import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { TimerManager, createTimerManager } from '../TimerManager'

describe('TimerManager', () => {
  let timerManager: TimerManager

  beforeEach(() => {
    vi.useFakeTimers()
    timerManager = createTimerManager()
  })

  afterEach(() => {
    timerManager.clearAll()
    vi.useRealTimers()
  })

  describe('newSession', () => {
    it('should increment session ID', () => {
      const session1 = timerManager.newSession()
      const session2 = timerManager.newSession()

      expect(session2).toBe(session1 + 1)
    })

    it('should clear all existing timers when starting new session', () => {
      const callback = vi.fn()
      timerManager.setTimeout(callback, 1000)

      timerManager.newSession()
      vi.advanceTimersByTime(1000)

      expect(callback).not.toHaveBeenCalled()
    })
  })

  describe('setTimeout', () => {
    it('should execute callback after delay', () => {
      const callback = vi.fn()
      timerManager.setTimeout(callback, 1000)

      vi.advanceTimersByTime(999)
      expect(callback).not.toHaveBeenCalled()

      vi.advanceTimersByTime(1)
      expect(callback).toHaveBeenCalledOnce()
    })

    it('should not execute callback if session changed', () => {
      const callback = vi.fn()
      const sessionId = timerManager.getSessionId()
      timerManager.setTimeout(callback, 1000, sessionId)

      // Start a new session before timeout fires
      timerManager.newSession()
      vi.advanceTimersByTime(1000)

      expect(callback).not.toHaveBeenCalled()
    })

    it('should execute callback for current session', () => {
      const callback = vi.fn()
      timerManager.newSession()
      const sessionId = timerManager.getSessionId()
      timerManager.setTimeout(callback, 1000, sessionId)

      vi.advanceTimersByTime(1000)

      expect(callback).toHaveBeenCalledOnce()
    })
  })

  describe('setInterval', () => {
    it('should execute callback repeatedly', () => {
      const callback = vi.fn()
      timerManager.setInterval(callback, 1000)

      vi.advanceTimersByTime(3000)

      expect(callback).toHaveBeenCalledTimes(3)
    })

    it('should stop executing if session changed', () => {
      const callback = vi.fn()
      const sessionId = timerManager.getSessionId()
      timerManager.setInterval(callback, 1000, sessionId)

      vi.advanceTimersByTime(2000) // Called twice
      timerManager.newSession()
      vi.advanceTimersByTime(3000) // Should not be called

      expect(callback).toHaveBeenCalledTimes(2)
    })
  })

  describe('clearTimeout', () => {
    it('should prevent callback from executing', () => {
      const callback = vi.fn()
      const timeout = timerManager.setTimeout(callback, 1000)

      timerManager.clearTimeout(timeout)
      vi.advanceTimersByTime(1000)

      expect(callback).not.toHaveBeenCalled()
    })
  })

  describe('clearInterval', () => {
    it('should stop interval from executing', () => {
      const callback = vi.fn()
      const interval = timerManager.setInterval(callback, 1000)

      vi.advanceTimersByTime(2000) // Called twice
      timerManager.clearInterval(interval)
      vi.advanceTimersByTime(3000) // Should not be called

      expect(callback).toHaveBeenCalledTimes(2)
    })
  })

  describe('clearAll', () => {
    it('should clear all timeouts and intervals', () => {
      const timeout1 = vi.fn()
      const timeout2 = vi.fn()
      const interval1 = vi.fn()

      timerManager.setTimeout(timeout1, 1000)
      timerManager.setTimeout(timeout2, 2000)
      timerManager.setInterval(interval1, 500)

      timerManager.clearAll()
      vi.advanceTimersByTime(5000)

      expect(timeout1).not.toHaveBeenCalled()
      expect(timeout2).not.toHaveBeenCalled()
      expect(interval1).not.toHaveBeenCalled()
    })
  })

  describe('getSessionId', () => {
    it('should return current session ID', () => {
      const session1 = timerManager.newSession()
      expect(timerManager.getSessionId()).toBe(session1)

      const session2 = timerManager.newSession()
      expect(timerManager.getSessionId()).toBe(session2)
    })
  })
})
