package tray

import (
	"os"
	"os/exec"
	"os/signal"

	"de.telekom-mms.corp-net-indicator/internal/assets"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"github.com/slytomcat/systray"
)

type tray struct {
	ctx *model.Context

	statusItem   *systray.MenuItem
	actionItem   *systray.MenuItem
	startSystray func()
	quitSystray  func()

	window    *os.Process
	closeChan chan struct{}
}

// starts tray
func New() *tray {
	t := &tray{ctx: model.NewContext()}
	// create tray
	t.startSystray, t.quitSystray = systray.RunWithExternalLoop(t.onReady, func() {})
	return t
}

// init tray
func (t *tray) onReady() {
	// set up menu
	t.statusItem = systray.AddMenuItem(i18n.L.Sprintf("Status"), i18n.L.Sprintf("Show Status"))
	t.statusItem.SetIcon(assets.GetIcon(assets.Status))
	t.actionItem = systray.AddMenuItem(i18n.L.Sprintf("Connect VPN"), i18n.L.Sprintf("Connect to VPN"))
	t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
	t.actionItem.Hide()
}

func (t *tray) OpenWindow(quickConnect bool) {
	t.closeWindow()
	self, err := os.Executable()
	if err != nil {
		logger.Verbose(err)
		return
	}
	var cmd *exec.Cmd
	args := []string{}
	if quickConnect {
		args = append(args, "-quick")
	}
	if logger.IsVerbose {
		args = append(args, "-v")
	}
	cmd = exec.Command(self, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	t.closeChan = make(chan struct{})
	err = cmd.Start()
	go func() {
		_, err := cmd.Process.Wait()
		logger.Verbose("Waited for closing window")

		if err != nil {
			logger.Verbose(err)
		}
		t.window = nil
		close(t.closeChan)
	}()
	if err != nil {
		logger.Verbose(err)
	}
	t.window = cmd.Process
}

func (t *tray) closeWindow() {
	if t.window != nil {
		err := t.window.Signal(os.Interrupt)
		if err != nil {
			logger.Verbosef("SIGINT not working: %v\n", err)
			err = t.window.Kill()
			if err != nil {
				logger.Verbosef("SIGKILL not working: %v\n", err)
			}
		}
		<-t.closeChan
	}
}

func (t *tray) Run() {
	// start tray
	t.startSystray()
	// create services
	vSer := service.NewVPNService()
	iSer := service.NewIdentityService()
	// update tray
	vStatus, err := vSer.GetStatus()
	if err != nil {
		logger.Logf("DBUS error: %v\n", err)
		os.Exit(1)
	}
	iStatus, err := iSer.GetStatus()
	if err != nil {
		logger.Logf("DBUS error: %v\n", err)
		os.Exit(1)
	}
	ctx := t.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = vStatus.InProgress(ctx.VPNInProgress)
		ctx.Connected = vStatus.IsConnected(ctx.Connected)
		ctx.TrustedNetwork = vStatus.IsTrustedNetwork(ctx.TrustedNetwork)
		ctx.IdentityInProgress = iStatus.InProgress(ctx.IdentityInProgress)
		ctx.LoggedIn = iStatus.IsLoggedIn(ctx.LoggedIn)
	})
	t.apply(ctx)

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	// catch interrupt and clean up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// main loop
	for {
		select {
		// handle tray menu clicks
		case <-t.statusItem.ClickedCh:
			logger.Verbose("Open window to connect")

			t.OpenWindow(false)
		case <-t.actionItem.ClickedCh:
			if t.ctx.Read().Connected {
				logger.Verbose("Try to disconnect")

				t.actionItem.Disable()
				err := vSer.Disconnect()
				if err != nil {
					logger.Logf("DBUS error: %v\n", err)

					t.actionItem.Enable()
				}
			} else {
				logger.Verbose("Open window to quick connect")

				t.OpenWindow(true)
			}
		// handle status updates
		case status := <-iChan:
			logger.Verbosef("Apply identity status: %+v\n", status)

			ctx := t.ctx.Write(func(ctx *model.ContextValues) {
				ctx.IdentityInProgress = status.InProgress(ctx.IdentityInProgress)
				ctx.LoggedIn = status.IsLoggedIn(ctx.LoggedIn)
			})
			t.apply(ctx)
		case status := <-vChan:
			logger.Verbosef("Apply vpn status: %+v\n", status)

			ctx := t.ctx.Write(func(ctx *model.ContextValues) {
				ctx.VPNInProgress = status.InProgress(ctx.VPNInProgress)
				ctx.Connected = status.IsConnected(ctx.Connected)
				ctx.TrustedNetwork = status.IsTrustedNetwork(ctx.TrustedNetwork)
			})
			t.apply(ctx)
		case <-c:
			logger.Verbose("Received SIGINT -> closing")

			t.closeWindow()
			vSer.Close()
			iSer.Close()
			t.quitSystray()
			return
		}
	}
}

func (t *tray) apply(ctx model.ContextValues) {
	// icon
	if ctx.LoggedIn && (ctx.Connected || ctx.TrustedNetwork) {
		systray.SetIcon(assets.GetIcon(assets.Umbrella))
	} else {
		if ctx.Connected || ctx.TrustedNetwork {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		}
	}
	// action menu item
	if ctx.VPNInProgress {
		t.actionItem.Disable()
	} else {
		t.actionItem.Enable()
	}
	if ctx.Connected {
		t.actionItem.SetTitle(i18n.L.Sprintf("Disconnect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Disconnect))
		t.actionItem.Show()
	} else {
		t.actionItem.SetTitle(i18n.L.Sprintf("Connect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
		if ctx.TrustedNetwork {
			t.actionItem.Hide()
		} else {
			t.actionItem.Show()
		}
	}
}
