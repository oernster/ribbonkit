// Package zones resolves IANA time zone ids for every ribbon that shows local time. The tz database
// is built into the binary through time/tzdata. Windows has no zone files of its own, so there the
// built-in rules are the only ones read; on macOS and Linux time.LoadLocation reads the system's zone
// files first and falls back to the built-in rules only when a zone is missing there.
package zones

import (
	"fmt"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
)

// localName is the id Go reads as the machine's own zone rather than as a zone of the database.
const localName = "Local"

// Resolver resolves zone ids, loading each once. The zero Resolver is ready to use; it is safe for
// use from several goroutines.
type Resolver struct {
	resolved sync.Map
}

// Resolve answers the location for zone, cached once loaded. An empty id and "Local" are refused
// rather than read as UTC and as the machine's zone, which is what Go would make of them.
func (r *Resolver) Resolve(zone string) (*time.Location, error) {
	if cached, ok := r.resolved.Load(zone); ok {
		return cached.(*time.Location), nil
	}
	if zone == "" || strings.EqualFold(zone, localName) {
		return nil, fmt.Errorf("%q is not a time zone id", zone)
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return nil, err
	}
	r.resolved.Store(zone, location)
	return location, nil
}
