package desktop

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// openVerb is the shell's verb for opening an address with whatever Windows has chosen for it.
const openVerb = "open"

// OpenInBrowser hands address to the program Windows opens such addresses with, answering an error
// when Windows cannot: no default browser, no program at all for the address. The application
// never fetches the address itself.
func OpenInBrowser(address string) error {
	verb, err := windows.UTF16PtrFromString(openVerb)
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(address)
	if err != nil {
		return fmt.Errorf("opening %q: %w", address, err)
	}
	if err := windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return fmt.Errorf("opening %s: %w", address, err)
	}
	return nil
}
