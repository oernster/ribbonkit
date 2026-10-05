import { describe, expect, it, vi } from 'vitest'
import { watchPixelRatio } from './pixelRatio'

/** A window whose ratio can be changed, firing its media query's change as a browser does. */
function fakeView(ratio: number) {
  const listeners = new Set<() => void>()
  const view = {
    devicePixelRatio: ratio,
    matchMedia: () => ({
      addEventListener: (_: string, listener: () => void) => listeners.add(listener),
      removeEventListener: (_: string, listener: () => void) => listeners.delete(listener),
    }),
  }
  const change = (next: number) => {
    view.devicePixelRatio = next
    for (const listener of [...listeners]) listener()
  }
  return { view: view as unknown as Window, change, listeners }
}

describe('watchPixelRatio', () => {
  it('reports the ratio at once and again on every change, then stops', () => {
    const { view, change, listeners } = fakeView(1.25)
    const report = vi.fn()
    const stop = watchPixelRatio(report, view)
    expect(report).toHaveBeenLastCalledWith(1.25)
    change(1.5)
    expect(report).toHaveBeenLastCalledWith(1.5)
    expect(listeners.size).toBe(1)
    stop()
    expect(listeners.size).toBe(0)
  })

  it('reports once on a page with no matchMedia', () => {
    const report = vi.fn()
    watchPixelRatio(report, { devicePixelRatio: 2 } as unknown as Window)()
    expect(report).toHaveBeenCalledWith(2)
  })
})
