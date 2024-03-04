// Package kobject provides access to Linux kobject userspace events.
//
// Userspace events occur whenever a kobject's state changes.  As an example,
// events are triggered whenever a USB device is added or removed from a system,
// or whenever a virtual network interface is added or removed.
//
// For more information on kobjects, please see:
//   - https://www.kernel.org/doc/Documentation/kobject.txt
package kobject

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// A Client provides access to Linux kobject userspace events. Clients are safe
// for concurrent use.
type Client struct {
	// To read events and close the handle, a tryReadCloser contains the
	// minimum required functionality. rc must also implement conn (e.g. as the
	// Linux sysConn type) for deadline support.
	rc tryReadCloser

	// A buffer used to read kobject uevents directly from the socket, before
	// they are parsed into Events.
	mu sync.Mutex
	b  []byte
}

// New creates a new Client.
func New() (*Client, error) {
	// OS-specific (netlink) initialization.
	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	return newClient(conn)
}

// newClient is the internal constructor for a Client, used in tests.
func newClient(rc tryReadCloser) (*Client, error) {
	return &Client{
		rc: rc,

		// Allocate a reasonable amount of space for a single Event, but the
		// buffer can also be grown if an event arrives and is too large to fit
		// in the buffer.
		b: make([]byte, os.Getpagesize()),
	}, nil
}

// Close releases resources used by a Client.
func (c *Client) Close() error {
	return c.rc.Close()
}

// SetReceiveBuffer sets the size of the buffer used to read kobject uevents.
func (c *Client) SetReceiveBuffer(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.b = make([]byte, size)
}

// Receive waits until a kobject userspace event is triggered, and then returns
// the Event.
func (c *Client) Receive() (*Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		// Attempt to read an Event using the buffer we have allocated.
		n, done, err := c.rc.TryRead(c.b)
		if err != nil {
			return nil, err
		}

		if !done {
			// The read couldn't complete because our buffer was too small;
			// double the size and try again.
			c.b = append(c.b, make([]byte, len(c.b))...)
			continue
		}

		// We've completed reading, now parse the Event.

		// Fields are NULL-delimited.  Expect at least two fields, though the
		// first is ignored because it provides identical information to fields
		// which occur later on in the easy to parse KEY=VALUE format.
		fields := bytes.Split(c.b[:n], []byte{0x00})
		if len(fields) < 2 {
			return nil, io.ErrUnexpectedEOF
		}

		return parseEvent(fields[1:])
	}
}

// SetDeadline sets the read deadline associated with the connection.
func (c *Client) SetDeadline(t time.Time) error {
	conn, ok := c.rc.(conn)
	if !ok {
		panicf("kobject: BUG: deadlines not supported on internal conn type: %#v", c.rc)
	}

	return conn.SetDeadline(t)
}

// SetReadBuffer sets the read buffer size associated with the connection.
//
// When the buffer is too small, kernel events are dropped with ENOBUFS error.
func (c *Client) SetReadBuffer(bytes int) error {
	conn, ok := c.rc.(conn)
	if !ok {
		panicf("kobject: BUG: read buffer size not supported on internal conn type: %#v", c.rc)
	}

	return conn.SetReadBuffer(bytes)
}

// A conn is the full set of required functionality for an internal type to
// expose via Client.
type conn interface {
	tryReadCloser
	SetDeadline(t time.Time) error
	SetReadBuffer(bytes int) error
}

type tryReadCloser interface {
	io.Closer

	// TryRead attempts to read from an io.Reader and reports whether it was
	// able to make any progress. If done is false, no bytes were read because
	// the buffer b was too small. If done is true, n indicates the number of
	// bytes read from the underlying io.Reader.
	TryRead(b []byte) (n int, done bool, err error)
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
