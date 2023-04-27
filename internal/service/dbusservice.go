package service

import (
	"github.com/godbus/dbus/v5"
)

type dbusService struct {
	conn *dbus.Conn

	iface string
	path  string
}

func (d *dbusService) listen(onSignal func(sig map[string]dbus.Variant)) {
	opts := []dbus.MatchOption{
		dbus.WithMatchSender(d.iface),
		dbus.WithMatchObjectPath(dbus.ObjectPath(d.path)),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	}
	err := d.conn.AddMatchSignal(opts...)
	if err != nil {
		panic(err)
	}
	go func() {
		// should not be necessary, is removed on disconnect to dbus
		// defer d.conn.RemoveMatchSignal(opts...)

		c := make(chan *dbus.Signal, 1)
		d.conn.Signal(c)

		for sig := range c {
			// make sure it's a properties changed signal
			if string(sig.Path) != d.path || sig.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
				continue
			}

			// check properties changed signal
			if v, ok := sig.Body[0].(string); !ok || v != d.iface {
				continue
			}

			// get changed properties
			changed, ok := sig.Body[1].(map[string]dbus.Variant)
			if !ok {
				continue
			}

			onSignal(changed)
		}
	}()
}

func (d *dbusService) getStatus() (map[string]dbus.Variant, error) {
	var status map[string]dbus.Variant
	err := d.getObject().Call("org.freedesktop.DBus.Properties.GetAll", 0, d.iface).Store(&status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (d *dbusService) getProp(name string) (result dbus.Variant, err error) {
	err = d.getObject().Call("org.freedesktop.DBus.Properties.Get", 0, d.iface, name).Store(&result)
	return
}

func (d *dbusService) callMethod(method string, args ...any) *dbus.Call {
	return d.getObject().Call(d.iface+"."+method, 0, args...)
}

func (d *dbusService) getObject() dbus.BusObject {
	return d.conn.Object(d.iface, dbus.ObjectPath(d.path))
}
