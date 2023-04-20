package service_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/testserver"
	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	s := testserver.NewIdentityServer(false)
	defer s.Close()

	c := service.NewIdentityService()
	defer c.Close()

	status, err := c.GetStatus()
	assert.Nil(t, err)
	assert.Equal(t, &model.IdentityStatus{
		TrustedNetwork:  true,
		LoggedIn:        false,
		LastKeepAliveAt: 60 * 60,
		KrbIssuedAt:     0,
		InProgress:      false,
	}, status)
}

func TestGetStatusError(t *testing.T) {
	s := testserver.NewIdentityServer(false)
	c := service.NewIdentityService()
	defer c.Close()

	s.Close()

	status, err := c.GetStatus()
	assert.EqualError(t, err, "The name de.telekomMMS.identity was not provided by any .service files")
	assert.Nil(t, status)
}

func TestReLogin(t *testing.T) {
	s := testserver.NewIdentityServer(false)
	defer s.Close()

	c := service.NewIdentityService()

	msgs := make(chan []model.IdentityStatus, 1)
	go func() {
		sC := c.ListenToIdentity()
		var results []model.IdentityStatus
		for status := range sC {
			results = append(results, *status)
		}
		msgs <- results
	}()
	assert.Nil(t, c.ReLogin())
	c.Close()

	results := <-msgs
	assert.Equal(t, model.IdentityStatus{
		TrustedNetwork:  true,
		LoggedIn:        false,
		LastKeepAliveAt: 60 * 60,
		KrbIssuedAt:     0,
		InProgress:      true,
	}, results[0])
	assert.Equal(t, model.IdentityStatus{
		TrustedNetwork:  true,
		LoggedIn:        true,
		LastKeepAliveAt: 60 * 60,
		KrbIssuedAt:     0,
		InProgress:      false,
	}, results[1])
}
