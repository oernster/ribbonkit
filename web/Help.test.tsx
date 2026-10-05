import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { about, install, sampleCalls, windowBridge } from './testing'
import { About, Licence, Update } from './Help'
import type { UpdateStatus } from './wire'
import { autoScroll } from './autoScroll'

// icon stands for the application's picture at the head of About.
const icon = 'app-icon.png'

afterEach(() => vi.useRealTimers())

async function open(panel: 'about' | 'licence') {
  const bridge = install(windowBridge())
  const onClose = vi.fn()
  await act(async () => {
    render(panel === 'about' ? <About onClose={onClose} calls={sampleCalls} icon={icon} /> : <Licence onClose={onClose} calls={sampleCalls} />)
  })
  return { bridge, onClose }
}

describe('About (FR-607)', () => {
  it('shows the icon, the name and version, the author, the copyright, then every credit, in that order', async () => {
    await open('about')
    const body = document.querySelector('.panel-body') as HTMLElement
    const order = ['IMG', 'Product 0.1.0', 'by The Author', '© The Author', 'Credits'].map((mark) =>
      Array.from(body.querySelectorAll('*')).findIndex((el) => (mark === 'IMG' ? el.tagName === mark : el.textContent === mark)),
    )
    expect(order.every((at) => at >= 0)).toBe(true)
    expect([...order].sort((a, b) => a - b)).toEqual(order)
    const credits = Array.from(body.querySelectorAll('.credits li')).map((li) => li.textContent)
    expect(credits).toEqual(about.credits.map((c) => `${c.name}, ${c.licence}: ${c.role}`))
  })

  it('opens on Close; Close and Escape both return to the ribbon', async () => {
    const { onClose } = await open('about')
    expect(document.activeElement?.textContent).toBe('Close')
    fireEvent.click(screen.getByText('Close'))
    fireEvent.keyDown(screen.getByText('About'), { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('says why when About cannot be read', async () => {
    const bridge = install(windowBridge())
    bridge.About.mockRejectedValueOnce('no facade')
    await act(async () => {
      render(<About onClose={vi.fn()} calls={sampleCalls} icon={icon} />)
    })
    expect(screen.getByRole('alert').textContent).toBe('no facade')
  })
})

describe('Update (FR-509)', () => {
  const newer: UpdateStatus = { current: '2.0.0', latest: 'v2.1.0', updateAvailable: true }

  function showUpdate(status: UpdateStatus) {
    const bridge = install(windowBridge())
    const onClose = vi.fn()
    render(<Update status={status} onClose={onClose} calls={sampleCalls} />)
    return { bridge, onClose }
  }

  it('offers a newer release with Download, Skip this version and Later', () => {
    showUpdate(newer)
    expect(screen.getByText('Update available')).toBeTruthy()
    expect(screen.getByText('Version v2.1.0 is available. You are running 2.0.0.')).toBeTruthy()
    const choices = Array.from(document.querySelectorAll('.update-actions button')).map((b) => b.textContent)
    expect(choices).toEqual(['Download', 'Skip this version', 'Later'])
  })

  it('asks Go to open or skip what it offered, then returns to the ribbon; Later only returns', async () => {
    const { bridge, onClose } = showUpdate(newer)
    await act(async () => fireEvent.click(screen.getByText('Download')))
    await act(async () => fireEvent.click(screen.getByText('Skip this version')))
    fireEvent.click(screen.getByText('Later'))
    expect(bridge.OpenUpdate).toHaveBeenCalledTimes(1)
    expect(bridge.SkipUpdate).toHaveBeenCalledTimes(1)
    expect(onClose).toHaveBeenCalledTimes(3)
  })

  it('stays open and says why when Go refuses', async () => {
    const { bridge, onClose } = showUpdate(newer)
    bridge.OpenUpdate.mockRejectedValueOnce('no browser; it is at https://example.test')
    await act(async () => fireEvent.click(screen.getByText('Download')))
    expect(screen.getByRole('alert').textContent).toBe('no browser; it is at https://example.test')
    expect(onClose).not.toHaveBeenCalled()
  })

  it('says when this is the latest version and when GitHub could not be reached', () => {
    showUpdate({ current: '2.0.0', latest: 'v2.0.0', updateAvailable: false })
    expect(screen.getByText('You are running the latest version.')).toBeTruthy()
    expect(document.querySelector('.update-actions')).toBeNull()
    showUpdate({ current: '2.0.0', latest: '', updateAvailable: false })
    expect(screen.getByText('The update check could not reach GitHub. Please try again later.')).toBeTruthy()
  })
})

describe('Licence (FR-608)', () => {
  it('shows the whole of the terms the facade answers', async () => {
    await open('licence')
    expect(document.querySelector('.licence-text')?.textContent).toBe('GNU GENERAL PUBLIC LICENSE\nVersion 3')
  })

  it('reads itself when it holds more than fits (FR-609)', async () => {
    vi.useFakeTimers()
    await open('licence')
    const body = document.querySelector('.panel-body') as HTMLElement
    let top = 0
    Object.defineProperty(body, 'scrollTop', { get: () => top, set: (value: number) => (top = value) })
    Object.defineProperty(body, 'scrollHeight', { get: () => 1000 })
    Object.defineProperty(body, 'clientHeight', { get: () => 0 })
    act(() => vi.advanceTimersByTime(autoScroll.START_HOLD_MS - autoScroll.TICK_MS))
    expect(top).toBe(0)
    act(() => vi.advanceTimersByTime(autoScroll.TICK_MS * 20))
    expect(top).toBeGreaterThan(0)
  })
})
