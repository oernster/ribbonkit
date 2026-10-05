// The shipped setup page laid out for a test, with the fake setup program that answers it.
//
// The setup page has no build step and no test runner of its own, so it is laid out here from the
// very files it ships and its scripts are run the way the page runs them. Every suite over the
// page reads it from here, so no two of them can lay it out differently.

import page from '../../installer/page/index.html?raw'
import ring from '../../installer/page/setup-ring.js?raw'
import reading from '../../installer/page/auto-scroll.js?raw'
import shell from '../../installer/page/setup-shell.js?raw'
import routes from '../../installer/page/setup-routes.js?raw'

/** State is the reading of the machine the setup program hands the page. */
export interface State {
  appName: string
  route: string
  uninstall: boolean
  installedVersion: string
  thisVersion: string
  startMenu: boolean
  desktop: boolean
  startWithWindows: boolean
  prefersDark: boolean
  logPath: string
  problem: string
}

/** Choices are the three boxes the page hands an install or an apply. */
export interface Choices {
  startMenu: boolean
  desktop: boolean
  startWithWindows: boolean
}

/** FakeSetup stands in for the setup program, recording every call that changes anything. */
export class FakeSetup {
  running = false
  quits = 0
  launches = 0
  /** refusal is what every call that changes the machine answers with in place of doing it. */
  refusal = ''
  /** hold, while set, keeps every change pending, so a test can look at the screen mid-work. */
  hold: Promise<void> | null = null
  installs: Choices[] = []
  applies: Choices[] = []
  repairs = 0
  uninstalls: boolean[] = []
  licence = 'the terms'

  answer = (): Promise<void> => {
    if (this.hold !== null) return this.hold
    return this.refusal === '' ? Promise.resolve() : Promise.reject(this.refusal)
  }
  AppRunning = (): Promise<boolean> => Promise.resolve(this.running)
  CloseRunningApp = (): Promise<void> => this.answer()
  Install = (choices: Choices): Promise<void> => {
    this.installs.push(choices)
    return this.answer()
  }
  Repair = (): Promise<void> => {
    this.repairs++
    return this.answer()
  }
  Uninstall = (forget: boolean): Promise<void> => {
    this.uninstalls.push(forget)
    return this.answer()
  }
  Apply = (choices: Choices): Promise<void> => {
    this.applies.push(choices)
    return this.answer()
  }
  LaunchApp = (): Promise<void> => {
    this.launches++
    return Promise.resolve()
  }
  Licence = (): Promise<string> => Promise.resolve(this.licence)
  TakeKeyboard = (): Promise<void> => Promise.resolve()
  Quit = (): void => {
    this.quits++
  }
}

/** The page's own names, as its scripts leave them on the global scope. */
export interface SetupPage {
  route: (state: State) => void
  init: () => Promise<void>
  applyTheme: (theme: string) => void
}

/** Webview is the part of the window the page may close through without the setup program. */
interface Webview {
  go?: unknown
  runtime?: unknown
  chrome?: unknown
}

const bodyOfPage = /<body>([\s\S]*)<\/body>/

/** installed is a machine holding this very version, opened as a double-click opens it. */
export const installed: State = {
  appName: 'Product',
  route: 'manage',
  uninstall: false,
  installedVersion: '0.1.0',
  thisVersion: '0.1.0',
  startMenu: true,
  desktop: false,
  startWithWindows: false,
  prefersDark: false,
  logPath: 'C:\\Temp\\ProductSetup.log',
  problem: '',
}

/** fresh is a machine with nothing installed, offered the defaults of FR-805. */
export const fresh: State = { ...installed, route: 'install', installedVersion: '', startMenu: true }

/** The scripts go in once: the ring's key listener sits on the document for the whole file. */
let ringLoaded = false

/**
 * layPage puts the shipped page on the document and runs its scripts, as the page does. The setup
 * program is the one given; null lays out a page that never reached one.
 */
export function layPage(setup: FakeSetup | null): SetupPage {
  const body = bodyOfPage.exec(page)
  if (body === null) throw new Error('the setup page has no body')
  document.body.innerHTML = body[1]
  const webview = window as unknown as Webview
  delete webview.runtime
  delete webview.chrome
  if (setup === null) delete webview.go
  else webview.go = { main: { App: setup } }
  // An indirect eval runs the scripts in the global scope, as script tags do; the shell and the
  // routes go in as one because the routes read names the shell declares.
  const evaluate = eval
  if (!ringLoaded) {
    evaluate(ring)
    evaluate(reading)
    ringLoaded = true
  }
  evaluate(shell + '\n' + routes)
  return window as unknown as SetupPage
}

/** focusedLabel names what holds focus; empty when nothing does and the page itself holds it. */
export function focusedLabel(): string {
  const focused = document.activeElement
  return focused === null || focused === document.body ? '' : (focused.textContent ?? '')
}

/**
 * layOutByParent stands in for layout, which jsdom does not perform: every element would report no
 * offset parent, so the ring, which passes over what is not on screen, would find no stops at all.
 * An attached element's offset parent becomes its parent. Answers the function that puts it back.
 */
export function layOutByParent(): () => void {
  const original = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetParent')
  Object.defineProperty(HTMLElement.prototype, 'offsetParent', {
    configurable: true,
    get(this: HTMLElement) {
      return this.parentElement
    },
  })
  return () => {
    if (original !== undefined) Object.defineProperty(HTMLElement.prototype, 'offsetParent', original)
    else delete (HTMLElement.prototype as unknown as Record<string, unknown>).offsetParent
  }
}

/** pressTab sends Tab to whatever holds focus, as the keyboard would. */
export function pressTab(): void {
  const event = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true })
  ;(document.activeElement ?? document.body).dispatchEvent(event)
}

/** activeScreen names the screen on show. */
export function activeScreen(): string {
  return document.querySelector('.screen.active')?.id ?? ''
}

/** footerLabels names the footer's buttons in order. */
export function footerLabels(): string[] {
  return Array.from(document.querySelectorAll('#footer .btn')).map((button) => button.textContent ?? '')
}

/** footerButton finds a footer button by its label, failing loudly rather than answering null. */
export function footerButton(label: string): HTMLButtonElement {
  const found = Array.from(document.querySelectorAll<HTMLButtonElement>('#footer .btn')).find(
    (button) => button.textContent === label,
  )
  if (found === undefined) throw new Error(`no ${label} button in the footer`)
  return found
}

/** pageElement finds one element of the page by id, failing loudly rather than answering null. */
export function pageElement(id: string): HTMLElement {
  const found = document.getElementById(id)
  if (found === null) throw new Error(`no #${id} on the page`)
  return found
}

/** boxes answers the boxes a container holds, by their labels, with whether each is ticked. */
export function boxes(container: string): [string, boolean][] {
  return Array.from(pageElement(container).querySelectorAll('.option')).map((option) => [
    option.querySelector('.label')?.textContent ?? '',
    (option.querySelector('input') as HTMLInputElement).checked,
  ])
}

/** settle lets the page's awaited calls to the fake answer. */
export function settle(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 0))
}
