/** The CSS property #root's opacity reads (app.css). */
export const opacityProperty = '--window-opacity'

/** percentOfWhole is wholly opaque; dividing by it turns a percentage into CSS opacity's fraction. */
export const percentOfWhole = 100

/** showOpacity draws the window's backgrounds at percent opaque, the clocks on them wholly so (FR-622). */
export function showOpacity(percent: number, doc: Document = document): void {
  doc.documentElement.style.setProperty(opacityProperty, String(percent / percentOfWhole))
}
