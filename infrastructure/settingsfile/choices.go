package settingsfile

import (
	"encoding/json"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The keys the ribbon's own choices are written under. Each application places them among its own
// keys in its own writing order.
const (
	KeyColour        = "colour"
	KeyOrientation   = "orientation"
	KeyTheme         = "theme"
	KeyAlwaysOnTop   = "alwaysOnTop"
	KeyPlacement     = "placement"
	KeySkippedUpdate = "skippedUpdate"
	KeyPinned        = "pinned"
	KeyLastEdge      = "lastEdge"
	KeyOpacity       = "opacity"
	KeyScale         = "scale"
	KeyPullOutSide   = "pullOutSide"
)

// storedPlacement is the placement as the file holds it.
type storedPlacement struct {
	Device string `json:"device"`
	Work   struct {
		Left   int `json:"left"`
		Top    int `json:"top"`
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
	} `json:"work"`
	DPI    int `json:"dpi"`
	Offset struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"offset"`
}

// storedEdge is the edge the ribbon last stood against as the file holds it.
type storedEdge struct {
	Device string `json:"device"`
	Edge   string `json:"edge"`
}

// ReadChoices reads the ribbon's choices from object into choices; a key missing or holding the
// wrong type leaves its value as it was.
func ReadChoices(object Object, choices *ribbon.Choices) {
	Read(object, KeyColour, &choices.Colour)
	Read(object, KeyOrientation, &choices.Orientation)
	Read(object, KeyTheme, &choices.Theme)
	Read(object, KeyAlwaysOnTop, &choices.AlwaysOnTop)
	Read(object, KeySkippedUpdate, &choices.SkippedUpdate)
	Read(object, KeyPinned, &choices.Pinned)
	Read(object, KeyOpacity, &choices.Opacity)
	Read(object, KeyScale, &choices.Scale)
	Read(object, KeyPullOutSide, &choices.PullOutSide)
	choices.Placement = decodePlacement(object[KeyPlacement])
	choices.LastEdge = decodeEdge(object[KeyLastEdge])
}

// ChoiceValues answers the value of every key the ribbon's choices are written under.
func ChoiceValues(choices ribbon.Choices) map[string]any {
	return map[string]any{
		KeyColour: choices.Colour, KeyOrientation: choices.Orientation, KeyTheme: choices.Theme,
		KeyAlwaysOnTop: choices.AlwaysOnTop, KeyPlacement: encodePlacement(choices.Placement),
		KeySkippedUpdate: choices.SkippedUpdate, KeyPinned: choices.Pinned,
		KeyLastEdge: encodeEdge(choices.LastEdge), KeyOpacity: choices.Opacity, KeyScale: choices.Scale,
		KeyPullOutSide: choices.PullOutSide,
	}
}

// decodePlacement reads the placement; none where it is missing, malformed or names no display.
func decodePlacement(raw json.RawMessage) *placement.Stored {
	var stored storedPlacement
	if len(raw) == 0 || json.Unmarshal(raw, &stored) != nil || stored.Device == "" {
		return nil
	}
	return &placement.Stored{
		Device: stored.Device,
		Work:   placement.Rect{Left: stored.Work.Left, Top: stored.Work.Top, Right: stored.Work.Right, Bottom: stored.Work.Bottom},
		DPI:    stored.DPI,
		Offset: placement.Point{X: stored.Offset.X, Y: stored.Offset.Y},
	}
}

// decodeEdge reads the remembered edge; none where it is missing, malformed or names no display. An
// edge that names no side is forgotten as the choices are normalised.
func decodeEdge(raw json.RawMessage) *placement.Against {
	var stored storedEdge
	if len(raw) == 0 || json.Unmarshal(raw, &stored) != nil || stored.Device == "" {
		return nil
	}
	return &placement.Against{Device: stored.Device, Edge: placement.Edge(stored.Edge)}
}

// encodePlacement answers the placement as the file holds it; nil for none.
func encodePlacement(stored *placement.Stored) *storedPlacement {
	if stored == nil {
		return nil
	}
	var out storedPlacement
	out.Device = stored.Device
	out.Work.Left, out.Work.Top, out.Work.Right, out.Work.Bottom = stored.Work.Left, stored.Work.Top, stored.Work.Right, stored.Work.Bottom
	out.DPI = stored.DPI
	out.Offset.X, out.Offset.Y = stored.Offset.X, stored.Offset.Y
	return &out
}

// encodeEdge answers the remembered edge as the file holds it; nil for none.
func encodeEdge(last *placement.Against) *storedEdge {
	if last == nil {
		return nil
	}
	return &storedEdge{Device: last.Device, Edge: string(last.Edge)}
}
