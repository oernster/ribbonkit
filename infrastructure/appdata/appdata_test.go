package appdata

import (
	"errors"
	"testing"
)

func TestAnUnsetOrEmptyVariableIsRefused(t *testing.T) {
	t.Parallel()
	for _, lookup := range []func(string) (string, bool){
		func(string) (string, bool) { return "", false },
		func(string) (string, bool) { return "", true },
	} {
		if _, err := Dir(testApp, lookup); !errors.Is(err, ErrNoAppData) {
			t.Errorf("got %v", err)
		}
	}
}
