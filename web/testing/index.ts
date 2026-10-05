// A stand-in for the window's half of the Go facade in tests: every call is recorded, every answer
// is canned. An application's own stand-in spreads it and adds its own methods.

import { act } from '@testing-library/react'
import { vi } from 'vitest'
import { connect, windowCalls, type WindowBridge } from '../bridge'
import type { AboutFacts } from '../wire'

export { describePageTimers, FRAME, type AllowedSite } from './timers'

/** sampleName stands for the application the kit's tests run for. */
export const sampleName = 'SampleRibbon'

export const about: AboutFacts = {
  name: 'Product', version: '0.1.0', author: 'The Author', copyright: '© The Author',
  credits: [
    { name: 'Go standard library', licence: 'BSD-3-Clause', role: 'the language and its runtime' },
    { name: 'Wails v2', licence: 'MIT', role: 'the desktop shell' },
  ],
}

/** windowBridge answers a recording stand-in for every method of the window's half. */
export function windowBridge() {
  return {
    SetTheme: vi.fn(async () => undefined),
    StartWithWindows: vi.fn(async () => false),
    SetStartWithWindows: vi.fn(async () => undefined),
    SetScrollbar: vi.fn(async () => undefined),
    SetOpacity: vi.fn(async () => undefined),
    BeginScale: vi.fn(async () => undefined),
    DragScale: vi.fn(async () => undefined),
    EndScale: vi.fn(async () => undefined),
    SetScale: vi.fn(async () => undefined),
    SetPixelRatio: vi.fn(async () => undefined),
    SetBackground: vi.fn(async () => undefined),
    RibbonDrawn: vi.fn(async () => undefined),
    TogglePullOut: vi.fn(async () => undefined),
    ShowContextMenu: vi.fn(async () => undefined),
    Choose: vi.fn(async () => undefined),
    OpenPanel: vi.fn(async () => undefined),
    FitPanel: vi.fn(async () => undefined),
    ClosePanel: vi.fn(async () => undefined),
    Hide: vi.fn(async () => undefined),
    OpenDonation: vi.fn(async () => undefined),
    OpenUpdate: vi.fn(async () => undefined),
    SkipUpdate: vi.fn(async () => undefined),
    About: vi.fn(async () => about),
    Licence: vi.fn(async () => 'GNU GENERAL PUBLIC LICENSE\nVersion 3'),
  } satisfies WindowBridge
}

/** install puts bridge on window as the bound App beside a recording drag, then answers it. */
export function install<B extends object>(bridge: B): B {
  window.go = { main: { App: bridge } }
  window.WailsInvoke = vi.fn()
  return bridge
}

/** installEvents puts a runtime on window whose handlers the answered call fires by name. */
export function installEvents() {
  const handlers = new Map<string, (...data: unknown[]) => void>()
  window.runtime = {
    EventsOn: vi.fn((name: string, callback: (...data: unknown[]) => void) => {
      handlers.set(name, callback)
      return () => handlers.delete(name)
    }),
  }
  return (name: string, ...data: unknown[]) => act(() => handlers.get(name)?.(...data))
}

/** sampleCalls are the window's calls over whatever bridge is installed, for the kit's own tests. */
export const sampleCalls = windowCalls(connect<WindowBridge>(sampleName))
