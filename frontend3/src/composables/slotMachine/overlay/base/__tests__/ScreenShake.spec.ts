import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { ScreenShake, createScreenShake } from '../ScreenShake'

describe('ScreenShake', () => {
  let container: Container
  let screenShake: ScreenShake

  beforeEach(() => {
    vi.useFakeTimers()
    container = new Container()
    container.x = 0
    container.y = 0
    screenShake = createScreenShake(container)
  })

  afterEach(() => {
    screenShake.destroy()
    vi.useRealTimers()
  })

  describe('start', () => {
    it('should set shake to active', () => {
      expect(screenShake.isActive()).toBe(false)
      screenShake.start()
      expect(screenShake.isActive()).toBe(true)
    })

    it('should auto-stop after duration', () => {
      screenShake.start(15, 500)
      expect(screenShake.isActive()).toBe(true)

      vi.advanceTimersByTime(500)
      expect(screenShake.isActive()).toBe(false)
    })
  })

  describe('stop', () => {
    it('should reset container position', () => {
      container.x = 100
      container.y = 100
      screenShake = createScreenShake(container)

      screenShake.start(20, 1000)
      // Simulate some update calls that would offset the container
      screenShake.update()

      screenShake.stop()
      expect(container.x).toBe(100)
      expect(container.y).toBe(100)
    })

    it('should set active to false', () => {
      screenShake.start()
      expect(screenShake.isActive()).toBe(true)

      screenShake.stop()
      expect(screenShake.isActive()).toBe(false)
    })
  })

  describe('update', () => {
    it('should not change position when not active', () => {
      const originalX = container.x
      const originalY = container.y

      screenShake.update()

      expect(container.x).toBe(originalX)
      expect(container.y).toBe(originalY)
    })

    it('should offset container position when active', () => {
      // Set Math.random to return predictable values
      vi.spyOn(Math, 'random').mockReturnValue(0.75)

      screenShake.start(10, 1000, 1) // decay of 1 means no decay
      screenShake.update()

      // With random = 0.75, offset = (0.75 - 0.5) * 10 * 2 = 5
      expect(container.x).toBe(5)
      expect(container.y).toBe(5)
    })
  })

  describe('isActive', () => {
    it('should return correct state', () => {
      expect(screenShake.isActive()).toBe(false)
      screenShake.start()
      expect(screenShake.isActive()).toBe(true)
      screenShake.stop()
      expect(screenShake.isActive()).toBe(false)
    })
  })

  describe('destroy', () => {
    it('should stop shake and clear timeout', () => {
      screenShake.start(15, 1000)
      screenShake.destroy()

      expect(screenShake.isActive()).toBe(false)
    })
  })
})
