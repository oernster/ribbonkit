package startup

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/oernster/ribbonkit/domain/identity"
)

// On macOS the entry is a launchd agent in the user's LaunchAgents folder, which launchd reads at
// every sign-in. Its label is the app id. LimitLoadToSessionType keeps it to a desktop sign-in, so
// a remote shell signing in never starts the ribbon.

const (
	agentsFolder = "Library/LaunchAgents"
	entrySuffix  = ".plist"
)

// agentTemplate is the agent written: its label, then the program to run.
const agentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>LimitLoadToSessionType</key>
	<string>Aqua</string>
</dict>
</plist>
`

// disabledKey matches the key launchd reads as switched off, whoever wrote it.
var disabledKey = regexp.MustCompile(`<key>Disabled</key>\s*<true/>`)

// New answers the agent for program, the full path of the application's executable, in the user's
// LaunchAgents folder.
func New(app identity.App, program string) Entry {
	dir, err := agentsDir(os.UserHomeDir)
	entry := At(app, dir, program)
	entry.problem = err
	return entry
}

// At answers the agent for program in dir. Only a test names a folder of its own, so the real
// user's agents are never touched.
func At(app identity.App, dir, program string) Entry {
	return Entry{app: app, dir: dir, command: program}
}

// entryText answers the agent that runs command, the path escaped for the property list.
func entryText(app identity.App, command string) string {
	return fmt.Sprintf(agentTemplate, escaped(app.AppID), escaped(command))
}

// switchedOff answers whether an agent's text carries the key launchd reads as off.
func switchedOff(text string) bool { return disabledKey.MatchString(text) }

// agentsDir answers the folder launchd reads the user's agents from.
func agentsDir(home func() (string, error)) (string, error) {
	dir, err := home()
	if err != nil || dir == "" {
		return "", errNoHome
	}
	return filepath.Join(dir, agentsFolder), nil
}

// escaped answers text with the characters XML reserves escaped.
func escaped(text string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(text))
	return out.String()
}
