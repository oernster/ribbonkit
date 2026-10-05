// The ribbon's page wakes no more often than once a minute while it is shown: every timer call any
// module of the page makes is on an allow-list with the delay it is written with; the one periodic
// timer is the Help panel's self-reading cycle, running only while Help is shown
// (TimeRibbon NFR-P-4 and FR-609, WeatherRibbon NFR-P-3).
//
// The page is every module the application's entry reaches through the source's own imports,
// ribbonkit's half included (reached through the package's exports) plus auto-scroll.js beside the
// setup page. Packages other than the kit and the runtime Wails injects are not read. An
// application hands in its sources, read with import.meta.glob in its own test file (the glob's
// patterns must be written there), its own test support and its own timer calls; the kit's are
// stated here, once.

import { describe, expect, it } from 'vitest'
import kit from '../../package.json'

const TIMER_APIS = ['setInterval', 'setTimeout', 'requestAnimationFrame'] as const

/** TimerApi is a call that schedules the page to wake. */
export type TimerApi = (typeof TIMER_APIS)[number]

/** FRAME stands for the delay of a requestAnimationFrame call, which takes none. */
export const FRAME = 'next frame'

/** TimerSite is one timer call: its module, its API and the delay it is written with. */
export interface TimerSite {
  file: string
  api: TimerApi
  delay: string
}

/** AllowedSite is a timer call allowed on the page, with why it cannot wake it more than once a minute. */
export interface AllowedSite extends TimerSite {
  reason: string
}

/** PageTimers is what an application hands in. Paths are as its glob keys them: relative to its test file. */
export interface PageTimers {
  /** sources are the application's own scripts and the kit's, as import.meta.glob read them raw. */
  sources: Record<string, string>
  /** kitRoot is the kit package's folder, as the glob keys it. */
  kitRoot: string
  /** testSupport is the application's own source that only its tests import, with why. */
  testSupport: Record<string, string>
  /** allowed are the application's own timer calls. */
  allowed: AllowedSite[]
  /** requirement names the rule in the application's own specification, for the suite's title. */
  requirement: string
  /** entry is the module index.html loads; main.tsx when not given. */
  entry?: string
}

/** Findings is what a scan of the page found, each list empty on a page that keeps the rule. */
export interface Findings {
  /** unreached is shipped source the page never imports. */
  unreached: string[]
  /** supportReached is test support the page imports. */
  supportReached: string[]
  /** periodic are setInterval calls outside the Help cycle. */
  periodic: string[]
  /** helpStarters are the page modules that start the Help cycle. */
  helpStarters: string[]
  /** found and allowed are every timer call on the page and every one allowed, sorted. */
  found: string[]
  allowed: string[]
  /** cycleBalanced says the Help cycle clears every interval it sets; hookUnmounts that its hook ends it with the panel. */
  cycleBalanced: boolean
  hookUnmounts: boolean
  /** reachesHelpCycle and reachesKit say the walk reached the cycle and the kit's half at all. */
  reachesHelpCycle: boolean
  reachesKit: boolean
}

/** The forms an import may resolve to; any other import is an asset. */
const SCRIPT_EXTENSIONS = ['.ts', '.tsx', '.js']

const IMPORT_FORMS = [
  /(?:import|export)\s[^'"]*?\sfrom\s*['"]([^'"]+)['"]/g,
  /import\s*['"]([^'"]+)['"]/g,
  /import\(\s*['"]([^'"]+)['"]\s*\)/g,
]

/** kitPaths answers where the kit's own modules lie under kitRoot. */
function kitPaths(kitRoot: string) {
  const source = `${kitRoot}/web/`
  return {
    source,
    helpCycle: `${kitRoot}/installer/page/auto-scroll.js`,
    helpCycleHook: `${source}autoScroll.ts`,
    helpCycleUsers: [`${source}Help.tsx`],
    testSupport: {
      [`${source}testing/index.ts`]: "the kit's stand-in bridge for the suites",
      [`${source}testing/setup.ts`]: "the kit's set-up for every suite",
      [`${source}testing/timers.ts`]: 'this scan',
      [`${source}setup/setupPage.ts`]: "lays out the setup page for the kit's own suites",
    } as Record<string, string>,
    allowed: [
      { file: `${kitRoot}/installer/page/auto-scroll.js`, api: 'setInterval', delay: 'TICK_MS',
        reason: "the Help panel's self-reading cycle; attach starts it for a mounted panel and its cleanup clears it" },
      { file: `${source}Band.tsx`, api: 'requestAnimationFrame', delay: FRAME,
        reason: 'first of two frames telling Go an opened ribbon has been painted; once per opening' },
      { file: `${source}Band.tsx`, api: 'requestAnimationFrame', delay: FRAME,
        reason: 'second of the two frames, scheduled once from the first; it re-arms nothing' },
    ] as AllowedSite[],
  }
}

/** normalise folds '.' and '..' out of a path, keeping leading '..'. */
export function normalise(path: string): string {
  const parts: string[] = []
  for (const part of path.split('/')) {
    if (part === '' || part === '.') {
      continue
    }
    if (part === '..' && parts.length > 0 && parts[parts.length - 1] !== '..') {
      parts.pop()
    } else {
      parts.push(part)
    }
  }
  return parts.join('/')
}

/** delayOf answers the last top-level argument of the call whose '(' is at open. */
export function delayOf(text: string, open: number): string {
  let depth = 0
  let start = open + 1
  for (let at = open; at < text.length; at++) {
    const char = text[at]
    if ('([{'.includes(char)) {
      depth++
    } else if (')]}'.includes(char)) {
      depth--
      if (depth === 0) {
        return text.slice(start, at).trim()
      }
    } else if (char === ',' && depth === 1) {
      start = at + 1
    }
  }
  return ''
}

/** timerSites answers every timer call in text, as file's, with the delay it is written with. */
export function timerSites(file: string, text: string): TimerSite[] {
  const sites: TimerSite[] = []
  for (const match of text.matchAll(new RegExp(`\\b(${TIMER_APIS.join('|')})\\s*\\(`, 'g'))) {
    const api = match[1] as TimerApi
    const open = (match.index ?? 0) + match[0].length - 1
    sites.push({ file, api, delay: api === 'requestAnimationFrame' ? FRAME : delayOf(text, open) })
  }
  return sites
}

function key(site: TimerSite): string {
  return `${site.file} ${site.api}(${site.delay})`
}

function count(text: string, pattern: RegExp): number {
  return [...text.matchAll(pattern)].length
}

/** scanPage walks the page from its entry and answers what it found. */
export function scanPage(page: Omit<PageTimers, 'requirement'>): Findings {
  const paths = kitPaths(page.kitRoot)
  const modules = new Map(Object.entries(page.sources).map(([file, text]) => [normalise(file), text]))
  const support = { ...page.testSupport, ...paths.testSupport }

  const kitExport = (specifier: string): string | null => {
    if (specifier !== kit.name && !specifier.startsWith(`${kit.name}/`)) {
      return null
    }
    const target = (kit.exports as Record<string, string>)[`.${specifier.slice(kit.name.length)}`]
    if (target === undefined) {
      throw new Error(`${specifier} is not among ribbonkit's exports`)
    }
    return normalise(`${page.kitRoot}/${target}`)
  }

  const importsOf = (file: string): string[] => {
    const text = modules.get(file) ?? ''
    const folder = file.includes('/') ? file.slice(0, file.lastIndexOf('/')) : ''
    const found: string[] = []
    for (const form of IMPORT_FORMS) {
      for (const match of text.matchAll(form)) {
        const specifier = match[1]
        const exported = kitExport(specifier)
        if (exported == null && !specifier.startsWith('.')) {
          continue
        }
        const bare = exported ?? normalise(`${folder}/${specifier}`)
        const dot = bare.lastIndexOf('.')
        const written = dot > bare.lastIndexOf('/') && !SCRIPT_EXTENSIONS.includes(bare.slice(dot))
        if (written || specifier.includes('?')) {
          continue
        }
        const resolved = ['', ...SCRIPT_EXTENSIONS].map((tail) => bare + tail).find((path) => modules.has(path))
        if (resolved == null) {
          throw new Error(`${file} imports ${specifier}, which the scan was not given; widen the glob`)
        }
        found.push(resolved)
      }
    }
    return found
  }

  const reached = new Set<string>()
  const waiting = [page.entry ?? 'main.tsx']
  while (waiting.length > 0) {
    const file = waiting.pop() as string
    if (!reached.has(file)) {
      reached.add(file)
      waiting.push(...importsOf(file))
    }
  }

  const ours = (file: string) => !file.startsWith('..') || file.startsWith(paths.source)
  const isTest = (file: string) => /\.test\.tsx?$/.test(file)
  const sites = [...reached].flatMap((file) => timerSites(file, modules.get(file) ?? ''))
  const cycle = modules.get(paths.helpCycle) ?? ''
  return {
    unreached: [...modules.keys()].filter((file) => ours(file) && !isTest(file) && !(file in support) && !reached.has(file)),
    supportReached: Object.keys(support).filter((file) => reached.has(file)),
    periodic: sites.filter((site) => site.api === 'setInterval' && site.file !== paths.helpCycle).map(key),
    helpStarters: [...reached].filter((file) => importsOf(file).includes(paths.helpCycleHook)).sort(),
    found: sites.map(key).sort(),
    allowed: [...paths.allowed, ...page.allowed].map(key).sort(),
    cycleBalanced: count(cycle, /\bclearInterval\s*\(/g) === count(cycle, /\bsetInterval\s*\(/g),
    hookUnmounts: /useEffect\(\(\) => \(node == null \? undefined : autoScroll\.attach\(node\)\)/.test(modules.get(paths.helpCycleHook) ?? ''),
    reachesHelpCycle: reached.has(paths.helpCycle),
    reachesKit: [...reached].some((file) => file.startsWith(paths.source)),
  }
}

/** describePageTimers holds the application's page to the rule, as one suite. */
export function describePageTimers(page: PageTimers) {
  const users = kitPaths(page.kitRoot).helpCycleUsers
  describe(`the ribbon page's timers (${page.requirement})`, () => {
    const findings = scanPage(page)

    it('reads every shipped source file and no test support, so nothing escapes the scan', () => {
      expect(findings.unreached, 'source the page never imports; list it as test support or remove it').toEqual([])
      expect(findings.supportReached, 'test support the page imports').toEqual([])
      expect(findings.reachesHelpCycle, 'the Help cycle').toBe(true)
      expect(findings.reachesKit, "ribbonkit's half of the page").toBe(true)
    })

    it('schedules a periodic timer only in the Help self-reading cycle', () => {
      expect(findings.periodic, 'setInterval outside the Help cycle wakes the shown ribbon').toEqual([])
    })

    it('starts the self-reading cycle only from Help, which clears it as the panel goes', () => {
      expect(findings.helpStarters).toEqual(users)
      expect(findings.cycleBalanced).toBe(true)
      expect(findings.hookUnmounts).toBe(true)
    })

    it('has every timer call on the allow-list with the delay it is written with', () => {
      expect(findings.found, 'a timer call not on the allow-list: allow it with why it cannot wake the page more than once a minute').toEqual(findings.allowed)
    })
  })
}
