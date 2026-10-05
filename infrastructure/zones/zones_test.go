package zones

import "testing"

// An id the database does not know is refused; so are the two Go would quietly accept.
func TestAnUnknownZoneIsRefused(t *testing.T) {
	t.Parallel()
	var resolver Resolver
	for _, zone := range []string{"Not/AZone", "", "Local", "local"} {
		if _, err := resolver.Resolve(zone); err == nil {
			t.Errorf("%q resolved", zone)
		}
	}
}

// A zone is loaded once; the second resolve answers the same location.
func TestAResolvedZoneIsCached(t *testing.T) {
	t.Parallel()
	var resolver Resolver
	first, err := resolver.Resolve("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	second, _ := resolver.Resolve("Europe/London")
	if first != second || first.String() != "Europe/London" {
		t.Errorf("answered %v then %v", first, second)
	}
}
