import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { MenuGroup, MenuToggle } from './MenuChoices'
import { choiceGroup, choiceItem } from './testing'

describe('the menus\' choices in Settings (FR-624, FR-408)', () => {
  it('draws a group of ticked items as radio buttons with the ticked one chosen, handing a choice back', () => {
    const choose = vi.fn()
    render(<MenuGroup choice={choiceGroup('Colour', [choiceItem('colour-classic', 'Classic', true), choiceItem('colour-neon', 'Neon', false)])} choose={choose} />)
    expect((screen.getByRole('radio', { name: 'Classic' }) as HTMLInputElement).checked).toBe(true)
    fireEvent.click(screen.getByRole('radio', { name: 'Neon' }))
    expect(choose).toHaveBeenCalledWith('colour-neon')
  })

  it('draws moves as buttons, greying one that would not move the ribbon', () => {
    const greyed = { ...choiceItem('right-edge', 'Centre on right edge'), disabled: true }
    render(<MenuGroup choice={choiceGroup('Position', [choiceItem('left-edge', 'Centre on left edge'), greyed])} choose={vi.fn()} />)
    expect((screen.getByRole('button', { name: 'Centre on right edge' }) as HTMLButtonElement).disabled).toBe(true)
    expect((screen.getByRole('button', { name: 'Centre on left edge' }) as HTMLButtonElement).disabled).toBe(false)
  })

  it('draws an item standing alone as a checkbox handing its action back', () => {
    const choose = vi.fn()
    render(<MenuToggle choice={choiceItem('pin', 'Pin ribbon', true)} choose={choose} />)
    const box = screen.getByRole('checkbox', { name: 'Pin ribbon' }) as HTMLInputElement
    expect(box.checked).toBe(true)
    fireEvent.click(box)
    expect(choose).toHaveBeenCalledWith('pin')
  })
})
