import { useState } from 'react'
import type { Refused, WindowCalls } from './bridge'
import { percentOfWhole } from './opacity'

/** The slider moves in whole steps of this many percent. */
const step = 5

interface Props {
  /** The opacity chosen now and the least the setting allows, in percent. */
  opacity: number
  minOpacity: number
  calls: Pick<WindowCalls, 'setOpacity'>
  refused: Refused
  then: () => void
}

/**
 * OpacitySlider chooses how opaque the ribbon is drawn, from the least the setting allows to wholly
 * opaque (FR-622). The panel it sits in stays opaque, so the slider shows the value as it moves; the
 * choice is kept once it is let go, so a drag is one change rather than one for every step it passes.
 */
export function OpacitySlider({ opacity, minOpacity, calls, refused, then }: Props) {
  const [moving, setMoving] = useState<number | null>(null)
  const shown = moving ?? opacity
  const keep = () => {
    if (moving == null) {
      return
    }
    void calls.setOpacity(moving, refused).then(() => {
      setMoving(null)
      then()
    })
  }
  return (
    <fieldset>
      <legend>Opacity</legend>
      <label>
        <input
          type="range"
          min={minOpacity}
          max={percentOfWhole}
          step={step}
          value={shown}
          aria-valuetext={`${shown} percent`}
          onChange={(event) => setMoving(Number(event.target.value))}
          onPointerUp={keep}
          onKeyUp={keep}
          onBlur={keep}
        />
        <span className="opacity-value">{shown}%</span>
      </label>
    </fieldset>
  )
}
