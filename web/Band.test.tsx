import { fireEvent, render, screen } from '@testing-library/react'
import type { ComponentProps } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Band } from './Band'
import { scrollbarThickness } from './scrollbar'
import { install, sampleCalls, windowBridge } from './testing'

// The distance a press moves before it is a drag; the words the grip is named by.
const threshold = { width: 4, height: 4 }
const tip = 'Drag to resize'
const control = 'A control'

afterEach(() => vi.restoreAllMocks())

/** band answers a band holding one control, horizontal and open unless overrides say otherwise. */
function band(overrides: Partial<ComponentProps<typeof Band>> = {}) {
  return (
    <Band
      collapsed={false}
      vertical={false}
      scrolls={false}
      lane=""
      className="large"
      style={{}}
      dragThreshold={threshold}
      calls={sampleCalls}
      refused={vi.fn()}
      gripTip={tip}
      {...overrides}
    >
      <button type="button">{control}</button>
    </Band>
  )
}

function ribbon(): HTMLElement {
  return document.querySelector('.ribbon') as HTMLElement
}

describe('Band', () => {
  it('draws only the tab while collapsed: no words, no drag, no menu (FR-614)', () => {
    const bridge = install(windowBridge())
    render(band({ collapsed: true }))
    const tab = screen.getByTestId('tab')
    expect(tab.textContent).toBe('')
    expect(screen.queryByText(control)).toBeNull()
    fireEvent.pointerDown(tab, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(tab, { buttons: 1, screenX: 140, screenY: 100 })
    fireEvent.contextMenu(tab)
    expect(window.WailsInvoke).not.toHaveBeenCalled()
    expect(bridge.ShowContextMenu).not.toHaveBeenCalled()
  })

  it('says it has drawn the full ribbon once painted, never for the tab (FR-615)', async () => {
    const bridge = install(windowBridge())
    const { rerender } = render(band({ collapsed: true }))
    await new Promise((settled) => window.requestAnimationFrame(() => window.requestAnimationFrame(settled)))
    expect(bridge.RibbonDrawn).not.toHaveBeenCalled()
    rerender(band())
    expect(bridge.RibbonDrawn).not.toHaveBeenCalled()
    await vi.waitFor(() => expect(bridge.RibbonDrawn).toHaveBeenCalledOnce())
  })

  it('carries the application\'s classes and its handle\'s lane, none without a handle (FR-903)', () => {
    install(windowBridge())
    const { rerender } = render(band({ vertical: true, scrolls: true }))
    expect([...ribbon().classList]).toEqual(['ribbon', 'vertical', 'large', 'scrolls'])
    rerender(band({ lane: 'bottom' }))
    expect([...ribbon().classList]).toEqual(['ribbon', 'horizontal', 'large', 'lane-bottom'])
  })

  it('starts a drag only past the threshold and never from a control (FR-401, FR-402)', () => {
    install(windowBridge())
    render(band())
    fireEvent.pointerDown(ribbon(), { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(ribbon(), { buttons: 1, screenX: 103, screenY: 102 })
    expect(window.WailsInvoke).not.toHaveBeenCalled()
    fireEvent.pointerMove(ribbon(), { buttons: 1, screenX: 105, screenY: 100 })
    expect(window.WailsInvoke).toHaveBeenCalledWith('drag')
    const invoke = window.WailsInvoke as ReturnType<typeof vi.fn>
    invoke.mockClear()
    const button = screen.getByText(control)
    fireEvent.pointerDown(button, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(button, { buttons: 1, screenX: 140, screenY: 100 })
    expect(invoke).not.toHaveBeenCalled()
  })

  it('moves only a scrolling horizontal ribbon along with a plain wheel (FR-106)', () => {
    install(windowBridge())
    const along = (overrides: Partial<ComponentProps<typeof Band>>, deltaX: number) => {
      const view = render(band(overrides))
      let left = 0
      Object.defineProperty(ribbon(), 'scrollLeft', { get: () => left, set: (value: number) => (left = value) })
      fireEvent.wheel(ribbon(), { deltaX, deltaY: 100 })
      view.unmount()
      return left
    }
    expect(along({ scrolls: true }, 0)).toBe(100)
    expect(along({ scrolls: true }, 30)).toBe(0)
    expect(along({ scrolls: false }, 0)).toBe(0)
    expect(along({ vertical: true, scrolls: true }, 0)).toBe(0)
  })

  it('opens the native menu on right-click (FR-108)', () => {
    const bridge = install(windowBridge())
    render(band())
    fireEvent.contextMenu(screen.getByText(control))
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
  })

  it('carries the corner grip, named in the application\'s words (FR-623)', () => {
    install(windowBridge())
    render(band())
    expect(screen.getByRole('separator', { name: tip })).toBeTruthy()
  })
})

describe('scrollbarThickness', () => {
  it('measures the scroll bar as the room it takes from a box that must scroll (FR-106)', () => {
    expect(scrollbarThickness()).toBe(0)
    vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(100)
    vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(85)
    expect(scrollbarThickness()).toBe(15)
    expect(document.body.children.length).toBe(0)
  })
})
