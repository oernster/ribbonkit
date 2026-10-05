import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// The one test setup for every page built on the kit: the kit's own tests run under it and an
// application's vitest config names it as a setup file, so the two cannot drift apart.

// jsdom has no PointerEvent (measured: typeof PointerEvent was 'undefined'), so a fired pointer
// event carried neither its position nor its buttons. MouseEvent carries both.
if (typeof window.PointerEvent === 'undefined') {
  class PointerEventStandIn extends MouseEvent {}
  window.PointerEvent = PointerEventStandIn as typeof PointerEvent
}

afterEach(() => {
  cleanup()
})
