// Package update answers the application's ReleaseSource from GitHub's latest-release endpoint, the
// one network request a ribbon makes (TimeRibbon NFR-S-1, FR-509). The endpoint answers only a published
// release that is neither a draft nor a prerelease, so a tag pushed during development is never
// seen: the guard is the endpoint's own contract, not a check here. The HTTP client is injected, so
// the tests never touch the network. Ported from PigeonPost.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/oernster/ribbonkit/application/release"
)

// LatestReleaseAPIURL answers GitHub's latest-release endpoint for repository, written as
// owner/name. The kit names no repository of its own: each application passes its own, so one
// product's check never asks after another's releases.
func LatestReleaseAPIURL(repository string) string {
	return "https://api.github.com/repos/" + repository + "/releases/latest"
}

// acceptHeader asks the GitHub API for its JSON.
const acceptHeader = "application/vnd.github+json"

// requestTimeout bounds the one request, so a check never waits long.
const requestTimeout = 5 * time.Second

// maxBody caps what is read of the answer: a release's JSON is a few kilobytes, so anything near
// this is not one. A size the server states is never trusted.
const maxBody = 1 << 20

// releaseScheme and releaseHost are the only scheme and host a release's page or download may name:
// both are handed to the desktop to open, which must never be handed a file, a share or another
// site (FR-509). A download refused here is left out, so the release page stands in for it.
const (
	releaseScheme = "https"
	releaseHost   = "github.com"
)

// onGitHub answers whether address is an https address on GitHub's own host, naming no user and
// no port.
func onGitHub(address string) bool {
	parsed, err := url.Parse(address)
	return err == nil && parsed.Scheme == releaseScheme && parsed.Host == releaseHost && parsed.User == nil
}

// Doer sends one HTTP request; *http.Client is one and the tests stand in another.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// releasePayload is the part of GitHub's latest-release JSON the check reads.
type releasePayload struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Assets  []assetPayload `json:"assets"`
}

type assetPayload struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// GitHub is an release.Source over the GitHub API.
type GitHub struct {
	apiURL string
	client Doer
}

// New answers the production source for repository (owner/name), with a client that gives up
// after requestTimeout.
func New(repository string) *GitHub {
	return NewWith(LatestReleaseAPIURL(repository), &http.Client{Timeout: requestTimeout})
}

// NewWith answers a source asking apiURL through client.
func NewWith(apiURL string, client Doer) *GitHub {
	return &GitHub{apiURL: apiURL, client: client}
}

// LatestRelease answers the latest published release; an error when it cannot be read, which the
// service treats as no answer.
func (g *GitHub) LatestRelease(ctx context.Context) (release.Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiURL, nil)
	if err != nil {
		return release.Info{}, fmt.Errorf("building the release request: %w", err)
	}
	req.Header.Set("Accept", acceptHeader)
	resp, err := g.client.Do(req)
	if err != nil {
		return release.Info{}, fmt.Errorf("asking for the latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return release.Info{}, fmt.Errorf("asking for the latest release: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return release.Info{}, fmt.Errorf("reading the release: %w", err)
	}
	var payload releasePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return release.Info{}, fmt.Errorf("reading the release: %w", err)
	}
	if payload.TagName == "" || payload.HTMLURL == "" {
		return release.Info{}, fmt.Errorf("the release names no version or no page")
	}
	if !onGitHub(payload.HTMLURL) {
		return release.Info{}, fmt.Errorf("the release's page %q is not on %s", payload.HTMLURL, releaseHost)
	}
	assets := make([]release.Asset, 0, len(payload.Assets))
	for _, asset := range payload.Assets {
		if asset.Name != "" && onGitHub(asset.DownloadURL) {
			assets = append(assets, release.Asset{Name: asset.Name, DownloadURL: asset.DownloadURL})
		}
	}
	return release.Info{Version: payload.TagName, PageURL: payload.HTMLURL, Assets: assets}, nil
}
