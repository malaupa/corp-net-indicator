package service_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/model"
	testserver "de.telekom-mms.corp-net-indicator/internal/schema"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/test"
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
		TrustedNetwork:   test.Pointer(model.TrustUnknown),
		LoginState:       test.Pointer(model.LoginUnknown),
		LastKeepAliveAt:  test.Pointer(int64(60 * 60)),
		KerberosIssuedAt: test.Pointer(int64(0)),
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
	defer c.Close()

	msgs := make(chan []model.IdentityStatus, 1)
	go func() {
		sC := c.ListenToIdentity()
		var results []model.IdentityStatus
		count := 0
		for status := range sC {
			count++
			results = append(results, *status)
			if count == 2 {
				break
			}
		}
		msgs <- results
	}()
	assert.Nil(t, c.ReLogin())

	results := <-msgs
	assert.Equal(t, 2, len(results))
	assert.Equal(t, model.IdentityStatus{
		LoginState: test.Pointer(model.LoggingIn),
	}, results[0])
	assert.Equal(t, model.IdentityStatus{
		LoginState:      test.Pointer(model.LoggedIn),
		LastKeepAliveAt: test.Pointer(int64(60 * 60)),
	}, results[1])
}
