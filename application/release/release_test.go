package release

import (
	"context"
	"errors"
	"testing"
)

// fakeSource answers every check with info; with err instead when err is set.
type fakeSource struct {
	info Info
	err  error
}

func (f fakeSource) LatestRelease(context.Context) (Info, error) { return f.info, f.err }

const runningVersion = "2.0.0"

var running = Build{Version: runningVersion, Platform: PlatformWindows}

var newerRelease = Info{
	Version: "v2.1.0",
	PageURL: "https://example.test/release",
	Assets: []Asset{
		{Name: "App.dmg", DownloadURL: "https://example.test/mac"},
		{Name: "AppSetup.EXE", DownloadURL: "https://example.test/windows"},
	},
}

func TestIsNewerVersionComparesDottedIntegers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		latest, current string
		newer           bool
	}{
		{"2.1.0", "2.0.0", true},
		{"v2.1.0", "2.0.0", true},
		{"V3.0.0", "2.9.9", true},
		{" 2.0.1 ", "2.0.0", true},
		{"2.0.0.1", "2.0.0", true},
		{"2.10.0", "2.9.0", true},
		{"2.0.0", "2.0.0", false},
		{"v2.0.0", "2.0.0", false},
		{"1.9.0", "2.0.0", false},
		{"2.0.0", "2.0.0.1", false},
		{"2.1.0-rc1", "2.0.0", false},
		{"banana", "2.0.0", false},
		{"2.1.0", "0.0.0-dev", false},
		{"", "2.0.0", false},
	}
	for _, c := range cases {
		if got := IsNewerVersion(c.latest, c.current); got != c.newer {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", c.latest, c.current, got, c.newer)
		}
	}
}

func TestEachSystemDownloadsItsOwnAsset(t *testing.T) {
	t.Parallel()
	for goos, want := range map[string]string{"windows": PlatformWindows, "darwin": PlatformMacOS, "linux": PlatformLinux, "freebsd": PlatformLinux} {
		if got := PlatformKeyFor(goos); got != want {
			t.Errorf("%s is %s, want %s", goos, got, want)
		}
	}
	for platform, want := range map[string]string{
		PlatformWindows: "https://example.test/windows", PlatformMacOS: "https://example.test/mac", PlatformLinux: "", "amiga": "",
	} {
		if got := SelectAssetURL(newerRelease.Assets, platform); got != want {
			t.Errorf("%s downloads %q, want %q", platform, got, want)
		}
	}
	if got := SelectAssetURL(nil, PlatformWindows); got != "" {
		t.Errorf("a release with no assets offered %q", got)
	}
}

// A newer release is offered with this platform's download and its page.
func TestANewerReleaseIsOffered(t *testing.T) {
	t.Parallel()
	want := Status{
		Current: runningVersion, Latest: "v2.1.0", Available: true,
		DownloadURL: "https://example.test/windows", PageURL: "https://example.test/release",
	}
	for _, manual := range []bool{false, true} {
		if got := Check(context.Background(), fakeSource{info: newerRelease}, running, "", manual); got != want {
			t.Errorf("manual %v answered %+v", manual, got)
		}
	}
}

// A source out of reach is no update and no latest version, whoever asked.
func TestAnUnreachableSourceOffersNothing(t *testing.T) {
	t.Parallel()
	got := Check(context.Background(), fakeSource{err: errors.New("offline")}, running, "", true)
	if got != (Status{Current: runningVersion}) {
		t.Errorf("answered %+v", got)
	}
}

// The running version itself is seen but not offered.
func TestTheRunningVersionIsNotOffered(t *testing.T) {
	t.Parallel()
	same := Info{Version: "v" + runningVersion, PageURL: "https://example.test/release"}
	if got := Check(context.Background(), fakeSource{info: same}, running, "", true); got.Available || got.Latest != same.Version {
		t.Errorf("answered %+v", got)
	}
}

// A skipped release is never offered unasked; asking still offers it. A later release than the one
// skipped is offered either way.
func TestASkippedReleaseIsOfferedOnlyWhenAskedFor(t *testing.T) {
	t.Parallel()
	source := fakeSource{info: newerRelease}
	if got := Check(context.Background(), source, running, newerRelease.Version, false); got.Available || got.Latest != newerRelease.Version {
		t.Errorf("the automatic check answered %+v", got)
	}
	if got := Check(context.Background(), source, running, newerRelease.Version, true); !got.Available {
		t.Errorf("the manual check answered %+v", got)
	}
	if got := Check(context.Background(), source, running, "v2.0.5", false); !got.Available {
		t.Errorf("a release later than the one skipped answered %+v", got)
	}
}
