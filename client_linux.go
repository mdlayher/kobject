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

func (sc *sysConn) TryRead(b []byte) (int, bool, error) {
	raw, err := sc.c.SyscallConn()
	if err != nil {
		return 0, false, err
	}

	var (
		n      int
		done   bool
		outErr error
	)

	// Deals with any errors to raw.Read, populating outErr for the caller if
	// necessary, and handling the appropriate runtime network poller integration
	// logic.
	handle := func(err error) bool {
		eagain := err == unix.EAGAIN
		if !eagain {
			// Only report non-EAGAIN errors to the caller.
			outErr = err
		}

		// When the socket is in non-blocking mode, we might see
		// EAGAIN and end up here. In that case, return false to
		// let the poller wait for readiness. See the source code
		// for internal/poll.FD.RawRead for more details.
		return !eagain
	}

	doErr := raw.Read(func(fd uintptr) bool {
		// How many bytes are available from the socket?
		n, _, _, _, err = unix.Recvmsg(int(fd), b, nil, unix.MSG_PEEK)
		if err != nil {
			return handle(err)
		}

		if len(b) < n {
			// The buffer isn't large enough to be filled in one call to Read.
			// Inform the poller that we're immediately ready for more I/O, but
			// also inform the caller that TryRead didn't complete.
			return true
		}

		// The buffer is large enough to be filled in one call to Read, perform
		// the Read and inform the caller it completed as well.
		done = true
		n, err = unix.Read(int(fd), b)
		return handle(err)
	})
	if doErr != nil {
		return 0, false, doErr
	}

	return n, done, outErr
}

func (sc *sysConn) Close() error                  { return sc.c.Close() }
func (sc *sysConn) SetDeadline(t time.Time) error { return sc.c.SetDeadline(t) }
