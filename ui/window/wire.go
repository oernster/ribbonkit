package window

// The window's half of the wire between Go and the page; the application's own types are its half.
// Each type here is stated a second time in the page's wire.ts; a structural test compares the two,
// since the type checker sees only the TypeScript and the marshaller sees only these.

import "github.com/oernster/ribbonkit/application/release"

// aboutDTO is what the About panel shows (FR-607).
type aboutDTO struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Author    string      `json:"author"`
	Copyright string      `json:"copyright"`
	Credits   []creditDTO `json:"credits"`
}

// creditDTO is one component the application ships.
type creditDTO struct {
	Name    string `json:"name"`
	Licence string `json:"licence"`
	Role    string `json:"role"`
}

// updateDTO is what the update panel shows (FR-509). Latest is empty when GitHub could not be
// reached. The addresses stay in Go: the page asks Go to open what it offered, never names one.
type updateDTO struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

// Box is a rectangle inside the window, in the page's units: where Shown places the ribbon and what
// is pulled out beside it. The application carries it to the page as it is.
type Box struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func updateOf(status release.Status) updateDTO {
	return updateDTO{Current: status.Current, Latest: status.Latest, UpdateAvailable: status.Available}
}
