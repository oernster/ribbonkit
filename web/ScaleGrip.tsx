import { useRef, type PointerEvent } from 'react'
import type { Refused, WindowCalls } from './bridge'
import { percentOfWhole } from './opacity'

interface Props {
  /** Whether the ribbon runs top to bottom, so its thickness is its width. */
  vertical: boolean
  calls: Pick<WindowCalls, 'beginScale' | 'dragScale' | 'endScale' | 'setScale'>
  refused: Refused
  /** tip says what dragging and double-clicking the grip do, in the application's words (NFR-U-4). */
  tip: string
}

/**
 * ScaleGrip is the ribbon's corner grip (FR-623). Dragging it draws the ribbon's content larger or
 * smaller, everything in it together, with the window following as it moves; the scale is kept once
 * it is let go. Double-clicking it draws the content at its own size again. It is a control, so
 * pressing it starts no window drag (FR-402).
 *
 * Go works out the scale: it reads the pointer from the desktop where it can, since the page's own
 * reading was measured jumping backwards while the window was resized under it; the page's reading
 * is sent along for where the desktop cannot give one.
 */
export function ScaleGrip({ vertical, calls, refused, tip }: Props) {
  const dragging = useRef(false)
  const sending = useRef(false)
  const waiting = useRef(false)
  const pointer = useRef({ x: 0, y: 0 })

  // One move on its way at a time; the newest pointer follows it.
  const send = () => {
    if (sending.current) {
      waiting.current = true
      return
    }
    sending.current = true
    void calls.dragScale(pointer.current.x, pointer.current.y, refused).then(() => {
      sending.current = false
      if (waiting.current && dragging.current) {
        waiting.current = false
        send()
      }
    })
  }

  const down = (event: PointerEvent<HTMLDivElement>) => {
    if (event.button !== 0) {
      return
    }
    event.currentTarget.setPointerCapture?.(event.pointerId)
    const ribbon = event.currentTarget.parentElement?.getBoundingClientRect()
    const thickness = (vertical ? ribbon?.width : ribbon?.height) ?? 0
    pointer.current = { x: event.screenX, y: event.screenY }
    dragging.current = true
    waiting.current = false
    void calls.beginScale(thickness, event.screenX, event.screenY, refused)
  }

  const move = (event: PointerEvent<HTMLDivElement>) => {
    if (dragging.current) {
      pointer.current = { x: event.screenX, y: event.screenY }
      send()
    }
  }

  const up = (event: PointerEvent<HTMLDivElement>) => {
    if (dragging.current) {
      dragging.current = false
      void calls.endScale(event.screenX, event.screenY, refused)
    }
  }

  return (
    <div
      className="scale-grip"
      data-control
      role="separator"
      aria-label={tip}
      title={tip}
      onPointerDown={down}
      onPointerMove={move}
      onPointerUp={up}
      onPointerCancel={up}
      onDoubleClick={() => void calls.setScale(percentOfWhole, refused)}
    />
  )
}
