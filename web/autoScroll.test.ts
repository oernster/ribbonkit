// The self-reading cycle (FR-609, FR-811), tested against the one script both the window and the
// setup page run. The machine is driven tick by tick; the element half under fake timers, since
// jsdom lays nothing out, so each surface states its own scroll metrics.
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { autoScroll, type AutoScrollState, type ScrollView } from './autoScroll'

const { TICK_MS, START_HOLD_MS, DESCENT_PX, DESCENT_TICKS_PER_STEP, BOTTOM_HOLD_MS, REWIND_PX, TOP_HOLD_MS, MANUAL_HOLD_MS } = autoScroll

const view = (scrollTop: number, maxScrollTop = 1000): ScrollView => ({ scrollTop, maxScrollTop })

/** run drives the machine for ticks against a view that moves with it. */
function run(state: AutoScrollState, ticks: number, start = 0, max = 1000) {
  let scrollTop = start
  let current = state
  for (let i = 0; i < ticks; i++) {
    const result = autoScroll.tick(current, view(scrollTop, max))
    current = result.state
    scrollTop += result.delta
  }
  return { state: current, scrollTop }
}

const ticksFor = (ms: number) => Math.ceil(ms / TICK_MS)
const reading = (): AutoScrollState => ({ phase: 'down', waitMs: 0, ticksToStep: DESCENT_TICKS_PER_STEP, opening: false })

describe('the cycle', () => {
  it('holds still on opening, then reads down', () => {
    const held = run(autoScroll.initial(), ticksFor(START_HOLD_MS) - 1)
    expect(held.state.phase).toBe('pauseTop')
    expect(held.scrollTop).toBe(0)
    const started = run(autoScroll.initial(), ticksFor(START_HOLD_MS))
    expect(started.state.phase).toBe('down')
    expect(started.state.opening).toBe(false)
  })

  it('reads a pixel every second tick and stops exactly at the end', () => {
    expect(run(reading(), 1).scrollTop).toBe(0)
    expect(run(reading(), 2).scrollTop).toBe(DESCENT_PX)
    const end = run(reading(), 4, 998)
    expect(end.scrollTop).toBe(1000)
    expect(end.state.phase).toBe('pauseBottom')
    expect(end.state.waitMs).toBe(BOTTOM_HOLD_MS)
  })

  it('holds at the end, then rewinds far faster than it reads, then holds at the top', () => {
    const held: AutoScrollState = { phase: 'pauseBottom', waitMs: BOTTOM_HOLD_MS, ticksToStep: 1, opening: false }
    expect(run(held, ticksFor(BOTTOM_HOLD_MS) - 1, 1000).state.phase).toBe('pauseBottom')
    expect(run(held, ticksFor(BOTTOM_HOLD_MS), 1000).state.phase).toBe('up')
    const rewinding: AutoScrollState = { ...reading(), phase: 'up' }
    expect(run(rewinding, 1, 500).scrollTop).toBe(500 - REWIND_PX)
    expect(REWIND_PX).toBeGreaterThan(DESCENT_PX * DESCENT_TICKS_PER_STEP)
    const top = run(rewinding, 1, REWIND_PX - 1)
    expect(top.scrollTop).toBe(0)
    expect(top.state).toMatchObject({ phase: 'pauseTop', waitMs: TOP_HOLD_MS })
    expect(run(top.state, ticksFor(TOP_HOLD_MS)).state.phase).toBe('down')
  })

  it('suspends for the manual hold, then resumes down from where the reader left it', () => {
    const held = run(autoScroll.suspended(reading()), ticksFor(MANUAL_HOLD_MS) - 1, 300)
    expect(held.scrollTop).toBe(300)
    expect(held.state.phase).toBe('manual')
    expect(run(autoScroll.suspended(reading()), ticksFor(MANUAL_HOLD_MS), 300).state.phase).toBe('down')
  })

  it('rewinds after a manual hold that left the reader at the very end', () => {
    expect(run(autoScroll.suspended(reading()), ticksFor(MANUAL_HOLD_MS), 1000).state.phase).toBe('up')
  })

  it('consumes nothing while the content fits', () => {
    const opening = autoScroll.initial()
    const { state, delta } = autoScroll.tick(opening, view(0, 0))
    expect(state).toBe(opening)
    expect(delta).toBe(0)
  })
})

/** surface is an element that overflows by 1000 pixels, with a scrollTop jsdom lets it keep. */
function surface(): HTMLElement {
  const node = document.createElement('div')
  let top = 0
  Object.defineProperty(node, 'scrollTop', { get: () => top, set: (value: number) => (top = value) })
  Object.defineProperty(node, 'scrollHeight', { get: () => 1000 })
  Object.defineProperty(node, 'clientHeight', { get: () => 0 })
  document.body.appendChild(node)
  return node
}

const readPast = () => vi.advanceTimersByTime(START_HOLD_MS + TICK_MS * 20)

describe('attach', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => {
    vi.useRealTimers()
    document.body.replaceChildren()
  })

  it('holds still on opening, then reads the element down', () => {
    const node = surface()
    const stop = autoScroll.attach(node)
    vi.advanceTimersByTime(START_HOLD_MS - TICK_MS)
    expect(node.scrollTop).toBe(0)
    vi.advanceTimersByTime(TICK_MS * 20)
    expect(node.scrollTop).toBeGreaterThan(0)
    stop()
  })

  it.each(['wheel', 'mousedown', 'touchstart', 'keydown', 'focusin'])('suspends on %s, then resumes', (type) => {
    const node = surface()
    const stop = autoScroll.attach(node)
    readPast()
    node.dispatchEvent(new Event(type, { bubbles: true }))
    const taken = node.scrollTop
    vi.advanceTimersByTime(TICK_MS * 20)
    expect(node.scrollTop).toBe(taken)
    vi.advanceTimersByTime(MANUAL_HOLD_MS + TICK_MS * 20)
    expect(node.scrollTop).toBeGreaterThan(taken)
    stop()
  })

  it('does not let focus arriving during the opening hold shorten it', () => {
    const node = surface()
    const stop = autoScroll.attach(node)
    node.dispatchEvent(new Event('focusin', { bubbles: true }))
    // Were focus a reader, the manual hold would have ended here and the reading begun.
    vi.advanceTimersByTime(MANUAL_HOLD_MS + TICK_MS * 20)
    expect(node.scrollTop).toBe(0)
    vi.advanceTimersByTime(START_HOLD_MS)
    expect(node.scrollTop).toBeGreaterThan(0)
    stop()
  })

  it('stands frozen beneath a modal taking no input, then carries on where it was', () => {
    const node = surface()
    const modal = document.createElement('div')
    modal.setAttribute('aria-modal', 'true')
    document.body.appendChild(modal)
    const stop = autoScroll.attach(node)
    readPast()
    node.dispatchEvent(new Event('wheel', { bubbles: true }))
    expect(node.scrollTop).toBe(0)
    modal.remove()
    // Nothing was consumed while frozen, the wheel included: the start hold is still owed in full.
    vi.advanceTimersByTime(START_HOLD_MS - TICK_MS)
    expect(node.scrollTop).toBe(0)
    vi.advanceTimersByTime(TICK_MS * 20)
    expect(node.scrollTop).toBeGreaterThan(0)
    stop()
  })

  it('reads a surface inside the topmost modal', () => {
    const modal = document.createElement('div')
    modal.setAttribute('aria-modal', 'true')
    document.body.appendChild(modal)
    const node = surface()
    modal.appendChild(node)
    const stop = autoScroll.attach(node)
    readPast()
    expect(node.scrollTop).toBeGreaterThan(0)
    stop()
  })

  it('stops ticking once ended', () => {
    const stop = autoScroll.attach(surface())
    stop()
    expect(vi.getTimerCount()).toBe(0)
  })
})
