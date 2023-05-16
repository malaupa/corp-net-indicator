package service

import (
	"com.telekom-mms.corp-net-indicator/internal/logger"
	"github.com/godbus/dbus/v5"
)

const (
	IFACE  string = "org.freedesktop.login1.Session"
	SIGNAL string = IFACE + ".Unlock"
)

type Watcher struct {
	conn   *dbus.Conn
	signal chan struct{}
}

// Creates new watcher
func NewWatcher() *Watcher {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	return &Watcher{conn: conn, signal: make(chan struct{})}
}

// Listen to login events
func (w *Watcher) Listen() <-chan struct{} {
	logger.Verbose("Listen to user actions")
	// setup signal
	opts := []dbus.MatchOption{
		dbus.WithMatchInterface(IFACE),
	}
	err := w.conn.AddMatchSignal(opts...)
	if err != nil {
		panic(err)
	}

	// create signal channel
	c := make(chan *dbus.Signal, 10)
	w.conn.Signal(c)

	go func() {
		for sig := range c {
			if sig.Name == SIGNAL {
				select {
				case w.signal <- struct{}{}:
				default:
				}
			}
		}
	}()

	return w.signal
}

// Cleanup dbus connection and channel
func (w *Watcher) Close() {
	w.conn.Close()
	close(w.signal)
}
