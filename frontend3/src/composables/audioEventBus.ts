/**
 * Audio Event Bus - Event-Driven Architecture
 * Central communication layer for all audio operations
 */

/**
 * Audio Event Types
 * All available audio events in the system
 */
export const AUDIO_EVENTS = {
  // Background Music Events
  MUSIC_START: 'music:start',
  MUSIC_STOP: 'music:stop', 
  MUSIC_PAUSE: 'music:pause',
  MUSIC_RESUME: 'music:resume',
  MUSIC_SWITCH_JACKPOT: 'music:switch:jackpot',
  MUSIC_SWITCH_NORMAL: 'music:switch:normal',
  MUSIC_SET_VOLUME: 'music:set:volume',

  // Sound Effect Events
  EFFECT_PLAY: 'effect:play',
  EFFECT_STOP: 'effect:stop',
  EFFECT_FADE: 'effect:fade',
  EFFECT_WIN: 'effect:win',
  EFFECT_CONSECUTIVE_WIN: 'effect:consecutive:win',
  EFFECT_WINNING_ANNOUNCEMENT: 'effect:winning:announcement',
  EFFECT_WINNING_ANNOUNCEMENT_STOP: 'effect:winning:announcement:stop',
  EFFECT_WINNING_HIGHLIGHT: 'effect:winning:highlight',
  EFFECT_LINE_WIN: 'effect:line:win',

  // Global Control Events
  AUDIO_ENABLE: 'audio:enable',
  AUDIO_DISABLE: 'audio:disable',
  AUDIO_SET_GLOBAL_VOLUME: 'audio:set:global:volume'
} as const

/**
 * Type definitions for audio events
 */
export type AudioEventType = typeof AUDIO_EVENTS[keyof typeof AUDIO_EVENTS]

/**
 * Event data interfaces
 */
export interface AudioEventData {
  [AUDIO_EVENTS.MUSIC_SET_VOLUME]: { volume: number }
  [AUDIO_EVENTS.EFFECT_PLAY]: { audioKey: string; volume?: number }
  [AUDIO_EVENTS.EFFECT_STOP]: { audioKey: string }
  [AUDIO_EVENTS.EFFECT_FADE]: { audioKey: string; from: number; to: number; duration: number }
  [AUDIO_EVENTS.EFFECT_WIN]: { wins: Array<{ symbol: string | number; count: number }> }
  [AUDIO_EVENTS.EFFECT_CONSECUTIVE_WIN]: { consecutiveWins: number; isFreeSpin?: boolean }
  [AUDIO_EVENTS.AUDIO_SET_GLOBAL_VOLUME]: { volume: number }
  // Events without data
  [AUDIO_EVENTS.MUSIC_START]: null
  [AUDIO_EVENTS.MUSIC_STOP]: null
  [AUDIO_EVENTS.MUSIC_PAUSE]: null
  [AUDIO_EVENTS.MUSIC_RESUME]: null
  [AUDIO_EVENTS.MUSIC_SWITCH_JACKPOT]: null
  [AUDIO_EVENTS.MUSIC_SWITCH_NORMAL]: null
  [AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT]: null
  [AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT_STOP]: null
  [AUDIO_EVENTS.EFFECT_WINNING_HIGHLIGHT]: null
  [AUDIO_EVENTS.EFFECT_LINE_WIN]: null
  [AUDIO_EVENTS.AUDIO_ENABLE]: null
  [AUDIO_EVENTS.AUDIO_DISABLE]: null
}

/**
 * Event handler type
 */
export type AudioEventHandler<T extends AudioEventType> = (data: AudioEventData[T]) => void

/**
 * Simple Event Emitter Implementation
 * Provides event emission and listening capabilities
 */
class EventBus {
  private events = new Map<string, Array<(data: any) => void>>()

  /**
   * Add event listener
   * @param eventType - Event type to listen for
   * @param handler - Handler function
   */
  on<T extends AudioEventType>(eventType: T, handler: AudioEventHandler<T>): void {
    if (!this.events.has(eventType)) {
      this.events.set(eventType, [])
    }
    this.events.get(eventType)!.push(handler)
  }

  /**
   * Remove event listener
   * @param eventType - Event type
   * @param handler - Handler function to remove
   */
  off<T extends AudioEventType>(eventType: T, handler: AudioEventHandler<T>): void {
    if (!this.events.has(eventType)) return
    
    const handlers = this.events.get(eventType)!
    const index = handlers.indexOf(handler)
    if (index > -1) {
      handlers.splice(index, 1)
    }
  }

  /**
   * Emit event
   * @param eventType - Event type to emit
   * @param data - Data to pass to handlers
   */
  emit<T extends AudioEventType>(eventType: T, data: AudioEventData[T] = null as AudioEventData[T]): void {
    if (!this.events.has(eventType)) return
    
    const handlers = this.events.get(eventType)!
    handlers.forEach(handler => {
      try {
        handler(data)
      } catch (error) {
        console.error(`Error in audio event handler for ${eventType}:`, error)
      }
    })
  }

  /**
   * Remove all listeners for an event type
   * @param eventType - Event type
   */
  removeAllListeners(eventType: AudioEventType): void {
    this.events.delete(eventType)
  }

  /**
   * Remove all listeners for all events
   */
  removeAllEventListeners(): void {
    this.events.clear()
  }
}

// Export singleton instance
export const audioEvents = new EventBus()
