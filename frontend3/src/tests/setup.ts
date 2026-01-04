import { config } from '@vue/test-utils'
import { vi } from 'vitest'

// Mock Howler globally
vi.mock('howler', () => ({
  Howl: vi.fn(() => ({
    play: vi.fn(),
    stop: vi.fn(),
    pause: vi.fn(),
    volume: vi.fn(),
    fade: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    once: vi.fn(),
  })),
  Howler: {
    volume: vi.fn(),
    mute: vi.fn(),
  }
}))

// Mock GSAP
vi.mock('gsap', () => ({
  gsap: {
    to: vi.fn((target, vars) => ({
      kill: vi.fn(),
      play: vi.fn(),
      pause: vi.fn(),
      isActive: vi.fn(() => false),
    })),
    from: vi.fn((target, vars) => ({
      kill: vi.fn(),
      play: vi.fn(),
    })),
    timeline: vi.fn(() => ({
      add: vi.fn(),
      play: vi.fn(),
      pause: vi.fn(),
      kill: vi.fn(),
      isActive: vi.fn(() => false),
    })),
    killTweensOf: vi.fn(),
  },
  default: {
    to: vi.fn(),
    from: vi.fn(),
    timeline: vi.fn(),
  }
}))

// Mock PixiJS (basic mocks)
vi.mock('pixi.js', async () => {
  const actual = await vi.importActual('pixi.js')
  return {
    ...actual,
    Application: vi.fn(() => ({
      stage: {
        addChild: vi.fn(),
        removeChild: vi.fn(),
      },
      renderer: {
        render: vi.fn(),
        resize: vi.fn(),
      },
      ticker: {
        add: vi.fn(),
        remove: vi.fn(),
      },
      destroy: vi.fn(),
    })),
  }
})

// Configure Vue Test Utils
config.global.mocks = {
  $t: (key: string) => key,
}

// Suppress console warnings in tests
global.console = {
  ...console,
  warn: vi.fn(),
  error: vi.fn(),
}
