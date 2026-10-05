import { useEffect, useState, type KeyboardEvent, type ReactNode } from 'react'
import { useAutoScroll } from './autoScroll'
import type { WindowCalls } from './bridge'
import type { AboutFacts, UpdateStatus } from './wire'

interface PanelProps {
  title: string
  problem: string
  onClose: () => void
  children: ReactNode
}

/**
 * Panel is the frame About and Licence share (FR-607, FR-608): the heading and Close stay put above
 * a body that reads itself when it holds more than fits (FR-609). The panel opens on Close rather
 * than in the body, so opening is never taken for a reader; Escape closes it too.
 */
function Panel({ title, problem, onClose, children }: PanelProps) {
  const reading = useAutoScroll()
  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }
  return (
    <main className="panel" onKeyDown={escape}>
      <header>
        <h1>{title}</h1>
        <button type="button" autoFocus onClick={onClose}>
          Close
        </button>
      </header>
      {problem !== '' && (
        <p className="problem" role="alert">
          {problem}
        </p>
      )}
      <div className="panel-body" ref={reading} tabIndex={0}>
        {children}
      </div>
    </main>
  )
}

interface AboutProps {
  onClose: () => void
  calls: Pick<WindowCalls, 'about'>
  /** icon is the application's own picture, shown at the head of the panel. */
  icon: string
}

/** About names the application, its author and every component it ships, in that order (FR-607). */
export function About({ onClose, calls, icon }: AboutProps) {
  const [facts, setFacts] = useState<AboutFacts | null>(null)
  const [problem, setProblem] = useState('')
  useEffect(() => {
    void calls.about(setProblem).then(setFacts)
  }, [calls])
  return (
    <Panel title="About" problem={problem} onClose={onClose}>
      <div className="about-head">
        <img src={icon} alt="" draggable={false} />
        {facts != null && (
          <>
            <h2>
              {facts.name} {facts.version}
            </h2>
            <p>by {facts.author}</p>
            <p className="muted">{facts.copyright}</p>
          </>
        )}
      </div>
      {facts != null && (
        <>
          <h2>Credits</h2>
          <ul className="credits">
            {facts.credits.map((credit) => (
              <li key={credit.name}>
                <span className="credit-name">{credit.name}</span>, {credit.licence}: {credit.role}
              </li>
            ))}
          </ul>
        </>
      )}
    </Panel>
  )
}

/**
 * Update is an update check's outcome (FR-509). A newer release offers Download, Skip this version
 * and Later; a check that found none says so. One that could not reach GitHub says that instead. Go keeps the
 * addresses and the version: Download and Skip ask it to act on what it offered.
 */
export function Update({
  status,
  onClose,
  calls,
}: {
  status: UpdateStatus
  onClose: () => void
  calls: Pick<WindowCalls, 'openUpdate' | 'skipUpdate'>
}) {
  const [problem, setProblem] = useState('')
  const thenClose = (done: unknown) => {
    if (done !== null) {
      onClose()
    }
  }
  if (!status.updateAvailable) {
    const outcome =
      status.latest !== ''
        ? 'You are running the latest version.'
        : 'The update check could not reach GitHub. Please try again later.'
    return (
      <Panel title="Check for updates" problem={problem} onClose={onClose}>
        <p>{outcome}</p>
      </Panel>
    )
  }
  return (
    <Panel title="Update available" problem={problem} onClose={onClose}>
      <p>
        Version {status.latest} is available. You are running {status.current}.
      </p>
      <div className="update-actions">
        <button type="button" onClick={() => void calls.openUpdate(setProblem).then(thenClose)}>
          Download
        </button>
        <button type="button" onClick={() => void calls.skipUpdate(setProblem).then(thenClose)}>
          Skip this version
        </button>
        <button type="button" onClick={onClose}>
          Later
        </button>
      </div>
    </Panel>
  )
}

/** Licence shows the whole of the terms the application was built with (FR-608). */
export function Licence({ onClose, calls }: { onClose: () => void; calls: Pick<WindowCalls, 'licence'> }) {
  const [text, setText] = useState('')
  const [problem, setProblem] = useState('')
  useEffect(() => {
    void calls.licence(setProblem).then((terms) => setText(terms ?? ''))
  }, [calls])
  return (
    <Panel title="Licence" problem={problem} onClose={onClose}>
      <div className="licence">
        <pre className="licence-text">{text}</pre>
      </div>
    </Panel>
  )
}
