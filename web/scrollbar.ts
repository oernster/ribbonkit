/**
 * scrollbarThickness measures, in DIP, the scroll bar this web engine draws, by laying out a box
 * that must scroll and reading how much of it the bar takes (FR-106). It is the engine's bar, so
 * only the page can measure it; Windows' own scroll bar metric is another size.
 */
export function scrollbarThickness(doc: Document = document): number {
  const box = doc.createElement('div')
  box.style.cssText = 'position:absolute;top:0;left:0;visibility:hidden;overflow:scroll;width:100px;height:100px'
  doc.body.appendChild(box)
  const thickness = box.offsetHeight - box.clientHeight
  box.remove()
  return Math.max(0, thickness)
}
