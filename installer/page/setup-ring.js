// The keyboard on the setup page (FR-809), ported from BridgeTalk's: written for a page that has no
// build step to share the application's code through.
//
// Tab and Right step forward; Shift+Tab and Left step back; both wrap. The stops are read again on
// every press, because each screen rebuilds its boxes and its footer: a list kept from the last
// screen would name buttons that no longer exist. Enter ticks a box as Space does; a button
// already answers to both.

// setupRing holds the stop the ring last stood on, so a press made while focus rests on no stop,
// as it does once the focused control is disabled or hidden, carries on from there rather than from
// the first. Each screen opens with none (setup-shell.js), so its first press enters at an end.
const setupRing = { mark: null }

// forgetRingMark lets a new screen open with the ring standing nowhere. A function rather than a
// reach into setupRing, since a script's const is its own and a function it declares is the page's.
function forgetRingMark() {
    setupRing.mark = null
}

// ringStops lists what the keyboard can reach in the order it is drawn: every button and box that
// can be used, plus the body while it holds more than fits, since a screen running past the window
// can be read from the keyboard only through its body. A container is otherwise never a stop.
function ringStops() {
    const body = document.querySelector('.body')
    const overflows = body !== null && body.scrollHeight > body.clientHeight
    if (body !== null) body.tabIndex = overflows ? 0 : -1
    return Array.from(document.querySelectorAll('button, input[type="checkbox"], .body')).filter(
        (el) => el.offsetParent !== null && !el.disabled && !el.hidden && (el !== body || overflows),
    )
}

// stepRing moves focus one stop along the ring: delta is 1 forward, -1 back.
function stepRing(delta) {
    const stops = ringStops()
    if (stops.length === 0) return
    const found = stops.indexOf(document.activeElement)
    const from = found >= 0 ? found : setupRing.mark
    const next = from === null
        ? (delta > 0 ? 0 : stops.length - 1)
        : (((from + delta) % stops.length) + stops.length) % stops.length
    setupRing.mark = next
    stops[next].focus()
}

// onRingKey answers the keys the ring owns. It listens in the capture phase, so nothing on the page
// can take a press before the ring sees it.
function onRingKey(event) {
    const forward = !event.shiftKey && (event.key === 'Tab' || event.key === 'ArrowRight')
    const back = event.key === 'ArrowLeft' || (event.key === 'Tab' && event.shiftKey)
    if (forward || back) {
        event.preventDefault()
        stepRing(forward ? 1 : -1)
        return
    }
    const target = event.target
    if (event.key === 'Enter' && target instanceof HTMLInputElement && target.type === 'checkbox') {
        event.preventDefault()
        target.click()
    }
}

document.addEventListener('keydown', onRingKey, true)
