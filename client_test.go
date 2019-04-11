package kobject

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// page is a byte slice large enough that any additional data will trigger
// multiple calls to tryReadCloser.TryRead.
var page = bytes.Repeat([]byte{'f'}, os.Getpagesize())

func TestClientReceive(t *testing.T) {

	tests := []struct {
		name  string
		b     []byte
		calls int
		e     *Event
	}{
		{
			name:  "empty",
			calls: 1,
		},
		{
			name:  "header",
			b:     []byte("add@/devices/test"),
			calls: 1,
		},
		{
			name: "exactly one page",
			// Verify we behave as expected and read once, even if our buffer
			// is entirely full.
			calls: 1,
			b:     page,
		},
		{
			name:  "no values",
			b:     []byte("add@/devices/test\x00ACTION=add\x00DEVPATH=/devices/test\x00SUBSYSTEM=test\x00SEQNUM=1"),
			calls: 1,
			e: &Event{
				Action:     Add,
				DevicePath: "/devices/test",
				Subsystem:  "test",
				Sequence:   1,
				Values:     map[string]string{},
			},
		},
		{
			name:  "USB device",
			b:     []byte("add@/devices/pci0000:00/0000:00:14.0/usb3/3-2/3-2:1.0/0003:046D:C52B.0026\x00ACTION=add\x00DEVPATH=/devices/pci0000:00/0000:00:14.0/usb3/3-2/3-2:1.0/0003:046D:C52B.0026\x00SUBSYSTEM=hid\x00SEQNUM=4618\x00HID_UNIQ=\x00MODALIAS=hid:b0003g0000v0000046Dp0000C52B\x00HID_ID=0003:0000046D:0000C52B\x00HID_NAME=Logitech USB Receiver\x00HID_PHYS=usb-0000:00:14.0-2/input0"),
			calls: 1,
			e: &Event{
				Action:     Add,
				DevicePath: "/devices/pci0000:00/0000:00:14.0/usb3/3-2/3-2:1.0/0003:046D:C52B.0026",
				Subsystem:  "hid",
				Sequence:   4618,
				Values: map[string]string{
					"HID_ID":   "0003:0000046D:0000C52B",
					"HID_NAME": "Logitech USB Receiver",
					"HID_PHYS": "usb-0000:00:14.0-2/input0",
					"HID_UNIQ": "",
					"MODALIAS": "hid:b0003g0000v0000046Dp0000C52B",
				},
			},
		},
		{
			name:  "TAP interface",
			b:     []byte("remove@/devices/virtual/net/tap0\x00ACTION=remove\x00DEVPATH=/devices/virtual/net/tap0\x00SUBSYSTEM=net\x00SEQNUM=4636\x00INTERFACE=tap0\x00IFINDEX=28"),
			calls: 1,
			e: &Event{
				Action:     Remove,
				DevicePath: "/devices/virtual/net/tap0",
				Subsystem:  "net",
				Sequence:   4636,
				Values: map[string]string{
					"IFINDEX":   "28",
					"INTERFACE": "tap0",
				},
			},
		},
		{
			name: "large event",
			b: append(
				[]byte("remove@/devices/virtual/net/tap0\x00ACTION=remove\x00DEVPATH=/devices/virtual/net/tap0\x00SUBSYSTEM=net\x00SEQNUM=4636\x00INTERFACE=tap0\x00IFINDEX=28\x00ARBITRARY="),
				// Ensure this message is too large to fit in one page of memory,
				// triggering multiple TryRead calls.
				append(page, 0x00)...,
			),
			calls: 2,
			e: &Event{
				Action:     Remove,
				DevicePath: "/devices/virtual/net/tap0",
				Subsystem:  "net",
				Sequence:   4636,
				Values: map[string]string{
					"IFINDEX":   "28",
					"INTERFACE": "tap0",
					"ARBITRARY": string(page),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, done := testClient(t, tt.b)
			defer c.Close()

			e, err := c.Receive()

			if err != nil && tt.e != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.e == nil {
				t.Fatal("expected an error, but none occurred")
			}

			// Check the number of calls to TryRead.
			calls := done()
			if diff := cmp.Diff(tt.calls, calls); diff != "" {
				t.Fatalf("unexpected number of TryRead calls (-want +got):\n%s", diff)
			}

			// Check the actual Event produced, if any.
			if diff := cmp.Diff(tt.e, e); diff != "" {
				t.Fatalf("unexpected event (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClientReceiveConcurrent(t *testing.T) {
	b := append(
		[]byte("remove@/devices/virtual/net/tap0\x00ACTION=remove\x00DEVPATH=/devices/virtual/net/tap0\x00SUBSYSTEM=net\x00SEQNUM=4636\x00INTERFACE=tap0\x00IFINDEX=28\x00ARBITRARY="),
		// Ensure this message is too large to fit in one page of memory,
		// triggering multiple TryRead calls.
		append(page, 0x00)...,
	)

	want := &Event{
		Action:     Remove,
		DevicePath: "/devices/virtual/net/tap0",
		Subsystem:  "net",
		Sequence:   4636,
		Values: map[string]string{
			"IFINDEX":   "28",
			"INTERFACE": "tap0",
			"ARBITRARY": string(page),
		},
	}

	c, done := testClient(t, b)

	execN := func(n int) {
		for i := 0; i < n; i++ {
			e, err := c.Receive()
			if err != nil {
				panicf("failed to receive: %v", err)
			}

			if diff := cmp.Diff(want, e); diff != "" {
				panicf("unexpected event (-want +got):\n%s", diff)
			}
		}
	}

	const (
		workers    = 4
		iterations = 10000
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			execN(iterations)
		}()
	}

	wg.Wait()

	// Expect the buffer to be grown exactly once, because the event is larger
	// than one page of memory.
	size := workers*iterations + 1
	if diff := cmp.Diff(size, done()); diff != "" {
		t.Fatalf("unexpected number of reads (-want +got):\n%s", diff)
	}
}

func testClient(t *testing.T, b []byte) (*Client, func() int) {
	t.Helper()

	rc := &testTryReadCloser{
		b: b,
	}

	c, err := newClient(rc)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	return c, func() int {
		if err := c.Close(); err != nil {
			panicf("failed to close: %v", err)
		}

		// Return the number of times TryRead was called upon completion.
		return rc.calls
	}
}

type testTryReadCloser struct {
	mu    sync.Mutex
	b     []byte
	calls int
}

func (rc *testTryReadCloser) TryRead(b []byte) (int, bool, error) {
	rc.mu.Lock()
	rc.calls++
	rc.mu.Unlock()

	// Is b large enough to fit the contents of the bytes in one call?
	if len(b) < len(rc.b) {
		// No, indicate the caller should try again.
		return 0, false, nil
	}

	// Yes, proceed with Read. Create a new bytes.Reader on each invocation
	// so that the same data can be read multiple times.
	n, err := bytes.NewReader(rc.b).Read(b)
	return n, true, err
}

func (*testTryReadCloser) Close() error { return nil }
