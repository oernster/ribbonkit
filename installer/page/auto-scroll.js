// The self-reading cycle for long help content (FR-609, FR-811), ported from PigeonPost's
// autoScroll.ts and useAutoScroll.ts: hold still on opening, read down slowly, hold at the tail,
// rewind fast, repeat; step aside the moment the reader takes over.
//
// It has one home for both surfaces that wear it. The setup page has no build step, so it can load
// a script but import nothing; the window's build can import this file. So it lives here beside
// the setup page; the window imports it. It publishes one name, window.AutoScroll, which is
// what a classic script tag, an ES import and the setup tests' indirect eval all leave behind.
//
// The pace is the application's, never a surface's: if one surface seems to need another pace, the
// pace is wrong everywhere.
(function () {
    // TICK_MS is the clock. Every hold is counted down in whole ticks.
    const TICK_MS = 40
    // START_HOLD_MS is the stillness before the first descent: the reader orients first.
    const START_HOLD_MS = 5000
    // The descent is one pixel every second tick; the divider counts ticks so holds keep TICK_MS.
    const DESCENT_PX = 1
    const DESCENT_TICKS_PER_STEP = 2
    // BOTTOM_HOLD_MS is long enough to finish the tail before the rewind takes it away.
    const BOTTOM_HOLD_MS = 5000
    // REWIND_PX is a reposition rather than a reading pass, so it travels fast.
    const REWIND_PX = 15
    // TOP_HOLD_MS is the breath before the next pass.
    const TOP_HOLD_MS = 2000
    // MANUAL_HOLD_MS is the stillness after any reading by hand before the cycle picks up again,
    // from wherever the reader left it. Reading by hand suspends the cycle; it never ends it.
    const MANUAL_HOLD_MS = 2500

    // MANUAL_EVENTS count as reading by hand. mousedown covers a press on the scroll bar; focusin
    // covers the keyboard arriving in the surface.
    const MANUAL_EVENTS = ['wheel', 'mousedown', 'touchstart', 'keydown', 'focusin']

    // initial opens in the top hold seeded with the start hold. opening marks that hold as the
    // first one, which focus arriving may not shorten.
    function initial() {
        return { phase: 'pauseTop', waitMs: START_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP, opening: true }
    }

    // suspended is the hold a reader taking over puts the cycle in; it resumes from their position.
    function suspended(state) {
        return { ...state, phase: 'manual', waitMs: MANUAL_HOLD_MS, opening: false }
    }

    // tick advances the cycle by one tick and answers how far the surface should move, clamped to
    // its bounds. Content that does not overflow consumes nothing, so a surface that fits is free.
    function tick(state, view) {
        if (view.maxScrollTop <= 0) return { state, delta: 0 }
        if (state.phase === 'down') return descend(state, view)
        if (state.phase === 'up') return rewind(state, view)
        return hold(state, view)
    }

    // hold counts the wait down. When it runs out the opening is over: after the bottom hold comes
    // the rewind, after a manual hold at the very end the rewind too, otherwise the reading pass.
    function hold(state, view) {
        const waitMs = state.waitMs - TICK_MS
        if (waitMs > 0) return { state: { ...state, waitMs }, delta: 0 }
        const rewinding = state.phase === 'pauseBottom' || (state.phase === 'manual' && view.scrollTop >= view.maxScrollTop)
        if (rewinding) return { state: { ...state, phase: 'up', waitMs: 0, opening: false }, delta: 0 }
        return { state: { phase: 'down', waitMs: 0, ticksToStep: DESCENT_TICKS_PER_STEP, opening: false }, delta: 0 }
    }

    function descend(state, view) {
        const ticksToStep = state.ticksToStep - 1
        if (ticksToStep > 0) return { state: { ...state, ticksToStep }, delta: 0 }
        const remaining = view.maxScrollTop - view.scrollTop
        if (remaining <= DESCENT_PX) {
            return {
                state: { phase: 'pauseBottom', waitMs: BOTTOM_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP, opening: false },
                delta: Math.max(0, remaining),
            }
        }
        return { state: { ...state, ticksToStep: DESCENT_TICKS_PER_STEP }, delta: DESCENT_PX }
    }

    function rewind(state, view) {
        if (view.scrollTop <= REWIND_PX) {
            return {
                state: { phase: 'pauseTop', waitMs: TOP_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP, opening: false },
                delta: -view.scrollTop,
            }
        }
        return { state, delta: -REWIND_PX }
    }

    // frozen answers whether a dialog marked modal stands above the surface. Two surfaces reading
    // at once compete for the eye, so one beneath a modal holds exactly where it is.
    function frozen(node) {
        const modals = document.querySelectorAll('[aria-modal="true"]')
        if (modals.length === 0) return false
        return !modals[modals.length - 1].contains(node)
    }

    // attach gives node, the element that actually scrolls, the cycle, answering the call that
    // ends it. A frozen surface takes no input at all: nothing reaching it can be a reader. Nor may
    // the page's own reset as a modal closes read as one.
    function attach(node) {
        let state = initial()
        const onManual = (event) => {
            if (frozen(node)) return
            if (event.type === 'focusin' && state.opening) return
            state = suspended(state)
        }
        MANUAL_EVENTS.forEach((type) => node.addEventListener(type, onManual, { passive: true }))
        const timer = window.setInterval(() => {
            if (frozen(node)) return
            const view = { scrollTop: node.scrollTop, maxScrollTop: node.scrollHeight - node.clientHeight }
            const next = tick(state, view)
            state = next.state
            if (next.delta !== 0) node.scrollTop = view.scrollTop + next.delta
        }, TICK_MS)
        return () => {
            window.clearInterval(timer)
            MANUAL_EVENTS.forEach((type) => node.removeEventListener(type, onManual))
        }
    }

    window.AutoScroll = {
        TICK_MS, START_HOLD_MS, DESCENT_PX, DESCENT_TICKS_PER_STEP, BOTTOM_HOLD_MS, REWIND_PX, TOP_HOLD_MS, MANUAL_HOLD_MS,
        initial, suspended, tick, attach,
    }
})()
