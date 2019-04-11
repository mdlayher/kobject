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
	"fmt"
	"io"
	"time"
)

// A Client provides access to Linux kobject userspace events.
type Client struct {
	// To read events and close the handle, a tryReadCloser contains the
	// minimum required functionality. rc must also implement conn (e.g. as the
	// Linux sysConn type) for deadline support.
	rc tryReadCloser
}

// New creates a new Client.
func New(n int) (*Client, error) {
	// OS-specific (netlink) initialization.
	conn, err := newConn(n)
	if err != nil {
		return nil, err
	}

	return newClient(conn)
}

// newClient is the internal constructor for a Client, used in tests.
func newClient(rc tryReadCloser) (*Client, error) {
	return &Client{
		rc: rc,
	}, nil
}

// Close releases resources used by a Client.
func (c *Client) Close() error {
	return c.rc.Close()
}

// Receive waits until a kobject userspace event is triggered, and then returns
// the Event.
func (c *Client) Receive() (*Event, error) {
	return c.rc.Receive()
}

// SetDeadline sets the read deadline associated with the connection.
func (c *Client) SetDeadline(t time.Time) error {
	conn, ok := c.rc.(conn)
	if !ok {
		panicf("kobject: BUG: deadlines not supported on internal conn type: %#v", c.rc)
	}

	return conn.SetDeadline(t)
}

// A conn is the full set of required functionality for an internal type to
// expose via Client.
type conn interface {
	tryReadCloser
	SetDeadline(t time.Time) error
}

type tryReadCloser interface {
	io.Closer

	Receive() (*Event, error)
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
