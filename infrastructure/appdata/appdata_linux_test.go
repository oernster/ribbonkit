package appdata

import "testing"

func TestTheFolderIsSampleRibbonUnderTheConfigurationFolder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]string
		want   string
	}{
		{"XDG_CONFIG_HOME", map[string]string{configHomeVariable: "/cfg", homeVariable: "/home/someone"}, "/cfg/SampleRibbon"},
		{"home only", map[string]string{homeVariable: "/home/someone"}, "/home/someone/.config/SampleRibbon"},
	}
	for _, c := range cases {
		lookup := func(name string) (string, bool) {
			value, ok := c.values[name]
			return value, ok
		}
		if got, err := Dir(testApp, lookup); err != nil || got != c.want {
			t.Errorf("%s: got %q (%v), want %q", c.name, got, err, c.want)
		}
	}
}
