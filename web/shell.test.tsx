import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { swatchId } from './background'
import type { Refused } from './bridge'
import { opacityProperty } from './opacity'
import { useShell, type Drawn, type Panel } from './shell'
import { install, installEvents, sampleCalls, windowBridge } from './testing'

// The application's own open-panel word and the snapshot it answers.
const own = 'own-word'
const opens: Record<string, Panel> = { [own]: 'about' }
const drawn: Drawn = { theme: 'dark', colour: 'neon', opacity: 40 }

afterEach(() => {
  const root = document.documentElement
  root.style.removeProperty(opacityProperty)
  delete root.dataset.theme
  delete root.dataset.colour
  delete window.runtime
})

/** shell installs the window's stand-in and Go's events, then runs the hook over take. */
function shell(take: (refused: Refused) => Promise<Drawn | null> = vi.fn(async () => drawn)) {
  const bridge = install(windowBridge())
  const fire = installEvents()
  const hook = renderHook(() => useShell({ calls: sampleCalls, take, opens }))
  return { bridge, fire, hook, take }
}

describe('useShell', () => {
  it('takes the snapshot, then again whenever Go says something changed (FR-209)', async () => {
    const { fire, hook, take } = shell()
    await waitFor(() => expect(hook.result.current.snapshot).toEqual(drawn))
    await fire('refresh')
    await waitFor(() => expect(take).toHaveBeenCalledTimes(2))
  })

  it('becomes the panel Go names, tells Go which, then says when Go has opened it (CON-6, FR-621)', async () => {
    const { bridge, fire, hook } = shell()
    await fire('open-panel', 'licence')
    expect(hook.result.current.view).toBe('licence')
    expect(bridge.OpenPanel).toHaveBeenCalledWith('licence')
    await waitFor(() => expect(hook.result.current.opened).toBe(true))
    await fire('open-panel', own)
    expect(hook.result.current.view).toBe('about')
    expect(hook.result.current.at).toBe(own)
    await fire('open-panel', 'no such word')
    expect(hook.result.current.view).toBe('settings')
  })

  it('becomes a panel of the application\'s own, asking Go for it by its word', async () => {
    const bridge = install(windowBridge())
    installEvents()
    const take = vi.fn(async () => drawn)
    const hook = renderHook(() =>
      useShell<Drawn, 'detail'>({ calls: sampleCalls, take, opens: { detail: 'detail' } }),
    )
    act(() => hook.result.current.openPanel('detail'))
    expect(hook.result.current.view).toBe('detail')
    expect(bridge.OpenPanel).toHaveBeenCalledWith('detail')
    await waitFor(() => expect(hook.result.current.opened).toBe(true))
  })

  it('carries the update check\'s outcome to the update panel (FR-509)', async () => {
    const { fire, hook } = shell()
    const outcome = { current: '1.0.0', latest: '1.1.0', updateAvailable: true }
    await fire('open-panel', 'update', outcome)
    expect(hook.result.current.view).toBe('update')
    expect(hook.result.current.update).toEqual(outcome)
  })

  it('goes back to the ribbon on closing a panel, then takes the snapshot again', async () => {
    const { bridge, fire, hook, take } = shell()
    await fire('open-panel', 'settings')
    hook.result.current.closePanel()
    await waitFor(() => expect(hook.result.current.view).toBe('ribbon'))
    expect(bridge.ClosePanel).toHaveBeenCalledOnce()
    await waitFor(() => expect(take).toHaveBeenCalledTimes(2))
  })

  it('draws the ribbon at the chosen opacity and a panel wholly opaque (FR-622)', async () => {
    const { fire } = shell()
    const shown = () => document.documentElement.style.getPropertyValue(opacityProperty)
    await waitFor(() => expect(shown()).toBe('0.4'))
    await fire('open-panel', 'settings')
    await waitFor(() => expect(shown()).toBe('1'))
  })

  it('draws the window in the theme and colour scheme the snapshot names (FR-606, FR-611)', async () => {
    const take = vi.fn(async () => drawn)
    const { fire } = shell(take)
    await waitFor(() => expect(document.documentElement.dataset.theme).toBe('dark'))
    expect(document.documentElement.dataset.colour).toBe('neon')
    take.mockImplementation(async () => ({ ...drawn, theme: 'system' }))
    await fire('refresh')
    await waitFor(() => expect(document.documentElement.dataset.theme).toBeUndefined())
  })

  it('tells Go what only the page can measure: the scroll bar, the pixel ratio and the background', async () => {
    const swatch = document.createElement('div')
    swatch.id = swatchId
    swatch.style.backgroundColor = 'rgb(1, 2, 3)'
    document.body.append(swatch)
    try {
      const { bridge } = shell()
      await waitFor(() => expect(bridge.SetScrollbar).toHaveBeenCalledOnce())
      expect(bridge.SetPixelRatio).toHaveBeenCalledWith(window.devicePixelRatio)
      expect(bridge.SetBackground).toHaveBeenCalledWith(1, 2, 3)
    } finally {
      swatch.remove()
    }
  })

  it('says why when Go refuses the snapshot, keeping none', async () => {
    const { hook } = shell(async (refused) => {
      refused('no snapshot')
      return null
    })
    await waitFor(() => expect(hook.result.current.problem).toBe('no snapshot'))
    expect(hook.result.current.snapshot).toBeNull()
  })
})
