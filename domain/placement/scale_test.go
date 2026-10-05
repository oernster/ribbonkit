package placement

import "testing"

func TestPerDIPFollowsTheDisplaysDPI(t *testing.T) {
	t.Parallel()
	for dpi, want := range map[int]float64{96: 1, 120: 1.25, 144: 1.5, 240: 2.5, 0: 1} {
		if got := PerDIPOf(dpi); got != want {
			t.Errorf("PerDIPOf(%d) = %v, want %v", dpi, got, want)
		}
	}
}

// Measured 2026-09-28: a 188 by 104 DIP ribbon drawn at 1.25 on a 96 DPI display needs a window
// of 235 by 130 pixels; sized for 96 DPI it cut off the page.
func TestPixelsAreRoundedUpSoThePageIsNeverCutOff(t *testing.T) {
	t.Parallel()
	cases := []struct {
		length int
		perDIP float64
		want   int
	}{
		{188, 1.25, 235},
		{104, 1.25, 130},
		{188, 1, 188},
		{100, 1.15, 115},
		{188, 1.1, 207},
		{176, 2.5, 440},
	}
	for _, c := range cases {
		if got := PixelsOf(c.length, c.perDIP); got != c.want {
			t.Errorf("PixelsOf(%d, %v) = %d, want %d", c.length, c.perDIP, got, c.want)
		}
	}
}

func TestDIPAreRoundedDownSoNothingSizedInThemOverflows(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pixels int
		perDIP float64
		want   int
	}{
		{1920, 1.25, 1536},
		{1032, 1.25, 825},
		{1920, 1, 1920},
		{115, 1.15, 100},
	}
	for _, c := range cases {
		if got := DIPOf(c.pixels, c.perDIP); got != c.want {
			t.Errorf("DIPOf(%d, %v) = %d, want %d", c.pixels, c.perDIP, got, c.want)
		}
	}
}
