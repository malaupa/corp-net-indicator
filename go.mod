module de.telekom-mms.corp-net-indicator

go 1.20

require (
	github.com/diamondburned/gotk4/pkg v0.0.5
	github.com/slytomcat/systray v1.10.1
	golang.org/x/text v0.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/godbus/dbus/v5 v5.1.0
	github.com/stretchr/testify v1.8.2
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20230221090011-e4bae7ad2296 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.6.0 // indirect
)

replace github.com/godbus/dbus/v5 v5.1.0 => github.com/malaupa/dbus/v5 v5.2.0-next
