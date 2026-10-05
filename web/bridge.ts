// The page's way to the Go facade a ribbon's application binds. Wails injects window.go.main.App and
// window.runtime at load time; the kit names the calls its window answers (ribbonkit/ui/window) and
// the application adds its own over the same guarded call.
//
// Every call that Go can refuse takes a refusal handler as its last argument and answers null
// rather than rejecting, so a call without a handler does not compile (ported from Bridge Talk).

import type { AboutFacts } from './wire'

/** A handler told, in words, why a call was refused. */
export type Refused = (reason: string) => void

/** The window's methods the page calls, as Go's window.Window states them. */
export interface WindowBridge {
  SetTheme(theme: string): Promise<void>
  StartWithWindows(): Promise<boolean>
  SetStartWithWindows(on: boolean): Promise<void>
  SetScrollbar(dip: number): Promise<void>
  SetOpacity(percent: number): Promise<void>
  BeginScale(thickness: number, x: number, y: number): Promise<void>
  DragScale(x: number, y: number): Promise<void>
  EndScale(x: number, y: number): Promise<void>
  SetScale(percent: number): Promise<void>
  SetPixelRatio(ratio: number): Promise<void>
  SetBackground(red: number, green: number, blue: number): Promise<void>
  RibbonDrawn(): Promise<void>
  TogglePullOut(): Promise<void>
  ShowContextMenu(): Promise<void>
  Choose(action: string): Promise<void>
  OpenPanel(panel: string): Promise<void>
  FitPanel(height: number): Promise<void>
  ClosePanel(): Promise<void>
  Hide(): Promise<void>
  OpenDonation(): Promise<void>
  OpenUpdate(): Promise<void>
  SkipUpdate(): Promise<void>
  About(): Promise<AboutFacts>
  Licence(): Promise<string>
}

interface Runtime {
  EventsOn(name: string, callback: (...data: unknown[]) => void): () => void
}

declare global {
  interface Window {
    go?: { main?: { App?: unknown } }
    runtime?: Runtime
    WailsInvoke?: (message: string) => void
  }
}

/** Call runs one facade call over bridge B, answering null and telling refused why when it cannot. */
export type Call<B> = <T>(run: (bridge: B) => Promise<T>, refused: Refused) => Promise<T | null>

/**
 * connect answers the guarded call over the App Wails bound. Outside Wails (where the application,
 * named by name, is not behind the page) and where Go refuses, the call answers null and refused
 * hears why.
 */
export function connect<B>(name: string): Call<B> {
  const unreachable = `${name} is not running behind this page`
  return async (run, refused) => {
    const bridge = window.go?.main?.App as B | undefined
    if (bridge == null) {
      refused(unreachable)
      return null
    }
    try {
      return await run(bridge)
    } catch (failure) {
      refused(String(failure))
      return null
    }
  }
}

/** windowCalls answers the window's calls over call, each answering null where it was refused. */
export function windowCalls(call: Call<WindowBridge>) {
  return {
    setTheme: (theme: string, refused: Refused) => call((b) => b.SetTheme(theme), refused),
    startWithWindows: (refused: Refused) => call((b) => b.StartWithWindows(), refused),
    setStartWithWindows: (on: boolean, refused: Refused) => call((b) => b.SetStartWithWindows(on), refused),
    setScrollbar: (dip: number, refused: Refused) => call((b) => b.SetScrollbar(dip), refused),
    setOpacity: (percent: number, refused: Refused) => call((b) => b.SetOpacity(percent), refused),
    beginScale: (thickness: number, x: number, y: number, refused: Refused) =>
      call((b) => b.BeginScale(thickness, x, y), refused),
    dragScale: (x: number, y: number, refused: Refused) => call((b) => b.DragScale(x, y), refused),
    endScale: (x: number, y: number, refused: Refused) => call((b) => b.EndScale(x, y), refused),
    setScale: (percent: number, refused: Refused) => call((b) => b.SetScale(percent), refused),
    setPixelRatio: (ratio: number, refused: Refused) => call((b) => b.SetPixelRatio(ratio), refused),
    setBackground: (red: number, green: number, blue: number, refused: Refused) =>
      call((b) => b.SetBackground(red, green, blue), refused),
    ribbonDrawn: (refused: Refused) => call((b) => b.RibbonDrawn(), refused),
    togglePullOut: (refused: Refused) => call((b) => b.TogglePullOut(), refused),
    showContextMenu: (refused: Refused) => call((b) => b.ShowContextMenu(), refused),
    choose: (action: string, refused: Refused) => call((b) => b.Choose(action), refused),
    openPanel: (panel: string, refused: Refused) => call((b) => b.OpenPanel(panel), refused),
    fitPanel: (height: number, refused: Refused) => call((b) => b.FitPanel(height), refused),
    closePanel: (refused: Refused) => call((b) => b.ClosePanel(), refused),
    hide: (refused: Refused) => call((b) => b.Hide(), refused),
    openDonation: (refused: Refused) => call((b) => b.OpenDonation(), refused),
    openUpdate: (refused: Refused) => call((b) => b.OpenUpdate(), refused),
    skipUpdate: (refused: Refused) => call((b) => b.SkipUpdate(), refused),
    about: (refused: Refused) => call((b) => b.About(), refused),
    licence: (refused: Refused) => call((b) => b.Licence(), refused),
  }
}

/** WindowCalls is the window's calls, as windowCalls answers them. */
export type WindowCalls = ReturnType<typeof windowCalls>

/** on listens for a Go event, answering the call that stops listening. Outside Wails it hears nothing. */
export function on(name: string, callback: (...data: unknown[]) => void): () => void {
  return window.runtime?.EventsOn(name, callback) ?? (() => undefined)
}

/** startDrag hands the press to the system's own window drag, as Wails' drag regions do (FR-401). */
export function startDrag(): void {
  window.WailsInvoke?.('drag')
}
