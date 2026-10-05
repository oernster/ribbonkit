package update

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/application/release"
)

// fakeDoer answers every request with status and body (with err instead when err is set); it keeps
// the request it saw.
type fakeDoer struct {
	status int
	body   io.Reader
	err    error
	seen   *http.Request
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	f.seen = req
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{StatusCode: f.status, Body: io.NopCloser(f.body)}, nil
}

func answering(status int, body string) *fakeDoer {
	return &fakeDoer{status: status, body: strings.NewReader(body)}
}

// failingReader fails every read, as a connection dropped mid-answer does.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection dropped") }

// sampleAPIURL is the endpoint the tests ask, for the repository their sample release comes from.
var sampleAPIURL = LatestReleaseAPIURL("oernster/TimeRibbon")

const releaseJSON = `{"tag_name":"v2.1.0","html_url":"https://github.com/oernster/TimeRibbon/releases/tag/v2.1.0","assets":[
	{"name":"TimeRibbonSetup.exe","browser_download_url":"https://github.com/oernster/TimeRibbon/releases/download/v2.1.0/a.exe"},
	{"name":"","browser_download_url":"https://github.com/nameless"},
	{"name":"TimeRibbon.dmg","browser_download_url":""},
	{"name":"timeribbon.flatpak","browser_download_url":"https://github.com/oernster/TimeRibbon/releases/download/v2.1.0/a.flatpak"}]}`

func TestTheLatestReleaseIsReadWithOnlyWholeAssets(t *testing.T) {
	t.Parallel()
	doer := answering(http.StatusOK, releaseJSON)
	got, err := NewWith(sampleAPIURL, doer).LatestRelease(context.Background())
	want := []release.Asset{
		{Name: "TimeRibbonSetup.exe", DownloadURL: "https://github.com/oernster/TimeRibbon/releases/download/v2.1.0/a.exe"},
		{Name: "timeribbon.flatpak", DownloadURL: "https://github.com/oernster/TimeRibbon/releases/download/v2.1.0/a.flatpak"},
	}
	if err != nil || got.Version != "v2.1.0" || got.PageURL != "https://github.com/oernster/TimeRibbon/releases/tag/v2.1.0" || !slices.Equal(got.Assets, want) {
		t.Fatalf("answered %+v, %v", got, err)
	}
	if doer.seen.URL.String() != sampleAPIURL || doer.seen.Header.Get("Accept") != acceptHeader || doer.seen.Method != http.MethodGet {
		t.Errorf("asked %s %s with Accept %q", doer.seen.Method, doer.seen.URL, doer.seen.Header.Get("Accept"))
	}
}

func TestTheProductionSourceAsksTheNamedRepositoryAndGivesUp(t *testing.T) {
	t.Parallel()
	source := New("someone/SampleRibbon")
	client, ok := source.client.(*http.Client)
	if !ok || client.Timeout != requestTimeout {
		t.Errorf("asks through %#v", source.client)
	}
	if source.apiURL != "https://api.github.com/repos/someone/SampleRibbon/releases/latest" {
		t.Errorf("%s is not the named repository's latest release", source.apiURL)
	}
}

func TestEveryUnusableAnswerIsAnError(t *testing.T) {
	t.Parallel()
	cases := map[string]*fakeDoer{
		"unreachable":    {err: errors.New("no route")},
		"not found":      answering(http.StatusNotFound, releaseJSON),
		"dropped":        {status: http.StatusOK, body: failingReader{}},
		"not JSON":       answering(http.StatusOK, "<html>"),
		"no tag":         answering(http.StatusOK, `{"html_url":"https://github.com/oernster/TimeRibbon/releases/tag/v2.1.0"}`),
		"no page":        answering(http.StatusOK, `{"tag_name":"v2.1.0"}`),
		"wrong types":    answering(http.StatusOK, `{"tag_name":7,"html_url":"https://github.com/oernster/TimeRibbon/releases/tag/v2.1.0"}`),
		"oversized":      answering(http.StatusOK, `{"tag_name":"v2.1.0","html_url":"`+strings.Repeat("x", maxBody)+`"}`),
		"not an object":  answering(http.StatusOK, `[]`),
		"assets unusual": answering(http.StatusOK, `{"tag_name":"v2.1.0","html_url":"x","assets":{}}`),
	}
	for name, doer := range cases {
		if got, err := NewWith(sampleAPIURL, doer).LatestRelease(context.Background()); err == nil {
			t.Errorf("%s answered %+v with no error", name, got)
		}
	}
}

func TestAnAddressThatCannotBeAskedIsAnError(t *testing.T) {
	t.Parallel()
	if _, err := NewWith("://", answering(http.StatusOK, releaseJSON)).LatestRelease(context.Background()); err == nil {
		t.Error("a malformed address was asked")
	}
}

// FR-509: an address the release names is handed to the desktop to open, so only https on GitHub is
// taken. A page outside it refuses the release; a download outside it is left out.
func TestOnlyHTTPSAddressesOnGitHubAreTaken(t *testing.T) {
	t.Parallel()
	const page = "https://github.com/oernster/TimeRibbon/releases/tag/v2.1.0"
	for _, address := range []string{
		`\\203.0.113.9\share\TimeRibbonSetup.exe`, "file:///C:/Windows/System32/calc.exe", "http://github.com/x",
		"https://github.com.example.test/x", "https://user@github.com/x", "https://github.com:8443/x", "https:github.com/x",
		"https://api.github.com/x", "%zz",
	} {
		asset := `{"name":"TimeRibbonSetup.exe","browser_download_url":` + strconv.Quote(address) + `}`
		body := `{"tag_name":"v2.1.0","html_url":` + strconv.Quote(page) + `,"assets":[` + asset + `]}`
		got, err := NewWith(sampleAPIURL, answering(http.StatusOK, body)).LatestRelease(context.Background())
		if err != nil || len(got.Assets) != 0 {
			t.Errorf("download %s was taken: %+v (%v)", address, got, err)
		}
		body = `{"tag_name":"v2.1.0","html_url":` + strconv.Quote(address) + `}`
		if got, err := NewWith(sampleAPIURL, answering(http.StatusOK, body)).LatestRelease(context.Background()); err == nil {
			t.Errorf("page %s was taken: %+v", address, got)
		}
	}
}
