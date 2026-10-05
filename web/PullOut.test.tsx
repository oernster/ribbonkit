import { fireEvent, render, screen } from '@testing-library/react'
import type { ComponentProps } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { PullOut } from './PullOut'
import { install, sampleCalls, windowBridge } from './testing'

const words = { open: 'Open the part', close: 'Close the part' }
const beside = 'What is pulled out'

/** pullOut answers a vertical ribbon with its part to the left, put back, unless overrides say otherwise. */
function pullOut(overrides: Partial<ComponentProps<typeof PullOut>> = {}) {
  return (
    <PullOut
      side="left"
      ribbon={{ x: 0, y: 0, width: 176, height: 400 }}
      beside={null}
      pulledOut={false}
      scale={100}
      handleLane={16}
      dragThreshold={{ width: 4, height: 4 }}
      calls={sampleCalls}
      refused={vi.fn()}
      words={words}
      band={<div className="band">The ribbon</div>}
      {...overrides}
    >
      <div>{beside}</div>
    </PullOut>
  )
}

describe('PullOut (FR-902, FR-903)', () => {
  it('gives the ribbon a handle that asks Go to pull the part out, then to put it back', () => {
    const bridge = install(windowBridge())
    const { rerender } = render(pullOut())
    expect(screen.queryByText(beside)).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: words.open }))
    expect(bridge.TogglePullOut).toHaveBeenCalledOnce()
    rerender(pullOut({ pulledOut: true, beside: { x: 176, y: 0, width: 480, height: 400 } }))
    expect(screen.getByRole('button', { name: words.close }).getAttribute('title')).toBe(words.close)
    expect(screen.getByText(beside)).toBeTruthy()
  })

  it('points the handle the way the part will move, lying on its side along a horizontal ribbon (Amendment 22)', () => {
    install(windowBridge())
    const { rerender } = render(pullOut({ side: 'bottom', ribbon: { x: 0, y: 0, width: 336, height: 106 } }))
    const handle = screen.getByRole('button', { name: words.open })
    expect(handle.textContent).toBe('▼')
    expect(handle.className).toBe('pull-out-handle bottom')
    rerender(pullOut({ side: 'top', pulledOut: true }))
    expect(screen.getByRole('button', { name: words.close }).textContent).toBe('▼')
  })

  it('draws the handle as deep as its lane at the ribbon\'s scale (Amendment 23, FR-623)', () => {
    install(windowBridge())
    render(pullOut({ scale: 150 }))
    const surface = document.querySelector('.surface') as HTMLElement
    expect(surface.style.getPropertyValue('--handle-width')).toBe('24px')
    expect(surface.style.getPropertyValue('--scale')).toBe('1.5')
  })

  it('answers a right-click and a drag beside a ribbon shorter than its part as the ribbon does, once', () => {
    const bridge = install(windowBridge())
    render(pullOut({ pulledOut: true, beside: { x: 0, y: 106, width: 480, height: 240 } }))
    const surface = document.querySelector('.surface') as HTMLElement
    fireEvent.contextMenu(surface)
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
    // A right-click on the ribbon reaches the surface too; the surface leaves it to the ribbon.
    fireEvent.contextMenu(screen.getByText('The ribbon'))
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
    fireEvent.pointerDown(surface, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(surface, { buttons: 1, screenX: 140, screenY: 100 })
    expect(window.WailsInvoke).toHaveBeenCalledOnce()
    expect(window.WailsInvoke).toHaveBeenCalledWith('drag')
  })

  it('draws the ribbon at the box Go gives, whatever the window\'s width, so one drawn inside the tab fits once the window grows (FR-615)', () => {
    install(windowBridge())
    const width = Object.getOwnPropertyDescriptor(window, 'innerWidth')
    // The page draws an opening ribbon while the window is still the tab, 8 pixels wide; measured
    // 2026-09-29, a vertical ribbon with its pull out on then grew to 175 pixels drawn 8 wide.
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 8 })
    try {
      render(pullOut({ side: 'right', ribbon: { x: 0, y: 0, width: 140, height: 659.2 } }))
      const part = screen.getByText('The ribbon').closest('.surface-part') as HTMLElement
      expect(part.style.width).toBe('140px')
      expect(part.style.height).toBe('659.2px')
    } finally {
      Object.defineProperty(window, 'innerWidth', width ?? { configurable: true, value: 1024 })
    }
  })
})
