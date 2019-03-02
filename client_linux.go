//+build linux

package kobject

import (
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// newConn dials out to kobject uevent netlink and returns a conn to use to
// listen for events.
func newConn() (conn, error) {
	c, err := netlink.Dial(unix.NETLINK_KOBJECT_UEVENT, &netlink.Config{
		// TODO(mdlayher): replace with constant in x/sys/unix.
		Groups: 0x1,
	})
	if err != nil {
		return nil, err
	}

	return &sysConn{
		c: c,
	}, nil
}

var _ conn = &sysConn{}

// A sysConn implements conn over a *netlink.Conn.
type sysConn struct {
	c *netlink.Conn
}

func (sc *sysConn) Read(b []byte) (int, error) {
	raw, err := sc.c.SyscallConn()
	if err != nil {
		return 0, err
	}

	var n int
	doErr := raw.Read(func(fd uintptr) bool {
		n, err = unix.Read(int(fd), b)

		// When the socket is in non-blocking mode, we might see
		// EAGAIN and end up here. In that case, return false to
		// let the poller wait for readiness. See the source code
		// for internal/poll.FD.RawRead for more details.
		return err != unix.EAGAIN
	})
	if doErr != nil {
		return 0, doErr
	}

	return n, err
}

func (sc *sysConn) Close() error                  { return sc.c.Close() }
func (sc *sysConn) SetDeadline(t time.Time) error { return sc.c.SetDeadline(t) }
