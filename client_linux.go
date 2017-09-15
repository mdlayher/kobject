//+build linux

package kobject

import (
	"io"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// newReadCloser dials out to kobject uevent netlink and returns an
// io.ReadCloser to use to listen for events.
func newReadCloser() (io.ReadCloser, error) {
	c, err := netlink.Dial(unix.NETLINK_KOBJECT_UEVENT, &netlink.Config{
		// TODO(mdlayher): replace with constant in x/sys/unix.
		Groups: 0x1,
	})
	if err != nil {
		return nil, err
	}

	rwc, err := c.ReadWriteCloser()
	if err != nil {
		return nil, err
	}

	// Drop io.Writer; no need for it.
	return rwc, nil
}
