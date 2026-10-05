import { useEffect, useRef } from 'react'
import type { Refused, WindowCalls } from './bridge'

/**
 * The attributes whose change can alter how tall a panel's content is. Style is left out: measuring
 * sets the panel's own, so watching it would measure again at every measurement.
 */
const reshaping = ['class', 'hidden', 'open']

/** naturalHeight answers, in DIP, how tall panel is with room for all its content. */
export function naturalHeight(panel: HTMLElement): number {
  const held = panel.style.height
  panel.style.height = 'auto'
  const height = panel.getBoundingClientRect().height
  panel.style.height = held
  return Math.ceil(height)
}

/**
 * usePanelFit answers a ref for a panel that scrolls its own content. Whenever the height the panel
 * needs to show all of it changes, Go is told; it makes the window that tall where the display has
 * room (FR-621). Anything inside may change it: a clock added, a search answered; so may the panel's
 * own width, since narrower content stacks taller.
 *
 * Nothing is measured until ready: the window becomes the panel only once Go has opened it. A
 * measure taken before then was of the ribbon's narrower window (measured 2026-10-04: 1377 DIP at
 * 750 wide against 1332 at the panel's 900), which then stood until something inside changed. It
 * also reached Go ahead of the opening, which put the window back to its opening height after it.
 */
export function usePanelFit<T extends HTMLElement>(calls: Pick<WindowCalls, 'fitPanel'>, refused: Refused, ready = true) {
  const panel = useRef<T>(null)
  useEffect(() => {
    const element = panel.current
    if (element == null || !ready) {
      return
    }
    let reported = 0
    const fit = () => {
      const height = naturalHeight(element)
      if (height > 0 && height !== reported) {
        reported = height
        void calls.fitPanel(height, refused)
      }
    }
    fit()
    void document.fonts?.ready.then(fit)
    const watcher = new MutationObserver(fit)
    watcher.observe(element, { childList: true, subtree: true, characterData: true, attributeFilter: reshaping })
    const resized = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(fit)
    resized?.observe(element)
    return () => {
      watcher.disconnect()
      resized?.disconnect()
    }
  }, [calls, refused, ready])
  return panel
}
