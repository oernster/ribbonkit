package appdata

import "testing"

func TestTheFolderIsSampleRibbonUnderAppData(t *testing.T) {
	t.Parallel()
	got, err := Dir(testApp, func(string) (string, bool) { return `C:\Users\Someone\AppData\Roaming`, true })
	if err != nil || got != `C:\Users\Someone\AppData\Roaming\SampleRibbon` {
		t.Errorf("got %q (%v)", got, err)
	}
}
