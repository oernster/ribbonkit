import { describe, expect, it } from 'vitest'
import kit from '../../package.json'
import { FRAME, delayOf, normalise, scanPage, timerSites, type AllowedSite } from './timers'

/** quoted writes a module name as an import states it. The page below is data: written out whole, its
 * import of the kit by name would read to the kit's own structural scan as this file importing it. */
const quoted = (name: string) => `'${name}'`

/** A page of the shape an application's is, the kit's half under kit/ as an application's glob keys it. */
function page(extra: Record<string, string> = {}): Record<string, string> {
  return {
    'main.tsx': `import { Band, Help } from ${quoted(kit.name)}\nwindow.setTimeout(load, snapshot.refreshInMs)\n`,
    'kit/web/index.ts': "export { Band } from './Band'\nexport { Help } from './Help'\n",
    'kit/web/Band.tsx': 'let frame = window.requestAnimationFrame(() => {\n  frame = window.requestAnimationFrame(() => drawn())\n})\n',
    'kit/web/Help.tsx': "import { useAutoScroll } from './autoScroll'\n",
    'kit/web/autoScroll.ts':
      "import autoScroll from '../installer/page/auto-scroll.js'\nuseEffect(() => (node == null ? undefined : autoScroll.attach(node)))\n",
    'kit/installer/page/auto-scroll.js': 'const timer = window.setInterval(() => step(), TICK_MS)\nclearInterval(timer)\n',
    ...extra,
  }
}

const own: AllowedSite[] = [{ file: 'main.tsx', api: 'setTimeout', delay: 'snapshot.refreshInMs', reason: 'the next minute' }]

describe('the page timer scan', () => {
  it('finds nothing wrong on a page that keeps the rule', () => {
    const found = scanPage({ sources: page(), kitRoot: 'kit', testSupport: {}, allowed: own })
    expect(found.unreached).toEqual([])
    expect(found.supportReached).toEqual([])
    expect(found.periodic).toEqual([])
    expect(found.helpStarters).toEqual(['kit/web/Help.tsx'])
    expect(found.found).toEqual(found.allowed)
    expect([found.cycleBalanced, found.hookUnmounts, found.reachesHelpCycle, found.reachesKit]).toEqual([true, true, true, true])
  })

  it('reports an interval outside the Help cycle and a timer not on the list', () => {
    const sources = page({ 'main.tsx': `${page()['main.tsx']}window.setInterval(tick, 1000)\n` })
    const found = scanPage({ sources, kitRoot: 'kit', testSupport: {}, allowed: own })
    expect(found.periodic).toEqual(['main.tsx setInterval(1000)'])
    expect(found.found).not.toEqual(found.allowed)
  })

  it('reports source the page never imports unless it is named as test support', () => {
    const sources = page({ 'Orphan.tsx': 'export const x = 1\n', 'fakeBridge.ts': '' })
    const found = scanPage({ sources, kitRoot: 'kit', testSupport: { 'fakeBridge.ts': 'stands in for Go' }, allowed: own })
    expect(found.unreached).toEqual(['Orphan.tsx'])
  })

  it('refuses an import it was not given, rather than passing over it', () => {
    const sources = page({ 'main.tsx': "import { gone } from './Gone'\n" })
    expect(() => scanPage({ sources, kitRoot: 'kit', testSupport: {}, allowed: own })).toThrow(/widen the glob/)
  })

  it('reads each call with the delay it is written with', () => {
    expect(timerSites('a.ts', 'setTimeout(() => go(1, 2), later)\nrequestAnimationFrame(draw)')).toEqual([
      { file: 'a.ts', api: 'setTimeout', delay: 'later' },
      { file: 'a.ts', api: 'requestAnimationFrame', delay: FRAME },
    ])
    expect(delayOf('f(a, [b, c])', 1)).toBe('[b, c]')
    expect(normalise('../x/./y/../z')).toBe('../x/z')
  })
})
