import { describe, expect, it } from 'vitest'
import { opacityProperty, showOpacity } from './opacity'

describe('opacity (FR-622)', () => {
  it('draws the window at the fraction of the percentage', () => {
    showOpacity(40)
    expect(document.documentElement.style.getPropertyValue(opacityProperty)).toBe('0.4')
    document.documentElement.style.removeProperty(opacityProperty)
  })
})
