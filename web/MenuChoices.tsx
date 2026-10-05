import type { MenuChoice } from './wire'

/** ChoiceProps is one of the menus' choices and what hands its action back to Go. */
interface ChoiceProps {
  choice: MenuChoice
  choose: (action: string) => void
}

/**
 * MenuGroup draws one of the menus' submenus in an application's Settings: one choice among ticked
 * items as radio buttons, a set of moves such as Position as plain buttons; a disabled item greyed
 * (TimeRibbon FR-624, FR-408).
 */
export function MenuGroup({ choice, choose }: ChoiceProps) {
  return (
    <fieldset>
      <legend>{choice.label}</legend>
      {choice.children.map((item) =>
        item.checkable ? (
          <label key={item.action}>
            <input type="radio" name={choice.label} value={item.action} checked={item.checked} disabled={item.disabled} onChange={() => choose(item.action)} />
            {item.label}
          </label>
        ) : (
          <button key={item.action} type="button" disabled={item.disabled} onClick={() => choose(item.action)}>
            {item.label}
          </button>
        ),
      )}
    </fieldset>
  )
}

/** MenuToggle draws one of the menus' ticked items that stands alone, such as Pin ribbon (TimeRibbon FR-624). */
export function MenuToggle({ choice, choose }: ChoiceProps) {
  return (
    <label>
      <input type="checkbox" checked={choice.checked} onChange={() => choose(choice.action)} />
      {choice.label}
    </label>
  )
}
