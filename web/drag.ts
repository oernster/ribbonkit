import { useRef, type MouseEvent, type PointerEvent } from 'react'
import { startDrag, type Refused, type WindowCalls } from './bridge'

/** How far the pointer moves, in DIP, before a press becomes a drag; the desktop says (FR-401). */
export interface Distance {
  width: number
  height: number
}

/**
 * showsTheMenu answers the right-click handler that shows the ribbon's own menu (FR-108). The ribbon
 * and the pull out beside it share it, so a right-click on the pull out offers the same menu
 * (Oliver, 2026-09-29).
 */
export function showsTheMenu(calls: Pick<WindowCalls, 'showContextMenu'>, refused: Refused) {
  return (event: MouseEvent<HTMLElement>) => {
    event.preventDefault()
    void calls.showContextMenu(refused)
  }
}

/** Presses on these start no drag (FR-402). */
const controls = 'button, input, select, a, [data-control]'

/**
 * useDrag answers the pointer handlers that move the whole window once a press on anything but a
 * control has moved past threshold (FR-401, FR-402). The ribbon and the pull out share them, so a
 * drag started on the pull out moves both (FR-909).
 */
export function useDrag(threshold: Distance) {
  const pressed = useRef<{ x: number; y: number } | null>(null)
  return {
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      const target = event.target as HTMLElement
      pressed.current = event.button === 0 && target.closest(controls) == null ? { x: event.screenX, y: event.screenY } : null
    },
    onPointerMove: (event: PointerEvent<HTMLElement>) => {
      const start = pressed.current
      if (start == null || event.buttons !== 1) {
        return
      }
      if (Math.abs(event.screenX - start.x) > threshold.width || Math.abs(event.screenY - start.y) > threshold.height) {
        pressed.current = null
        startDrag()
      }
    },
    onPointerUp: () => {
      pressed.current = null
    },
  }
}
