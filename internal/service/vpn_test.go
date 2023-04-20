package service_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/testserver"
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
		TrustedNetwork: false,
		InProgress:     false,
		Connected:      false,
		IP:             "127.0.0.1",
		Device:         "vpn-tun0",
		ConnectedAt:    0,
		CertExpiresAt:  60 * 60 * 24 * 365,
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

	msgs := make(chan []model.VPNStatus, 1)
	go func() {
		sC := c.ListenToVPN()
		var results []model.VPNStatus
		for status := range sC {
			results = append(results, *status)
		}
		msgs <- results
	}()
	assert.Nil(c.Connect("pass", "server"))
	go func() {
		assert.Nil(c.Disconnect())
		c.Close()
	}()

	results := <-msgs
	assert.Equal(4, len(results))
	assert.Equal(model.VPNStatus{
		TrustedNetwork: false,
		InProgress:     true,
		Connected:      false,
		IP:             "127.0.0.1",
		Device:         "vpn-tun0",
		ConnectedAt:    0,
		CertExpiresAt:  60 * 60 * 24 * 365,
	}, results[0])
	assert.Equal(model.VPNStatus{
		TrustedNetwork: false,
		InProgress:     false,
		Connected:      true,
		IP:             "127.0.0.1",
		Device:         "vpn-tun0",
		ConnectedAt:    0,
		CertExpiresAt:  60 * 60 * 24 * 365,
	}, results[1])
	assert.Equal(model.VPNStatus{
		TrustedNetwork: false,
		InProgress:     true,
		Connected:      true,
		IP:             "127.0.0.1",
		Device:         "vpn-tun0",
		ConnectedAt:    0,
		CertExpiresAt:  60 * 60 * 24 * 365,
	}, results[2])
	assert.Equal(model.VPNStatus{
		TrustedNetwork: false,
		InProgress:     false,
		Connected:      false,
		IP:             "127.0.0.1",
		Device:         "vpn-tun0",
		ConnectedAt:    0,
		CertExpiresAt:  60 * 60 * 24 * 365,
	}, results[3])
}