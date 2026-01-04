// @ts-nocheck
import { Container, Graphics, Sprite } from 'pixi.js'
import gsap from 'gsap'
import { SPIN_BUTTON_CIRCLE_RADIUS_PER_MENU_HEIGHT, type ScalableSprite, type FooterHandlers } from './types'
import type { ParticleState } from './particleEffects'
import { getIconSprite } from '@/config/spritesheet'

export interface SpinButtonState {
  spinBtnSprite: ScalableSprite | null
  spinHoverCircle: Graphics | null
  spinTween: gsap.core.Tween | null
  isSpinning: boolean
}

export function createSpinButtonState(): SpinButtonState {
  return {
    spinBtnSprite: null,
    spinHoverCircle: null,
    spinTween: null,
    isSpinning: false
  }
}

export function buildSpinButton(
  container: Container,
  state: SpinButtonState,
  particleState: ParticleState,
  menuHeight: number,
  xPosition: number,
  yPosition: number,
  handlers: FooterHandlers,
  gameState: any,
  onStartSpin: () => void
): void {
  state.spinBtnSprite = getIconSprite('icon_spin.webp')
  if (!state.spinBtnSprite) return

  state.spinBtnSprite.anchor.set(0.5)
  state.spinBtnSprite.position.set(xPosition, yPosition)

  // Size is derived from the circle radius constant (diameter = 2 * radius)
  const radius = SPIN_BUTTON_CIRCLE_RADIUS_PER_MENU_HEIGHT * menuHeight
  const targetDiameter = radius * 2
  state.spinBtnSprite.scale.set(targetDiameter / state.spinBtnSprite.height)
  container.addChild(state.spinBtnSprite)

  state.spinBtnSprite.eventMode = 'static'
  state.spinBtnSprite

  state.spinBtnSprite.on('pointerdown', () => {
    if (gameState.showStartScreen?.value) return
    if (gameState.isSpinning?.value) return
    if (!gameState.canSpin?.value) return

    animateSpinButtonClick(state)
    onStartSpin()
    handlers.spin && handlers.spin()
  })

  // Hover effect
  let hoverTl: gsap.core.Timeline | null = null

  state.spinBtnSprite.on('pointerover', () => {
    if (!state.spinBtnSprite || state.isSpinning) return
    if (hoverTl) hoverTl.kill()
    hoverTl = gsap.timeline()
    hoverTl.to(state.spinBtnSprite.scale, {
      x: state.spinBtnSprite.scale.x * 1.05,
      y: state.spinBtnSprite.scale.y * 1.05,
      duration: 0.2,
      ease: 'power2.out'
    })
  })

  state.spinBtnSprite.on('pointerout', () => {
    if (!state.spinBtnSprite || state.isSpinning) return
    if (hoverTl) hoverTl.kill()
    hoverTl = gsap.timeline()
    hoverTl.to(state.spinBtnSprite.scale, {
      x: state.spinBtnSprite.originalScale?.x ?? state.spinBtnSprite.scale.x,
      y: state.spinBtnSprite.originalScale?.y ?? state.spinBtnSprite.scale.y,
      duration: 0.2,
      ease: 'power2.out'
    })
  })

  // Store original scale
  state.spinBtnSprite.originalScale = { x: state.spinBtnSprite.scale.x, y: state.spinBtnSprite.scale.y }

  // Particle containers
  particleState.spinParticlesContainer = new Container()
  particleState.spinParticlesContainer.x = state.spinBtnSprite.x
  particleState.spinParticlesContainer.y = state.spinBtnSprite.y
  container.addChild(particleState.spinParticlesContainer)

  particleState.spinLightningContainer = new Container()
  particleState.spinLightningContainer.x = state.spinBtnSprite.x
  particleState.spinLightningContainer.y = state.spinBtnSprite.y
  container.addChild(particleState.spinLightningContainer)
}

function animateSpinButtonClick(state: SpinButtonState): void {
  if (!state.spinBtnSprite) return

  const originalScaleX = state.spinBtnSprite.originalScale?.x ?? state.spinBtnSprite.scale.x
  const originalScaleY = state.spinBtnSprite.originalScale?.y ?? state.spinBtnSprite.scale.y

  gsap.timeline()
    .to(state.spinBtnSprite.scale, {
      x: originalScaleX * 0.9,
      y: originalScaleY * 0.9,
      duration: 0.1,
      ease: 'power2.in'
    })
    .to(state.spinBtnSprite.scale, {
      x: originalScaleX,
      y: originalScaleY,
      duration: 0.2,
      ease: 'back.out(2)'
    })
}

export function setSpinButtonSpinning(state: SpinButtonState, isSpinning: boolean, canSpin: boolean): void {
  if (!state.spinBtnSprite) return

  const wasSpinning = state.isSpinning
  state.isSpinning = isSpinning

  if (isSpinning && !wasSpinning) {
    // Start spinning animation
    if (state.spinTween) {
      state.spinTween.kill()
    }
    state.spinTween = gsap.to(state.spinBtnSprite, {
      rotation: state.spinBtnSprite.rotation + Math.PI * 2,
      duration: 0.5,
      ease: 'none',
      repeat: -1
    })
  } else if (!isSpinning && wasSpinning) {
    // Stop spinning animation smoothly
    if (state.spinTween) {
      state.spinTween.kill()
    }
    // Animate to nearest full rotation
    const currentRotation = state.spinBtnSprite.rotation
    const targetRotation = Math.ceil(currentRotation / (Math.PI * 2)) * (Math.PI * 2)
    gsap.to(state.spinBtnSprite, {
      rotation: targetRotation,
      duration: 0.3,
      ease: 'power2.out',
      onComplete: () => {
        if (state.spinBtnSprite) {
          state.spinBtnSprite.rotation = 0
        }
      }
    })
  }

  // Update opacity based on canSpin
  if (!isSpinning) {
    state.spinBtnSprite.alpha = canSpin ? 1 : 0.5
  }
}

export function updateSpinButton(
  state: SpinButtonState,
  particleState: ParticleState,
  isSpinning: boolean,
  dt: number
): void {
  // GSAP handles the animation, nothing to do here
}
