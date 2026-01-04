// @ts-nocheck
/**
 * Audio Effects Composable - Event-Driven Architecture
 * Helper composable that emits audio events for sound effects
 */

import { audioEvents, AUDIO_EVENTS } from './audioEventBus'
import type { WinCombination } from '@/features/spin/types'

/**
 * Audio effects composable interface
 */
export interface UseAudioEffects {
  playWinSound: (wins: WinCombination[]) => void
  playConsecutiveWinSound: (consecutiveWins: number, isFreeSpin?: boolean) => void
  playWinningAnnouncement: () => void
  stopWinningAnnouncement: () => void
  playEffect: (effect: string) => void
  stopEffect: (effect: string) => void
  fadeEffect: (effect: string, from: number, to: number, duration: number) => void
  playWinningHighlight: () => void
  playLineWinSound: () => void
}

/**
 * Audio effects composable - Event-driven implementation
 * All methods emit events to the audio event bus
 */
export function useAudioEffects(): UseAudioEffects {

  /**
   * Play winning announcement sound (looped) for win overlay
   */
  const playWinningAnnouncement = (): void => {
    audioEvents.emit(AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT)
  }

  /**
   * Stop winning announcement sound
   */
  const stopWinningAnnouncement = (): void => {
    audioEvents.emit(AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT_STOP)
  }

  /**
   * Play win sound for specific symbol combinations
   */
  const playWinSound = (wins: WinCombination[]): void => {
    if (!wins || wins.length === 0) return

    // Emit win event with wins data - audio player will handle the logic
    audioEvents.emit(AUDIO_EVENTS.EFFECT_WIN, { wins })
  }

  /**
   * Play consecutive wins sound based on multiplier
   */
  const playConsecutiveWinSound = (consecutiveWins: number, isFreeSpin: boolean = false): void => {
    // Emit consecutive win event - audio player will handle the logic
    audioEvents.emit(AUDIO_EVENTS.EFFECT_CONSECUTIVE_WIN, { 
      consecutiveWins, 
      isFreeSpin 
    })
  }

  /**
   * Play a generic sound effect
   */
  const playEffect = (effect: string): void => {
    // Emit generic effect event
    audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { 
      audioKey: effect 
    })
  }

  /**
   * Stop a generic sound effect
   */
  const stopEffect = (effect: string): void => {
    // Emit stop effect event
    audioEvents.emit(AUDIO_EVENTS.EFFECT_STOP, {
      audioKey: effect
    })
  }

  /**
   * Fade a sound effect volume
   * @param effect - Effect name
   * @param from - Starting volume (0.0 to 1.0)
   * @param to - Target volume (0.0 to 1.0)
   * @param duration - Duration in milliseconds
   */
  const fadeEffect = (effect: string, from: number, to: number, duration: number): void => {
    audioEvents.emit(AUDIO_EVENTS.EFFECT_FADE, {
      audioKey: effect,
      from,
      to,
      duration
    })
  }

  /**
   * Play winning highlight sound when winning frames appear
   */
  const playWinningHighlight = (): void => {
    audioEvents.emit(AUDIO_EVENTS.EFFECT_WINNING_HIGHLIGHT)
  }

  /**
   * Play line win sound when winning line is shown
   * Should be relatively quiet
   */
  const playLineWinSound = (): void => {
    audioEvents.emit(AUDIO_EVENTS.EFFECT_LINE_WIN)
  }

  return {
    playWinSound,
    playConsecutiveWinSound,
    playWinningAnnouncement,
    stopWinningAnnouncement,
    playEffect,
    stopEffect,
    fadeEffect,
    playWinningHighlight,
    playLineWinSound
  }
}
