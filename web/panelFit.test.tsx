import { render, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { install, sampleCalls, windowBridge } from './testing'
import { usePanelFit } from './panelFit'

// jsdom lays nothing out, so a panel measured with room for its content is this tall per child.
const perChild = 100

function Panel({ refused, ready = true }: { refused: () => void; ready?: boolean }) {
  const panel = usePanelFit<HTMLDivElement>(sampleCalls, refused, ready)
  return (
    <div ref={panel} data-testid="panel" style={{ height: '50px' }}>
      <p>one</p>
      <p>two</p>
    </div>
  )
}

beforeEach(() => {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    return { height: this.style.height === 'auto' ? this.children.length * perChild : 0 } as DOMRect
  })
})

afterEach(() => vi.restoreAllMocks())

describe('usePanelFit (FR-621)', () => {
  it('tells Go the height the content needs, again whenever anything inside changes it, never twice the same', async () => {
    const bridge = install(windowBridge())
    const refused = vi.fn()
    const { getByTestId } = render(<Panel refused={refused} />)
    expect(bridge.FitPanel).toHaveBeenLastCalledWith(2 * perChild)
    const panel = getByTestId('panel')
    expect(panel.style.height).toBe('50px')
    panel.appendChild(document.createElement('p'))
    await waitFor(() => expect(bridge.FitPanel).toHaveBeenLastCalledWith(3 * perChild))
    panel.firstElementChild?.setAttribute('class', 'same height')
    await Promise.resolve()
    expect(bridge.FitPanel).toHaveBeenCalledTimes(2)
    expect(refused).not.toHaveBeenCalled()
  })

  it('measures nothing until Go has made the window the panel, then measures it', () => {
    const bridge = install(windowBridge())
    const { rerender } = render(<Panel refused={vi.fn()} ready={false} />)
    expect(bridge.FitPanel).not.toHaveBeenCalled()
    rerender(<Panel refused={vi.fn()} ready />)
    expect(bridge.FitPanel).toHaveBeenLastCalledWith(2 * perChild)
  })

  it('measures again when the panel itself is resized, since narrower content stacks taller', () => {
    const bridge = install(windowBridge())
    const heard: Array<() => void> = []
    vi.stubGlobal('ResizeObserver', class {
      constructor(private readonly callback: () => void) {}
      observe() { heard.push(this.callback) }
      disconnect() {}
    })
    try {
      render(<Panel refused={vi.fn()} />)
      expect(bridge.FitPanel).toHaveBeenLastCalledWith(2 * perChild)
      const narrower = 3 * perChild
      vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({ height: narrower } as DOMRect)
      heard.forEach((callback) => callback())
      expect(bridge.FitPanel).toHaveBeenLastCalledWith(narrower)
    } finally {
      vi.unstubAllGlobals()
    }
  })
})
