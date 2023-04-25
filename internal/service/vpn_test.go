package service_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/model"
	testserver "de.telekom-mms.corp-net-indicator/internal/schema"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestGetVPNStatus(t *testing.T) {
	s := testserver.NewVPNServer(false)
	defer s.Close()

	c := service.NewVPNService()
	defer c.Close()

	status, err := c.GetStatus()
	assert.Nil(t, err)
	assert.Equal(t, &model.VPNStatus{
		TrustedNetwork:  test.Pointer(uint32(0)),
		ConnectionState: test.Pointer(uint32(0)),
		IP:              test.Pointer("127.0.0.1"),
		Device:          test.Pointer("vpn-tun0"),
		ConnectedAt:     test.Pointer(int64(0)),
		CertExpiresAt:   test.Pointer(int64(60 * 60 * 24 * 365)),
	}, status)
}

func TestGetVPNStatusError(t *testing.T) {
	s := testserver.NewVPNServer(false)
	c := service.NewVPNService()
	defer c.Close()

	s.Close()

	status, err := c.GetStatus()
	assert.EqualError(t, err, "The name de.telekomMMS.vpn was not provided by any .service files")
	assert.Nil(t, status)
}

func TestConnectAndDisconnect(t *testing.T) {
	assert := assert.New(t)
	s := testserver.NewVPNServer(false)
	defer s.Close()

	c := service.NewVPNService()
	defer c.Close()

	ready := make(chan struct{})
	msgs := make(chan []model.VPNStatus, 1)
	connected := make(chan struct{})
	go func() {
		sC := c.ListenToVPN()
		var results []model.VPNStatus
		count := 0
		close(ready)
		for status := range sC {
			count++
			results = append(results, *status)
			if count == 4 {
				break
			}
			if count == 2 {
				close(connected)
			}
		}
		msgs <- results
	}()

	<-ready
	assert.Nil(c.Connect("pass", "server"))
	go func() {
		<-connected
		assert.Nil(c.Disconnect())
	}()

	results := <-msgs
	assert.Equal(4, len(results))
	assert.Equal(model.VPNStatus{
		ConnectionState: test.Pointer(uint32(2)),
	}, results[0])
	assert.Equal(model.VPNStatus{
		ConnectionState: test.Pointer(uint32(3)),
		ConnectedAt:     test.Pointer(int64(0)),
	}, results[1])
	assert.Equal(model.VPNStatus{
		ConnectionState: test.Pointer(uint32(4)),
	}, results[2])
	assert.Equal(model.VPNStatus{
		ConnectionState: test.Pointer(uint32(1)),
		ConnectedAt:     test.Pointer(int64(0)),
	}, results[3])
}
