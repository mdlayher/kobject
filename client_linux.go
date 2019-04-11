//+build linux

package kobject

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// newConn dials out to kobject uevent netlink and returns a conn to use to
// listen for events.
func newConn(n int) (conn, error) {
	c, err := netlink.Dial(unix.NETLINK_KOBJECT_UEVENT, &netlink.Config{
		// TODO(mdlayher): replace with constant in x/sys/unix.
		Groups: 0x1,
	})
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	sc := &sysConn{
		c:      c,
		wg:     &wg,
		eventC: make(chan *Event, 8),
	}

	go func() {
		defer wg.Done()
		sc.readEvents(n)
	}()

	return sc, nil
}

var _ conn = &sysConn{}

// A sysConn implements conn over a *netlink.Conn.
type sysConn struct {
	c      *netlink.Conn
	wg     *sync.WaitGroup
	eventC chan *Event
}

func (sc *sysConn) Receive() (*Event, error) {
	return <-sc.eventC, nil
}

func (sc *sysConn) readEvents(num int) {
	b := make([]byte, num)
	var i int
	for {
		n, err := sc.read(b[i:])
		if err != nil {
			log.Println("READ ERR:", err)
			continue
		}

		log.Println("READ:", n, string(b))
		i += n

		if n == len(b) {

			log.Println("alloc")
			b = append(b, make([]byte, len(b))...)
			continue
		}

		log.Println("num:", bytes.Count(b, []byte("@")))

		i = 0

		// Look for 0x00action@

	}
}

func (sc *sysConn) read(b []byte) (int, error) {
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

func (sc *sysConn) Close() error {
	sc.wg.Wait()
	return sc.c.Close()
}
func (sc *sysConn) SetDeadline(t time.Time) error { return sc.c.SetDeadline(t) }

/*
	// Allocate a reasonable amount of space for a single Event.
	b := make([]byte, os.Getpagesize())

	for {
		// Attempt to read an Event using the buffer we have allocated.
		n, done, err := c.rc.TryRead(b)
		if err != nil {
			return nil, err
		}

		if !done {
			// The read couldn't complete because our buffer was too small;
			// double the size and try again.
			b = make([]byte, len(b)*2)
			continue
		}

		// We've completed reading, now parse the Event.

		// Fields are NULL-delimited.  Expect at least two fields, though the
		// first is ignored because it provides identical information to fields
		// which occur later on in the easy to parse KEY=VALUE format.
		fields := bytes.Split(b[:n], []byte{0x00})
		if len(fields) < 2 {
			return nil, io.ErrUnexpectedEOF
		}

		return parseEvent(fields[1:])
	}
*/
