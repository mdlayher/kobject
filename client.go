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
	"io"
)

// A Client provides access to Linux kobject userspace events.
type Client struct {
	rc io.ReadCloser
}

// New creates a new Client.
func New() (*Client, error) {
	// OS-specific (netlink) initialization.
	rc, err := newReadCloser()
	if err != nil {
		return nil, err
	}

	return newClient(rc)
}

// newClient is the internal constructor for a Client, used in tests.
func newClient(rc io.ReadCloser) (*Client, error) {
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
	b := make([]byte, 2048)
	n, err := c.rc.Read(b)
	if err != nil {
		return nil, err
	}

	// Fields are NULL-delimited.  Expect at least two fields, though the
	// first is ignored because it provides identical information to fields
	// which occur later on in the easy to parse KEY=VALUE format.
	fields := bytes.Split(b[:n], []byte{0x00})
	if len(fields) < 2 {
		return nil, io.ErrUnexpectedEOF
	}

	return parseEvent(fields[1:])
}
