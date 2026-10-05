package delivery

import (
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/identity"
)

// The lock's name follows Wails' formula, as its source builds it. Measured 2026-09-28: TimeRibbon's
// installed Flatpak owned exactly this name.
func TestTheSingleInstanceNameFollowsWails(t *testing.T) {
	t.Parallel()
	if got := singleInstanceName("uk.codecrafter.TimeRibbon"); got != "org.wails_app_uk_codecrafter_TimeRibbon.SingleInstance" {
		t.Errorf("got %q", got)
	}
}

func TestEveryNameIsWrittenForTheShell(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	ShellNames(&out, identity.App{Name: "TestRibbon", AppID: "uk.example.Test-Ribbon"}, "© A. Maker")
	want := "APP_NAME='TestRibbon'\nAPP_ID='uk.example.Test-Ribbon'\nBIN_NAME='testribbon'\n" +
		"SINGLE_INSTANCE_NAME='org.wails_app_uk_example_Test_Ribbon.SingleInstance'\nCOPYRIGHT='© A. Maker'\n"
	if out.String() != want {
		t.Errorf("wrote %q, want %q", out.String(), want)
	}
}
