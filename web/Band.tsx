import { useEffect, type CSSProperties, type ReactNode, type WheelEvent } from 'react'
import type { Refused, WindowCalls } from './bridge'
import { showsTheMenu, useDrag, type Distance } from './drag'
import { ScaleGrip } from './ScaleGrip'

interface Props {
  /** Whether the window is an unpinned ribbon's tab (FR-614). */
  collapsed: boolean
  /** Whether the ribbon runs top to bottom. */
  vertical: boolean
  /** Whether the content needs more room than the work area offers, so the band scrolls (FR-106). */
  scrolls: boolean
  /** The side the pull out's handle has its lane along; empty while there is no handle (FR-903). */
  lane: string
  /** The application's own classes for the band, its size say. */
  className: string
  /** The application's sizing for what it draws in the band. */
  style: CSSProperties
  dragThreshold: Distance
  calls: Pick<WindowCalls, 'ribbonDrawn' | 'showContextMenu' | 'beginScale' | 'dragScale' | 'endScale' | 'setScale'>
  refused: Refused
  /** gripTip says what dragging and double-clicking the grip do, in the application's words. */
  gripTip: string
  children: ReactNode
}

/**
 * Band is the ribbon around whatever the application shows in it. Pressing empty band and moving past
 * the desktop's drag distance moves the whole window (FR-401); a small wobble or a press on a control
 * does not. A right-click shows the ribbon's menu (FR-108); the corner grip scales it (FR-623).
 */
export function Band({ collapsed, vertical, scrolls, lane, className, style, dragThreshold, calls, refused, gripTip, children }: Props) {
  const drag = useDrag(dragThreshold)

  // An opening ribbon is drawn inside its tab first; Go grows the window once told it has been, so
  // the band is never seen stretched over the full window (FR-615). The second frame is the one
  // after the ribbon was painted.
  useEffect(() => {
    if (collapsed) {
      return
    }
    let frame = window.requestAnimationFrame(() => {
      frame = window.requestAnimationFrame(() => void calls.ribbonDrawn(refused))
    })
    return () => window.cancelAnimationFrame(frame)
  }, [collapsed, calls, refused])

  // An unpinned ribbon's tab is only a band in the scheme's accent: no words, no drag, no menu. It
  // opens when the pointer rests on it, which the desktop reports rather than the page (FR-614).
  if (collapsed) {
    return <div className="tab" data-testid="tab" />
  }

  // A plain wheel moves up and down, which a horizontal ribbon cannot; so while one scrolls, the
  // wheel moves it along instead (FR-106). A sideways wheel or a trackpad already moves it along.
  const wheel = (event: WheelEvent<HTMLDivElement>) => {
    if (!vertical && scrolls && event.deltaX === 0) {
      event.currentTarget.scrollLeft += event.deltaY
    }
  }

  const classes = [
    'ribbon',
    vertical ? 'vertical' : 'horizontal',
    className,
    scrolls ? 'scrolls' : '',
    lane === '' ? '' : `lane-${lane}`,
  ].join(' ')
  return (
    <>
      <div className={classes} style={style} {...drag} onWheel={wheel} onContextMenu={showsTheMenu(calls, refused)}>
        {children}
      </div>
      <ScaleGrip vertical={vertical} calls={calls} refused={refused} tip={gripTip} />
    </>
  )
}
