/**
 * watchPixelRatio hands report the page's devicePixelRatio now and again whenever it changes, so Go
 * sizes the window by the scale the page is really drawn at. Windows' text size changes that scale
 * without changing the display's DPI. It answers the call that stops watching. A page with no
 * matchMedia, as under jsdom, reports once.
 */
export function watchPixelRatio(report: (ratio: number) => void, view: Window = window): () => void {
  let query: MediaQueryList | undefined
  const changed = () => {
    query?.removeEventListener('change', changed)
    report(view.devicePixelRatio)
    query = view.matchMedia?.(`(resolution: ${view.devicePixelRatio}dppx)`)
    query?.addEventListener('change', changed)
  }
  changed()
  return () => query?.removeEventListener('change', changed)
}
