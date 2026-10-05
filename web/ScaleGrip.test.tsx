import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { percentOfWhole } from './opacity'
import { ScaleGrip } from './ScaleGrip'
import { install, sampleCalls, windowBridge } from './testing'

// The ribbon the grip sits in, in the page's pixels across it; the tip it is named by.
const thickness = 100
const tip = 'Drag to resize'

afterEach(() => vi.restoreAllMocks())

/** grip draws the grip in a ribbon twice as long as it is thick, answering the grip. */
function grip(vertical = false) {
  render(
    <div>
      <ScaleGrip vertical={vertical} calls={sampleCalls} refused={vi.fn()} tip={tip} />
    </div>,
  )
  const element = screen.getByRole('separator', { name: tip })
  vi.spyOn(element.parentElement as HTMLElement, 'getBoundingClientRect').mockReturnValue({ width: thickness * 2, height: thickness } as DOMRect)
  return element
}

describe('ScaleGrip (FR-623)', () => {
  it('hands Go the drag: its start with the ribbon thickness, each move and its end', async () => {
    const bridge = install(windowBridge())
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 5, screenY: 7 })
    expect(bridge.BeginScale).toHaveBeenCalledWith(thickness, 5, 7)
    fireEvent.pointerMove(handle, { screenX: 5, screenY: 57 })
    await vi.waitFor(() => expect(bridge.DragScale).toHaveBeenCalledWith(5, 57))
    expect(bridge.EndScale).not.toHaveBeenCalled()
    fireEvent.pointerUp(handle, { screenX: 5, screenY: 60 })
    expect(bridge.EndScale).toHaveBeenCalledWith(5, 60)
  })

  it('measures a vertical ribbon across its width', () => {
    const bridge = install(windowBridge())
    fireEvent.pointerDown(grip(true), { button: 0, screenX: 0, screenY: 0 })
    expect(bridge.BeginScale).toHaveBeenCalledWith(thickness * 2, 0, 0)
  })

  it('sends one move at a time, the newest pointer following it', async () => {
    const bridge = install(windowBridge())
    let finish: () => void = () => undefined
    bridge.DragScale.mockImplementationOnce(() => new Promise<undefined>((done) => { finish = () => done(undefined) }))
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 0, screenY: 0 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 10 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 20 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 30 })
    expect(bridge.DragScale).toHaveBeenCalledTimes(1)
    finish()
    await vi.waitFor(() => expect(bridge.DragScale).toHaveBeenCalledTimes(2))
    expect(bridge.DragScale).toHaveBeenLastCalledWith(0, 30)
  })

  it('a double-click returns the own size; a move with no press sends nothing', () => {
    const bridge = install(windowBridge())
    const handle = grip()
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 30 })
    expect(bridge.DragScale).not.toHaveBeenCalled()
    fireEvent.doubleClick(handle)
    expect(bridge.SetScale).toHaveBeenCalledWith(percentOfWhole)
  })
})
