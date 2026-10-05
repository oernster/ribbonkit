// The window's half of the wire between Go and the page, stated a second time here.
// ribbonkit/ui/window/wire.go is the other statement; a structural test compares the two.

export interface AboutFacts {
  name: string
  version: string
  author: string
  copyright: string
  credits: Credit[]
}

export interface UpdateStatus {
  current: string
  latest: string
  updateAvailable: boolean
}

/** A rectangle inside the window, in the page's units: where Go places the ribbon and what is beside it. */
export interface Box {
  x: number
  y: number
  width: number
  height: number
}

export interface Credit {
  name: string
  licence: string
  role: string
}

/** One of the menus' choices: either a group of children or one item whose action goes back to Choose. */
export interface MenuChoice {
  action: string
  label: string
  checkable: boolean
  checked: boolean
  /** Greyed, as a Position item that would leave the ribbon where it stands is. */
  disabled: boolean
  children: MenuChoice[]
}
