package arranger

import (
	"errors"
	"math"
	"testing"

	"github.com/oernster/ribbonkit/domain/placement"
)

// pageScale is the scale the page reported in the reproduction of 2026-09-28: Windows' text size
// at 125 percent on a display at 96 DPI draws the page at 1.25, which the display's DPI does not
// say.
const pageScale = 1.25

// The ribbon is sized by the scale the page is drawn at, not the display's DPI: two digital cells
// are 336 x 106 DIP, so 420 x 133 pixels at 1.25 on the 96 DPI primary, rounded up.
func TestTheRibbonIsSizedByTheScaleThePageIsDrawnAt(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	if got, _ := r.arranger.Launch(); got.Size != (placement.Size{Width: 336, Height: 106}) {
		t.Fatalf("before the page reports: %+v", got.Size)
	}
	if err := r.arranger.SetPixelsPerDIP(pageScale); err != nil {
		t.Fatal(err)
	}
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != (placement.Size{Width: 420, Height: 133}) || got.At.X != (1920-420)/2 {
		t.Errorf("at 1.25: %+v", got)
	}
}

// The page keeps the scale it was made at when the ribbon moves to a display at another DPI, so the
// reported scale holds there too: the 144 DPI secondary does not make it 1.5.
func TestTheReportedScaleHoldsOnADisplayAtAnotherDPI(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	if err := r.arranger.SetPixelsPerDIP(pageScale); err != nil {
		t.Fatal(err)
	}
	got, err := r.arranger.Moved(placement.Point{X: 2100, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != (placement.Size{Width: 420, Height: 133}) {
		t.Errorf("on the secondary: %+v", got.Size)
	}
}

// A panel is sized by the same scale: 560 x 760 DIP is 700 x 950 pixels at 1.25.
func TestAPanelIsSizedByTheScaleThePageIsDrawnAt(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	if err := r.arranger.SetPixelsPerDIP(pageScale); err != nil {
		t.Fatal(err)
	}
	got, err := r.arranger.Centred(placement.Point{X: 100, Y: 100}, placement.Size{Width: 560, Height: 760})
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != (placement.Size{Width: 700, Height: 950}) {
		t.Errorf("panel: %+v", got.Size)
	}
}

// A long ribbon at 1.25 fits the room the display offers at that scale, then scrolls: 12 cells in
// 1920 pixels are fitted to 1536 DIP, which is the whole 1920 again.
func TestAScaledRibbonFitsTheRoomTheDisplayOffersAtThatScale(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(12))
	if err := r.arranger.SetPixelsPerDIP(pageScale); err != nil {
		t.Fatal(err)
	}
	got, err := r.arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Scrolls || got.Size.Width > primaryMonitor.Work.Width() {
		t.Errorf("12 cells at 1.25: %+v", got)
	}
}

func TestAnUnusableScaleIsRefused(t *testing.T) {
	t.Parallel()
	r := newRig(horizontal(), cells(2))
	for _, scale := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if err := r.arranger.SetPixelsPerDIP(scale); !errors.Is(err, ErrUnusableScale) {
			t.Errorf("scale %v: got %v", scale, err)
		}
	}
	if got, _ := r.arranger.Launch(); got.Size != (placement.Size{Width: 336, Height: 106}) {
		t.Errorf("a refused scale changed the size: %+v", got.Size)
	}
}
