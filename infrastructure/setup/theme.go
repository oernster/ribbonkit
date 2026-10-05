//go:build windows

package setup

import (
	"errors"
	"regexp"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

// Ground is one surface colour as red, green and blue.
type Ground struct{ R, G, B uint8 }

// surfaceToken finds each --surface declaration in the setup page's sheet: light first, dark second.
var surfaceToken = regexp.MustCompile(`--surface:\s*#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})\s*;`)

// themesStated is how many --surface declarations the sheet holds: one for light, one for dark.
const themesStated = 2

// ErrNoSurfaces is answered when the sheet does not state a light and a dark surface.
var ErrNoSurfaces = errors.New("the setup sheet does not state a light and a dark --surface")

// Surfaces reads the light and the dark surface out of the setup page's sheet, so the window can be
// painted its ground before the page loads without holding a second copy of either colour.
func Surfaces(sheet string) (light, dark Ground, err error) {
	found := surfaceToken.FindAllStringSubmatch(sheet, -1)
	if len(found) != themesStated {
		return Ground{}, Ground{}, ErrNoSurfaces
	}
	return ground(found[0]), ground(found[1]), nil
}

// A colour channel is written as two hex digits holding eight bits.
const (
	hexBase     = 16
	channelBits = 8
)

// ground turns one match's three hex pairs into a colour; the pattern admits only hex digits.
func ground(match []string) Ground {
	channel := func(hex string) uint8 {
		value, _ := strconv.ParseUint(hex, hexBase, channelBits)
		return uint8(value)
	}
	return Ground{R: channel(match[1]), G: channel(match[2]), B: channel(match[3])}
}

// The Windows app theme, read so setup opens in the appearance the user has chosen.
const (
	themeKeyPath   = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	themeValueName = "AppsUseLightTheme"
	// darkThemeValue is what AppsUseLightTheme holds while apps are dark.
	darkThemeValue = 0
)

// SystemPrefersDark reports whether Windows is set to a dark app theme. A missing or unreadable
// value reads as light, as a fresh Windows install is.
func SystemPrefersDark() bool { return prefersDarkAt(themeKeyPath) }

// prefersDarkAt reads the theme value under key, so a test can point it at a scratch key.
func prefersDarkAt(key string) bool {
	opened, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer opened.Close()
	value, _, err := opened.GetIntegerValue(themeValueName)
	return err == nil && value == darkThemeValue
}
