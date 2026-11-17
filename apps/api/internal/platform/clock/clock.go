// Package clock gives the commands of the binary the clock they pass to the
// use cases. A use case declares the clock it needs as an interface and takes
// the current time from it, never from the machine, so a test decides what
// time it is.
package clock

import "time"

// System tells the time of the machine.
type System struct{}

// Now returns the current time, in UTC, cut to the microsecond: that is what
// PostgreSQL keeps of an instant, so a time a use case returns is the time it
// stored.
func (System) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
