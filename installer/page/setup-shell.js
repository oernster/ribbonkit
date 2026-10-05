const $ = (id) => document.getElementById(id)

// appName is the product's name, read from the setup program rather than written here, so a rename
// cannot leave this page announcing a product that no longer exists. The static text carries none
// of it; every screen fills it in from this.
let appName = ''

// currentState is the one reading of the machine, kept so a screen that goes back returns to the
// route it came from.
let currentState = null

// logPath is where the setup program writes its step log; a failure names it.
let logPath = ''

// lastView is the screen showing with its actions, so the Licence screen can hand it back.
let lastView = null

// stopReading ends the Licence screen reading itself (FR-811); null while it is not.
let stopReading = null

function backend() {
    return window.go && window.go.main && window.go.main.App
}

// quitMessage is what Wails' own dispatcher reads as quit on the web view's message channel
// (dispatcher.go in wails v2.12.0, case 'Q'); its runtime Quit sends exactly this.
const quitMessage = 'Q'

// closer answers how this page can close its window; null when it has no way to. The setup
// program's Quit comes first; without it the Wails runtime, then the web view's own channel, still
// reach the dispatcher that closes the window.
function closer() {
    if (backend()) return () => backend().Quit()
    if (window.runtime && window.runtime.Quit) return () => window.runtime.Quit()
    const webview = window.chrome && window.chrome.webview
    if (webview) return () => webview.postMessage(quitMessage)
    return null
}

function closeSetup() {
    const close = closer()
    if (close) close()
}

/* ------------------------------------------------------------------ theme */

// applyTheme repaints and re-faces the toggle in one step: the toggle shows the appearance it
// would switch TO, so the sun shows while the page is dark and the moon while it is light.
function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme)
    const to = theme === 'dark' ? 'light' : 'dark'
    $('theme-icon').src = to + '-mode.png'
    $('theme').title = 'Switch to ' + to
    $('theme').setAttribute('aria-label', 'Switch to ' + to)
}

function currentTheme() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

$('theme').onclick = () => applyTheme(currentTheme() === 'dark' ? 'light' : 'dark')

/* ---------------------------------------------------------------- views */

// showView shows one screen with the actions that belong to it. The footer is rebuilt from the
// list every time, never relabelled, so no button remembers what it used to mean. The progress
// screen is given no actions; the Licence button leaves the header while work runs.
function showView(name, actions) {
    if (stopReading !== null) {
        stopReading()
        stopReading = null
    }
    document.querySelectorAll('.screen').forEach((el) => el.classList.remove('active'))
    $('screen-' + name).classList.add('active')
    $('licence').hidden = name === 'progress'
    if (name !== 'licence') lastView = { name, actions }
    const footer = $('footer')
    footer.replaceChildren()
    actions.forEach((spec) => {
        const el = document.createElement('button')
        el.className = 'btn' + (spec.kind ? ' ' + spec.kind : '')
        el.textContent = spec.label
        el.onclick = spec.onClick
        footer.appendChild(el)
    })
    startNeutral()
}

// startNeutral leaves a screen opening with nothing focused (FR-809): no control wears a ring until
// the keyboard asks for one. The first Tab or Right enters the ring at its first stop, since the
// ring forgets where it stood on the last screen, whose footer has gone.
function startNeutral() {
    if (document.activeElement instanceof HTMLElement) document.activeElement.blur()
    forgetRingMark()
}

// keyboardSettleMs is how long the web view is given to come up holding the keyboard before the
// page decides it has not.
const keyboardSettleMs = 400

// settleKeyboard repairs a launch that came up with no keyboard at all, so the first Tab reaches the
// ring even though nothing starts focused. Only the page can tell whether the document holds it.
function settleKeyboard() {
    window.focus()
    window.setTimeout(() => {
        if (document.hasFocus()) return
        void backend().TakeKeyboard().then(() => window.focus())
    }, keyboardSettleMs)
}

// showLicence shows the licence setup carries, with Back returning to the screen it came from. The
// body it scrolls in reads it from the top, afresh each time the screen opens (FR-811); showView
// ends that the moment another screen shows.
async function showLicence() {
    const back = lastView
    try {
        $('licence-text').textContent = await backend().Licence()
    } catch (e) {
        $('licence-text').textContent = String(e)
    }
    showView('licence', [{ label: 'Back', kind: 'primary', onClick: () => showView(back.name, back.actions) }])
    const body = document.querySelector('.body')
    body.scrollTop = 0
    stopReading = window.AutoScroll.attach(body)
}

$('licence').onclick = showLicence

/* ---------------------------------------------------------------- options */

// renderOptions fills a container with boxes and answers a reader for their values, so no screen
// has to know the ids of its own boxes.
function renderOptions(container, specs) {
    container.replaceChildren()
    const boxes = {}
    specs.forEach((spec) => {
        const label = document.createElement('label')
        label.className = 'option'
        const input = document.createElement('input')
        input.type = 'checkbox'
        input.checked = !!spec.checked
        if (spec.onChange) input.onchange = () => spec.onChange(input.checked, input)
        const tick = document.createElement('span')
        tick.className = 'check'
        const text = document.createElement('span')
        const title = document.createElement('span')
        title.className = 'label'
        title.textContent = spec.label
        text.appendChild(title)
        if (spec.hint) {
            const hint = document.createElement('span')
            hint.className = 'hint'
            hint.textContent = spec.hint
            text.appendChild(hint)
        }
        label.append(input, tick, text)
        container.appendChild(label)
        boxes[spec.key] = input
    })
    return (key) => boxes[key].checked
}

// choiceOptions are the three boxes of FR-805, opening on what the reading says: the machine's own
// where the application is installed, the fresh defaults where it is not.
function choiceOptions(state, onChange) {
    return [
        { key: 'startMenu', label: 'Add to the Start Menu', checked: state.startMenu, onChange },
        { key: 'desktop', label: 'Add a Desktop shortcut', checked: state.desktop, onChange },
        { key: 'startWithWindows', label: 'Start with Windows', checked: state.startWithWindows, onChange },
    ]
}

// launchOption closes every screen that writes files: ticked, setup starts the application and
// closes once it is up, since that is why people run an installer (FR-805).
function launchOption() {
    return { key: 'launch', label: `Start ${appName} when setup closes`, checked: true }
}

// choicesFrom reads the three boxes into what the setup program takes.
function choicesFrom(read) {
    return { startMenu: read('startMenu'), desktop: read('desktop'), startWithWindows: read('startWithWindows') }
}

/* ------------------------------------------------------------------- work */

function onProgress(p) {
    $('progress-fill').style.width = p.pct + '%'
    $('progress-status').textContent = p.msg
}

// run moves to the progress screen, which offers nothing, does the work and ends in a verdict.
async function run(work, title, doneTitle, doneMsg) {
    $('progress-title').textContent = title
    onProgress({ pct: 0, msg: '' })
    showView('progress', [])
    try {
        await work()
    } catch (e) {
        showError(String(e))
        return
    }
    $('done-title').textContent = doneTitle
    $('done-msg').textContent = doneMsg
    showView('done', [{ label: 'Close', kind: 'primary', onClick: closeSetup }])
}

// writeFiles runs one write of the files, then honours the launch box: a launch that works brings
// the application forward and closes setup; one that fails leaves the reason on screen.
function writeFiles(work, launch, title, doneTitle, doneMsg) {
    return withAppClosed(() => run(
        () => work().then(() => {
            if (launch) return backend().LaunchApp().then(closeSetup)
        }),
        title, doneTitle, doneMsg + (launch ? ' It is starting now.' : ''),
    ))
}

// showError is the failure verdict (FR-808): the reason, where the step log is and Close, offered
// wherever the page has a way to close.
function showError(message) {
    $('error-msg').textContent = message
    $('error-log').textContent = logPath ? 'The steps taken are recorded in ' + logPath + '.' : ''
    showView('error', closer() ? [{ label: 'Close', kind: 'primary', onClick: closeSetup }] : [])
}

// withAppClosed runs the work once the application is not running. If it is, the offer to close it
// comes first, before any file is touched (FR-807).
async function withAppClosed(proceed) {
    if (!(await backend().AppRunning())) {
        proceed()
        return
    }
    showView('running', [
        { label: 'Cancel', onClick: () => route(currentState) },
        {
            label: 'Close it and continue', kind: 'primary', onClick: async () => {
                $('progress-title').textContent = `Closing ${appName}`
                onProgress({ pct: 0, msg: '' })
                showView('progress', [])
                try {
                    await backend().CloseRunningApp()
                } catch (e) {
                    showError(String(e))
                    return
                }
                proceed()
            },
        },
    ])
}
