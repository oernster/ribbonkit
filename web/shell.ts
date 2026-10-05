import { useCallback, useEffect, useRef, useState } from 'react'
import { backgroundReporter } from './background'
import { on, type Refused, type WindowCalls } from './bridge'
import { percentOfWhole, showOpacity } from './opacity'
import { watchPixelRatio } from './pixelRatio'
import { scrollbarThickness } from './scrollbar'
import type { UpdateStatus } from './wire'

/** The panels the window can become (CON-6); the window names each in its open-panel event. */
export type Panel = 'settings' | 'about' | 'licence' | 'update'
export type View = 'ribbon' | Panel

/**
 * The open-panel event's words the window sends (ribbonkit/ui/window/facade.go), each naming the
 * panel it opens; update carries the check's outcome with it. The keys are quoted so the structural
 * test can find each word the window sends.
 */
const panelFor: Record<string, Panel> = {
  'settings': 'settings', 'about': 'about', 'licence': 'licence', 'update': 'update',
}

/** What the shell reads of the application's snapshot: how the window is drawn. */
export interface Drawn {
  theme: string
  colour: string
  opacity: number
}

/** What the shell needs of the application. Each must keep its identity from render to render. */
export interface ShellOptions<S extends Drawn> {
  calls: Pick<WindowCalls, 'openPanel' | 'closePanel' | 'setScrollbar' | 'setPixelRatio' | 'setBackground'>
  /** take answers the application's snapshot; null where Go refused, having told refused why. */
  take: (refused: Refused) => Promise<S | null>
  /** opens names the application's own open-panel words, each with the panel it opens. */
  opens: Record<string, Panel>
}

/**
 * useShell holds the application's snapshot and which view the window shows: the ribbon, else the
 * panel Go opened it at. The snapshot is taken again whenever Go says something changed (FR-209). It
 * tells Go what only the page can measure: the scroll bar, the pixel ratio and the background. It
 * draws the window in the theme, colour scheme and opacity the snapshot names (FR-606, FR-611,
 * FR-622).
 */
export function useShell<S extends Drawn>({ calls, take, opens }: ShellOptions<S>) {
  const [snapshot, setSnapshot] = useState<S | null>(null)
  const [problem, setProblem] = useState('')
  const [view, setView] = useState<View>('ribbon')
  // at is the word the panel was opened at; update is the check's outcome the update panel shows.
  const [at, setAt] = useState('')
  const [update, setUpdate] = useState<UpdateStatus | null>(null)
  // opened is true once Go has made the window the panel shown, so a panel measures its content at
  // its own size rather than the ribbon's (FR-621).
  const [opened, setOpened] = useState(false)

  const load = useCallback(() => {
    void take(setProblem).then((next) => {
      if (next != null) {
        setSnapshot(next)
        setProblem('')
      }
    })
  }, [take])

  const openPanel = useCallback(
    (word?: unknown, outcome?: unknown) => {
      const said = String(word ?? '')
      setAt(said)
      setUpdate((outcome as UpdateStatus | undefined) ?? null)
      const panel = panelFor[said] ?? opens[said] ?? 'settings'
      setOpened(false)
      setView(panel)
      void calls.openPanel(panel, setProblem).then(() => setOpened(true))
    },
    [calls, opens],
  )

  const closePanel = useCallback(() => {
    setView('ribbon')
    void calls.closePanel(setProblem).then(load)
  }, [calls, load])

  useEffect(() => {
    // Go makes room for the scroll bar a scrolling ribbon shows, which only the page can measure.
    void calls.setScrollbar(scrollbarThickness(), setProblem)
  }, [calls])

  // Go sizes the window by the scale the page is really drawn at, which only the page knows.
  useEffect(() => watchPixelRatio((ratio) => void calls.setPixelRatio(ratio, setProblem)), [calls])

  useEffect(() => {
    load()
    const stopRefresh = on('refresh', load)
    const stopPanel = on('open-panel', openPanel)
    return () => {
      stopRefresh()
      stopPanel()
    }
  }, [load, openPanel])

  // Go paints the window in the page's own background, which only the page's CSS knows.
  const background = useRef<ReturnType<typeof backgroundReporter> | null>(null)
  useEffect(() => {
    const reporter = backgroundReporter((red, green, blue) => void calls.setBackground(red, green, blue, setProblem))
    background.current = reporter
    reporter.check()
    return reporter.stop
  }, [calls])

  useEffect(() => {
    const root = document.documentElement
    if (snapshot == null || snapshot.theme === 'system') {
      delete root.dataset.theme
    } else {
      root.dataset.theme = snapshot.theme
    }
    // The colour scheme (FR-611); the palette's schemes key off it.
    root.dataset.colour = snapshot?.colour ?? 'classic'
    // The chosen opacity is the ribbon's; a panel the window becomes is always drawn opaque (FR-622).
    if (snapshot != null) {
      showOpacity(view === 'ribbon' ? snapshot.opacity : percentOfWhole)
    }
    background.current?.check()
  }, [snapshot, view])

  return { snapshot, problem, refused: setProblem, view, at, update, opened, load, openPanel, closePanel }
}
