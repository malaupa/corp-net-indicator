package tray

import (
	"log"
	"os"
	"os/exec"
	"os/signal"

	"de.telekom-mms.corp-net-indicator/internal/assets"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
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
	// translations
	l := i18n.Localizer()
	// set up menu
	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
	t.statusItem = systray.AddMenuItem(l.Sprintf("Status"), l.Sprintf("Show Status"))
	t.statusItem.SetIcon(assets.GetIcon(assets.Status))
	t.actionItem = systray.AddMenuItem(l.Sprintf("Connect VPN"), l.Sprintf("Connect to VPN"))
	t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
	t.actionItem.Hide()
}

func (t *tray) OpenWindow(quickConnect bool) {
	t.closeWindow()
	self, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	var cmd *exec.Cmd
	if quickConnect {
		cmd = exec.Command(self, "-quick")
	} else {
		cmd = exec.Command(self)
	}

	t.closeChan = make(chan struct{})
	err = cmd.Start()
	go func() {
		cmd.Process.Wait()
		t.window = nil
		close(t.closeChan)
	}()
	if err != nil {
		log.Println(err)
	}
	t.window = cmd.Process
}

func (t *tray) closeWindow() {
	if t.window != nil {
		err := t.window.Signal(os.Interrupt)
		if err != nil {
			log.Println(err)
			err = t.window.Kill()
			if err != nil {
				log.Println(err)
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
	vStatus := vSer.GetStatus()
	iStatus := iSer.GetStatus()
	ctx := t.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = vStatus.InProgress
		ctx.IdentityInProgress = iStatus.InProgress
		ctx.TrustedNetwork = vStatus.TrustedNetwork
		ctx.Connected = vStatus.Connected
		ctx.LoggedIn = iStatus.LoggedIn
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
			t.OpenWindow(false)
		case <-t.actionItem.ClickedCh:
			if t.ctx.Read().Connected {
				t.actionItem.Disable()
				vSer.Disconnect()
			} else {
				t.OpenWindow(true)
			}
		// handle status updates
		case status := <-iChan:
			ctx := t.ctx.Write(func(ctx *model.ContextValues) {
				ctx.IdentityInProgress = status.InProgress
				ctx.LoggedIn = status.LoggedIn
			})
			t.apply(ctx)
		case status := <-vChan:
			ctx := t.ctx.Write(func(ctx *model.ContextValues) {
				ctx.VPNInProgress = status.InProgress
				ctx.TrustedNetwork = status.TrustedNetwork
				ctx.Connected = status.Connected
			})
			t.apply(ctx)
		case <-c:
			t.closeWindow()
			vSer.Close()
			iSer.Close()
			t.quitSystray()
			return
		}
	}
}

func (t *tray) apply(ctx model.ContextValues) {
	l := i18n.Localizer()
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
		t.actionItem.SetTitle(l.Sprintf("Disconnect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Disconnect))
		t.actionItem.Show()
	} else {
		t.actionItem.SetTitle(l.Sprintf("Connect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
		if ctx.TrustedNetwork {
			t.actionItem.Hide()
		} else {
			t.actionItem.Show()
		}
	}
}
