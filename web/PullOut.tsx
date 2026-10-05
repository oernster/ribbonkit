import type { CSSProperties, MouseEvent, PointerEvent, ReactNode } from 'react'
import type { Refused, WindowCalls } from './bridge'
import { showsTheMenu, useDrag, type Distance } from './drag'
import { percentOfWhole } from './opacity'
import type { Box } from './wire'

/** The handle's words, in the application's terms: each says what pressing it will do (NFR-U-4). */
export interface PullOutWords {
  open: string
  close: string
}

interface Props {
  /** The side of the ribbon the pulled out part adjoins: left, right, top or bottom. */
  side: string
  /** Where Go places the ribbon, in the page's units. */
  ribbon: Box
  /** Where Go places what is pulled out; null while it is put back. */
  beside: Box | null
  /** Whether it is pulled out, which turns the handle's arrow and words round. */
  pulledOut: boolean
  /** The ribbon's scale in percent, which the handle is drawn at (FR-623). */
  scale: number
  /** How deep the handle's lane is at full size, in DIP (FR-903). */
  handleLane: number
  dragThreshold: Distance
  calls: Pick<WindowCalls, 'togglePullOut' | 'showContextMenu'>
  refused: Refused
  words: PullOutWords
  /** The ribbon itself. */
  band: ReactNode
  /** What is pulled out beside it. */
  children: ReactNode
}

/** The handle's arrows point the way the part will move: out towards side when in, back when out. */
const arrows: Record<string, { out: string; back: string }> = {
  left: { out: '◀', back: '▶' },
  right: { out: '▶', back: '◀' },
  top: { out: '▲', back: '▼' },
  bottom: { out: '▼', back: '▲' },
}

/** A box in the page's units, as a style places it. */
interface Placed {
  left: number
  top: number
  width: number
  height: number
}

/** placed answers box as a style places it. */
function placed(box: Box): Placed {
  return { left: box.x, top: box.y, width: box.width, height: box.height }
}

/**
 * handleAt answers where the handle is anchored: on the ribbon's side, half way along it. The
 * handle's style for that side then draws it just inside the ribbon.
 */
function handleAt(side: string, ribbon: Placed): { left: number; top: number } {
  const middle = { left: ribbon.left + ribbon.width / 2, top: ribbon.top + ribbon.height / 2 }
  switch (side) {
    case 'left':
      return { left: ribbon.left, top: middle.top }
    case 'right':
      return { left: ribbon.left + ribbon.width, top: middle.top }
    case 'top':
      return { left: middle.left, top: ribbon.top }
    default:
      return { left: middle.left, top: ribbon.top + ribbon.height }
  }
}

/**
 * PullOut is the window's content while the ribbon has a part to pull out beside it, horizontal or
 * vertical, with the handle that pulls it out and puts it back (FR-902, FR-903). Go places both, in
 * the page's units, so each is drawn as it comes. Scaling by the window's width drew an opening
 * ribbon, which is drawn while the window is still its tab (FR-615), 8 wide in a window then grown to
 * 175, blank and deaf to a right-click (measured 2026-09-29).
 */
export function PullOut({ side, ribbon, beside, pulledOut, scale, handleLane, dragThreshold, calls, refused, words, band, children }: Props) {
  const drag = useDrag(dragThreshold)
  const at = placed(ribbon)
  const arrow = arrows[side]
  // The handle stands in the lane, which is drawn at the ribbon's scale (FR-623).
  const drawnAt = scale / percentOfWhole
  const lane = { '--handle-width': `${handleLane * drawnAt}px`, '--scale': drawnAt } as CSSProperties
  // On macOS and Linux the window stays a rectangle (FR-913), so beside a ribbon shorter than what
  // is pulled out the surface itself shows, painted as the ribbon is; it answers a right-click and a
  // drag as the ribbon does. A press on either part reaches here too, which they already answer, so
  // only one on the surface itself counts.
  const menu = showsTheMenu(calls, refused)
  const own = {
    ...drag,
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      if (event.target === event.currentTarget) {
        drag.onPointerDown(event)
      }
    },
    onContextMenu: (event: MouseEvent<HTMLElement>) => {
      if (event.target === event.currentTarget) {
        menu(event)
      }
    },
  }
  const said = pulledOut ? words.close : words.open
  return (
    <div className="surface" style={lane} {...own}>
      <div className="surface-part" style={at}>
        {band}
      </div>
      {beside != null && (
        <div className="surface-part" style={placed(beside)}>
          {children}
        </div>
      )}
      {arrow !== undefined && (
        <button
          type="button"
          className={`pull-out-handle ${side}`}
          style={handleAt(side, at)}
          aria-label={said}
          title={said}
          onClick={() => void calls.togglePullOut(refused)}
        >
          {pulledOut ? arrow.back : arrow.out}
        </button>
      )}
    </div>
  )
}
