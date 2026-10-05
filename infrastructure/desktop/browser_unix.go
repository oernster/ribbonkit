//go:build linux || darwin

package desktop

import (
	"fmt"
	"os/exec"
)

// OpenInBrowser hands address to the program the desktop opens such addresses with, answering an
// error when it cannot: no opener installed, nothing to open the address with. The application
// never fetches the address itself. Each platform names its opener in a file of its own.
func OpenInBrowser(address string) error { return openWith(opener, address) }

// openWith runs program on address, naming the address in any refusal.
func openWith(program, address string) error {
	if err := exec.Command(program, address).Run(); err != nil {
		return fmt.Errorf("opening %s: %w", address, err)
	}
	return nil
}
