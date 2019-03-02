package kobject

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClientReceive(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		e    *Event
	}{
		{
			name: "empty",
		},
		{
			name: "header",
			b:    []byte("add@/devices/test"),
		},
		{
			name: "no values",
			b:    []byte("add@/devices/test\x00ACTION=add\x00DEVPATH=/devices/test\x00SUBSYSTEM=test\x00SEQNUM=1"),
			e: &Event{
				Action:     Add,
				DevicePath: "/devices/test",
				Subsystem:  "test",
				Sequence:   1,
				Values:     map[string]string{},
			},
		},
		{
			name: "USB device",
			b:    []byte("add@/devices/pci0000:00/0000:00:14.0/usb3/3-2/3-2:1.0/0003:046D:C52B.0026\x00ACTION=add\x00DEVPATH=/devices/pci0000:00/0000:00:14.0/usb3/3-2/3-2:1.0/0003:046D:C52B.0026\x00SUBSYSTEM=hid\x00SEQNUM=4618\x00HID_UNIQ=\x00MODALIAS=hid:b0003g0000v0000046Dp0000C52B\x00HID_ID=0003:0000046D:0000C52B\x00HID_NAME=Logitech USB Receiver\x00HID_PHYS=usb-0000:00:14.0-2/input0"),
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
			name: "TAP interface",
			b:    []byte("remove@/devices/virtual/net/tap0\x00ACTION=remove\x00DEVPATH=/devices/virtual/net/tap0\x00SUBSYSTEM=net\x00SEQNUM=4636\x00INTERFACE=tap0\x00IFINDEX=28"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t, tt.b)
			defer c.Close()

			e, err := c.Receive()

			if err != nil && tt.e != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.e == nil {
				t.Fatal("expected an error, but none occurred")
			}

			if diff := cmp.Diff(tt.e, e); diff != "" {
				t.Fatalf("unexpected event (-want +got):\n%s", diff)
			}
		})
	}
}

func testClient(t *testing.T, b []byte) *Client {
	t.Helper()

	c, err := newClient(ioutil.NopCloser(bytes.NewReader(b)))
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	return c
}
