import { useGameStore } from '@/stores'

export interface UseBetControls {
  increaseBet: () => void
  decreaseBet: () => void
}

export function useBetControls(): UseBetControls {
  const gameStore = useGameStore()

  return {
    increaseBet: () => gameStore.increaseBet(),
    decreaseBet: () => gameStore.decreaseBet()
  }
}
