kobject [![Build Status](https://travis-ci.org/mdlayher/kobject.svg?branch=master)](https://travis-ci.org/mdlayher/kobject) [![GoDoc](https://godoc.org/github.com/mdlayher/kobject?status.svg)](https://godoc.org/github.com/mdlayher/kobject) [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/kobject)](https://goreportcard.com/report/github.com/mdlayher/kobject)
=======

Package `kobject` provides access to Linux kobject userspace events.

Userspace events occur whenever a kobject's state changes.  As an example,
events are triggered whenever a USB device is added or removed from a system,
or whenever a virtual network interface is added or removed.

For more information on kobjects, please see:
  - https://www.kernel.org/doc/Documentation/kobject.txt

MIT Licensed.