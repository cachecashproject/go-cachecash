package common

import (
	"os"
)

// Main provides a convenient thunk to allow writing a CLI with more idiomatic
// go - returning error rather than calling os.Exit. A small thunk is still
// needed. If mainC returns non-nil the error reported to stderr and a non-zero
// exit code is reported to the OS.
//
// ```
// func main() {
//     common.Main(mainC)
// }
// ```
func Main(mainC func() error) {
	if err := mainC(); err != nil {
		if _, err := os.Stderr.WriteString(err.Error() + "\n"); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
