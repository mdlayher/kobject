//+build !linux

package kobject

import (
	"fmt"
	"io"
	"runtime"
)

// newReadCloser always returns an error on unsupported platforms.
func newReadCloser() (io.ReadCloser, error) {
	return nil, fmt.Errorf("kobject unimplemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}
