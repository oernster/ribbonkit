/* ----------------------------------------------------------------- routes */

// Each route is one screen with its own actions, chosen once from the setup program's reading of
// the machine (FR-801). Uninstall is a screen reachable from every route with something installed;
// the route itself never becomes it.

function routeInstall(state) {
    $('install-title').textContent = `Install ${appName}`
    $('install-version').textContent = `This installs version ${state.thisVersion}.`
    const read = renderOptions($('install-options'), choiceOptions(state).concat([launchOption()]))
    showView('install', [
        { label: 'Cancel', onClick: closeSetup },
        {
            label: 'Install', kind: 'primary',
            onClick: () => writeFiles(() => backend().Install(choicesFrom(read)), read('launch'),
                `Installing ${appName}`, `${appName} is installed`, `Version ${state.thisVersion} is on this machine.`),
        },
    ])
}

// routeChange serves both directions of a version change, which differ only in wording. The two
// versions stand in a flow line rather than in the heading; the change itself leads.
function routeChange(state) {
    const back = state.route === 'downgrade'
    $('update-title').textContent = back ? 'Go back a version?' : `Update ${appName}`
    $('update-lead').textContent = back
        ? 'This setup file carries an older version than the one installed. Your clocks and choices are kept.'
        : 'A newer version is ready. Your clocks and choices are kept.'
    $('update-from').textContent = 'v' + state.installedVersion
    $('update-to').textContent = 'v' + state.thisVersion
    const read = renderOptions($('update-options'), choiceOptions(state).concat([launchOption()]))
    showView('update', [
        { label: 'Uninstall', kind: 'danger', onClick: () => routeUninstall(state) },
        { label: 'Close', onClick: closeSetup },
        {
            label: back ? 'Go back' : 'Update', kind: 'primary',
            onClick: () => writeFiles(() => backend().Install(choicesFrom(read)), read('launch'),
                back ? 'Going back a version' : `Updating ${appName}`,
                back ? 'Version changed' : `${appName} is updated`, `Version ${state.thisVersion} is on this machine.`),
        },
    ])
}

// routeManage is the screen for a matching version. There is nothing to install, so each box takes
// effect the moment it changes; one that fails is put back and the reason shown.
function routeManage(state) {
    $('manage-title').textContent = `${appName} is installed`
    $('manage-version').textContent = `Version ${state.installedVersion} is installed; this setup file carries the same version.`
    const apply = (on, box) => backend().Apply(choicesFrom(read)).catch((e) => {
        box.checked = !on
        showError(String(e))
    })
    const read = renderOptions($('manage-options'), choiceOptions(state, apply).concat([launchOption()]))
    showView('manage', [
        { label: 'Uninstall', kind: 'danger', onClick: () => routeUninstall(state) },
        { label: 'Close', onClick: closeSetup },
        {
            label: 'Reinstall',
            onClick: () => writeFiles(() => backend().Install(choicesFrom(read)), read('launch'),
                `Reinstalling ${appName}`, `${appName} is reinstalled`,
                'The files were written again with the boxes applied as they stood.'),
        },
        {
            label: 'Repair', kind: 'primary',
            onClick: () => writeFiles(() => backend().Repair(), read('launch'),
                `Repairing ${appName}`, 'Repair complete', 'The files were put back and nothing else was changed.'),
        },
    ])
}

function routeUninstall(state) {
    $('uninstall-title').textContent = `Remove ${appName}?`
    const read = renderOptions($('uninstall-options'), [{
        key: 'forget', label: 'Also forget my settings',
        hint: 'Deletes your clocks and choices. It cannot be undone.',
        checked: false,
    }])
    // Cancel returns to the route setup opened on. Opened from the Apps list to uninstall, there is
    // nothing to return to, so setup closes and leaves the reader on the list they came from.
    showView('uninstall', [
        { label: 'Cancel', onClick: () => state.uninstall ? closeSetup() : route(state) },
        {
            label: 'Uninstall', kind: 'danger',
            onClick: () => {
                const forget = read('forget')
                return withAppClosed(() => run(() => backend().Uninstall(forget),
                    `Removing ${appName}`, `${appName} is removed`,
                    forget ? 'The application and your settings are gone.'
                        : 'The application is gone. Your clocks and choices are kept for another time.'))
            },
        },
    ])
}

function route(state) {
    currentState = state
    appName = state.appName
    logPath = state.logPath
    if (state.uninstall) {
        routeUninstall(state)
    } else if (state.route === 'install') {
        routeInstall(state)
    } else if (state.route === 'manage') {
        routeManage(state)
    } else {
        routeChange(state)
    }
}

// backendWaitMs and backendTries bound the wait for the setup program's bound methods to appear.
const backendWaitMs = 50
const backendTries = 100

async function init() {
    applyTheme('light')
    let tries = 0
    while (!backend() && tries < backendTries) {
        await new Promise((resolve) => setTimeout(resolve, backendWaitMs))
        tries++
    }
    if (!backend()) {
        showError('Could not reach the setup program.')
        return
    }
    window.runtime.EventsOn('progress', onProgress)
    const state = await backend().DetectState()
    appName = state.appName
    logPath = state.logPath
    document.title = `${appName} Setup`
    $('app-name').textContent = appName
    $('running-title').textContent = `${appName} is open`
    applyTheme(state.prefersDark ? 'dark' : 'light')
    if (state.problem) {
        showError(state.problem)
    } else {
        route(state)
    }
    settleKeyboard()
}

window.addEventListener('DOMContentLoaded', init)
