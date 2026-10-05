// What the setup page shows when it never reaches the setup program (FR-808): it waits, gives up
// and says so. Close still closes the window through a way that does not need the setup
// program: the Wails runtime, else the web view's own message channel.

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { activeScreen, focusedLabel, footerButton, footerLabels, layPage, pageElement, type SetupPage } from './setupPage'

/** unreachable is the reason the page gives when the setup program never answered. */
const unreachable = 'Could not reach the setup program.'

/** quitMessage is what Wails' dispatcher reads as quit (dispatcher.go, wails v2.12.0). */
const quitMessage = 'Q'

let setupPage: SetupPage

beforeEach(() => {
  vi.useFakeTimers()
  setupPage = layPage(null)
})

afterEach(() => {
  vi.useRealTimers()
})

/** openPage runs the page's start past every wait for the setup program. */
async function openPage(): Promise<void> {
  const opened = setupPage.init()
  await vi.runAllTimersAsync()
  await opened
}

describe('a page that never reaches the setup program', () => {
  it('says why and closes through the Wails runtime', async () => {
    const quit = vi.fn()
    ;(window as unknown as { runtime: unknown }).runtime = { Quit: quit }
    await openPage()
    expect(activeScreen()).toBe('screen-error')
    expect(pageElement('error-msg').textContent).toBe(unreachable)
    expect(footerLabels()).toEqual(['Close'])
    expect(focusedLabel()).toBe('')
    footerButton('Close').click()
    expect(quit).toHaveBeenCalledTimes(1)
  })

  it('closes through the web view when the Wails runtime is not there either', async () => {
    const postMessage = vi.fn()
    ;(window as unknown as { chrome: unknown }).chrome = { webview: { postMessage } }
    await openPage()
    footerButton('Close').click()
    expect(postMessage).toHaveBeenCalledWith(quitMessage)
  })

  it('draws no Close where nothing could close the window', async () => {
    await openPage()
    expect(activeScreen()).toBe('screen-error')
    expect(footerLabels()).toEqual([])
  })
})
