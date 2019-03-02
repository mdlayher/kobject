package kobject

import (
	"strconv"
	"strings"
)

// An Action is an action which caused an Event to be triggered.
type Action string

// Possible Actions which trigger an Event.
const (
	ActionAdd     Action = "add"
	ActionBind    Action = "bind"
	ActionRemove  Action = "remove"
	ActionChange  Action = "change"
	ActionMove    Action = "move"
	ActionOnline  Action = "online"
	ActionOffline Action = "offline"
	ActionUnbind  Action = "unbind"
)

// An Event is a userspace event in response to a state change of a kobject.
type Event struct {
	// Fields which are present in all events.
	Action     Action
	DevicePath string
	Subsystem  string
	Sequence   int

	// Values contains arbitrary key/value pairs which are not present in
	// all Events.
	Values map[string]string
}

// parseEvent parses an Event from a series of KEY=VALUE pairs.
func parseEvent(fields [][]byte) (*Event, error) {
	e := &Event{
		Values: make(map[string]string),
	}

	for f := range fields {
		// Assume all information is in KEY=VALUE pairs.
		kv := strings.Split(string(fields[f]), "=")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "ACTION":
			e.Action = Action(kv[1])
		case "DEVPATH":
			e.DevicePath = kv[1]
		case "SUBSYSTEM":
			e.Subsystem = kv[1]
		case "SEQNUM":
			v, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, err
			}

			e.Sequence = v
		default:
			e.Values[kv[0]] = kv[1]
		}
	}

	return e, nil
}
