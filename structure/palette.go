package structure

// The page draws its colours from palette halves: each states Classic in a :root block (light),
// under [data-theme='dark'] and under the system's dark, then every other scheme in a
// :root[data-colour='name'] block whose tokens are written light-dark(light, dark). A scheme that
// leaves a token out draws Classic's value for it. Nothing ties the blocks to the schemes the menus
// offer, so these checks do (TimeRibbon FR-611, NFR-U-1).

import (
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strconv"
	"testing"
)

// Half is one half of the palette: Classic names the file stating Classic, Schemes the file stating
// every other scheme (the same file is fine).
type Half struct {
	Classic string
	Schemes string
}

// Theme is the light or the dark side of a scheme.
type Theme string

// The two sides every scheme has.
const (
	Light Theme = "light"
	Dark  Theme = "dark"
)

var (
	schemeBlock  = regexp.MustCompile(`(?m)^:root\[data-colour='(\w+)'\] \{([^}]*)\}`)
	classicBlock = regexp.MustCompile(`(?m)^:root \{([^}]*)\}`)
	darkBlock    = regexp.MustCompile(`(?m)^:root\[data-theme='dark'\] \{([^}]*)\}`)
	systemDark   = regexp.MustCompile(`@media \(prefers-color-scheme: dark\) \{\s*:root:not\(\[data-theme='light'\]\) \{([^}]*)\}`)
	declaration  = regexp.MustCompile(`--([\w-]+):\s*([^;]+);`)
	lightDark    = regexp.MustCompile(`^light-dark\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)$`)
	shortHex     = regexp.MustCompile(`^#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])$`)
	longHex      = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`)
)

// declarations answers each token a block's body states, with its value.
func declarations(body string) map[string]string {
	tokens := map[string]string{}
	for _, match := range declaration.FindAllStringSubmatch(body, -1) {
		tokens[match[1]] = match[2]
	}
	return tokens
}

// CheckEveryOfferedSchemeHasItsOwnCompleteBlock fails for each scheme in offered (classic aside)
// that half has no block for, each of Classic's tokens a block leaves to Classic unless optional
// names it; also for each block no menu offers.
func CheckEveryOfferedSchemeHasItsOwnCompleteBlock(t testing.TB, half Half, classic string, offered, optional []string) {
	t.Helper()
	found := classicBlock.FindStringSubmatch(Read(t, half.Classic))
	if found == nil {
		t.Fatalf("%s has no :root block", half.Classic)
	}
	blocks := map[string]map[string]string{}
	for _, match := range schemeBlock.FindAllStringSubmatch(Read(t, half.Schemes), -1) {
		blocks[match[1]] = declarations(match[2])
	}
	for _, scheme := range offered {
		if scheme == classic {
			continue
		}
		stated, ok := blocks[scheme]
		if !ok {
			t.Errorf("%s is offered but %s has no block for it", scheme, half.Schemes)
			continue
		}
		for token := range declarations(found[1]) {
			if _, ok := stated[token]; !ok && !slices.Contains(optional, token) {
				t.Errorf("%s leaves --%s to Classic's value in %s", scheme, token, half.Schemes)
			}
		}
	}
	for name := range blocks {
		if name == classic || !slices.Contains(offered, name) {
			t.Errorf("%s holds a block for %q, which no menu offers", half.Schemes, name)
		}
	}
}

// CheckClassicDarkIsTheSameUnderTheSystemAsWhenChosen fails when half's Classic file states the
// system's dark differently from the chosen dark.
func CheckClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t testing.TB, half Half) {
	t.Helper()
	text := Read(t, half.Classic)
	system, chosen := systemDark.FindStringSubmatch(text), darkBlock.FindStringSubmatch(text)
	if system == nil || chosen == nil {
		t.Fatalf("%s lacks a dark block", half.Classic)
	}
	if fmt.Sprint(declarations(system[1])) != fmt.Sprint(declarations(chosen[1])) {
		t.Errorf("%s: the system's dark block differs from the chosen one:\n%v\n%v", half.Classic, declarations(system[1]), declarations(chosen[1]))
	}
}

// Palette answers the colour a token has in one scheme and theme.
type Palette func(token string) (string, error)

// Palettes answers each scheme in offered with its palette per theme, read from every half
// together as the page cascades them; a scheme's missing token falls back to Classic's value for
// the same theme.
func Palettes(t testing.TB, halves []Half, offered []string) map[string]map[Theme]Palette {
	t.Helper()
	classic := map[Theme]map[string]string{Light: {}, Dark: {}}
	stated := map[string]map[string]string{}
	for _, half := range halves {
		text := Read(t, half.Classic)
		light, dark := classicBlock.FindStringSubmatch(text), darkBlock.FindStringSubmatch(text)
		if light == nil || dark == nil {
			t.Fatalf("%s lacks its :root or its [data-theme='dark'] block", half.Classic)
		}
		maps.Copy(classic[Light], declarations(light[1]))
		maps.Copy(classic[Dark], declarations(dark[1]))
		for _, match := range schemeBlock.FindAllStringSubmatch(Read(t, half.Schemes), -1) {
			if stated[match[1]] == nil {
				stated[match[1]] = map[string]string{}
			}
			maps.Copy(stated[match[1]], declarations(match[2]))
		}
	}
	all := map[string]map[Theme]Palette{}
	for _, scheme := range offered {
		all[scheme] = map[Theme]Palette{Light: paletteOf(scheme, Light, stated[scheme], classic[Light]),
			Dark: paletteOf(scheme, Dark, stated[scheme], classic[Dark])}
	}
	return all
}

// paletteOf answers scheme's palette on side: its own value for a token, else Classic's.
func paletteOf(scheme string, side Theme, own, classic map[string]string) Palette {
	return func(token string) (string, error) {
		value, found := own[token]
		if !found {
			value, found = classic[token]
		}
		if !found {
			return "", fmt.Errorf("--%s is stated by neither %s nor Classic", token, scheme)
		}
		if pair := lightDark.FindStringSubmatch(value); pair != nil {
			if side == Light {
				return pair[1], nil
			}
			return pair[2], nil
		}
		return value, nil
	}
}

// WCAG 2.x relative luminance: channels are linearised below linearLimit and by the sRGB curve
// above it, then weighted by the eye's sensitivity to each. flare is the ambient light WCAG adds to
// both luminances before dividing them.
const (
	channelMax    = 255
	linearLimit   = 0.04045
	linearDivisor = 12.92
	curveOffset   = 0.055
	curveDivisor  = 1.055
	curvePower    = 2.4
	flare         = 0.05
)

var luminanceWeights = [3]float64{0.2126, 0.7152, 0.0722}

// channels reads #rgb or #rrggbb, refusing any other form rather than skipping the check.
func channels(colour string) ([3]float64, error) {
	var digits []string
	if match := shortHex.FindStringSubmatch(colour); match != nil {
		for _, digit := range match[1:] {
			digits = append(digits, digit+digit)
		}
	} else if match := longHex.FindStringSubmatch(colour); match != nil {
		digits = match[1:]
	} else {
		return [3]float64{}, fmt.Errorf("cannot read the colour %q; teach the structure package its form", colour)
	}
	var rgb [3]float64
	for index, pair := range digits {
		value, err := strconv.ParseUint(pair, 16, 8)
		if err != nil {
			return rgb, err
		}
		rgb[index] = float64(value)
	}
	return rgb, nil
}

func luminance(colour string) (float64, error) {
	rgb, err := channels(colour)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for index, channel := range rgb {
		unit := channel / channelMax
		linear := unit / linearDivisor
		if unit > linearLimit {
			linear = math.Pow((unit+curveOffset)/curveDivisor, curvePower)
		}
		sum += linear * luminanceWeights[index]
	}
	return sum, nil
}

// Contrast answers WCAG 2.x's contrast ratio between two colours, whichever is lighter.
func Contrast(first, second string) (float64, error) {
	a, err := luminance(first)
	if err != nil {
		return 0, err
	}
	b, err := luminance(second)
	if err != nil {
		return 0, err
	}
	return (math.Max(a, b) + flare) / (math.Min(a, b) + flare), nil
}
