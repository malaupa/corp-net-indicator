package testserver

import (
	"sync"

	"github.com/godbus/dbus/v5"
)

type state struct {
	sync.Mutex
	state    map[string]dbus.Variant
	conn     *dbus.Conn
	simulate bool
}

func (s *state) getState() map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	return s.state
}

func (s *state) setInProgress() map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	s.state["InProgress"] = dbus.MakeVariant(true)
	return s.state
}
