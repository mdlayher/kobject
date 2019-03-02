//+build !linux

package kobject

import (
	"fmt"
	"runtime"
)

// newConn always returns an error on unsupported platforms.
func newConn() (conn, error) {
	return nil, fmt.Errorf("kobject: unimplemented on %s/%s", runtime.GOOS, runtime.GOARCH)
}
