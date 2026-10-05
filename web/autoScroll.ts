// The window's side of the self-reading cycle (FR-609). The cycle itself has one home,
// ribbonkit/installer/page/auto-scroll.js, which the setup page loads too; importing it here runs it,
// leaving window.AutoScroll behind. This file only states its shape for the type checker and gives
// React a hook over it.

import { useCallback, useEffect, useState } from 'react'
import '../installer/page/auto-scroll.js'

export type AutoScrollPhase = 'down' | 'pauseBottom' | 'up' | 'pauseTop' | 'manual'

export interface AutoScrollState {
  phase: AutoScrollPhase
  waitMs: number
  ticksToStep: number
  /** opening marks the first hold, which focus arriving may not shorten. */
  opening: boolean
}

export interface ScrollView {
  scrollTop: number
  maxScrollTop: number
}

export interface AutoScrollApi {
  TICK_MS: number
  START_HOLD_MS: number
  DESCENT_PX: number
  DESCENT_TICKS_PER_STEP: number
  BOTTOM_HOLD_MS: number
  REWIND_PX: number
  TOP_HOLD_MS: number
  MANUAL_HOLD_MS: number
  initial(): AutoScrollState
  suspended(state: AutoScrollState): AutoScrollState
  tick(state: AutoScrollState, view: ScrollView): { state: AutoScrollState; delta: number }
  /** attach gives an element the cycle, answering the call that ends it. */
  attach(node: HTMLElement): () => void
}

declare global {
  interface Window {
    AutoScroll: AutoScrollApi
  }
}

export const autoScroll: AutoScrollApi = window.AutoScroll

/**
 * useAutoScroll answers a ref callback for the element that actually scrolls. A panel mounted again
 * starts a fresh cycle, since the element it is handed is a new one.
 */
export function useAutoScroll(): (node: HTMLElement | null) => void {
  const [node, setNode] = useState<HTMLElement | null>(null)
  useEffect(() => (node == null ? undefined : autoScroll.attach(node)), [node])
  return useCallback((next: HTMLElement | null) => setNode(next), [])
}
