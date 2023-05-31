package service_test

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/model"
	testserver "com.telekom-mms.corp-net-indicator/internal/schema"
	"com.telekom-mms.corp-net-indicator/internal/service"
	"com.telekom-mms.corp-net-indicator/internal/test"
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
		TrustedNetwork:       test.Pointer(model.TrustUnknown),
		LoginState:           test.Pointer(model.LoginUnknown),
		LastKeepAliveAt:      test.Pointer(int64(60 * 60)),
		KerberosTGTStartTime: test.Pointer(int64(0)),
		KerberosTGTEndTime:   test.Pointer(int64(60 * 60)),
	}, status)
}

func TestGetStatusError(t *testing.T) {
	s := testserver.NewIdentityServer(false)
	c := service.NewIdentityService()
	defer c.Close()

	s.Close()

	status, err := c.GetStatus()
	assert.EqualError(t, err, "The name com.telekom_mms.fw_id_agent.Agent was not provided by any .service files")
	assert.Nil(t, status)
}

func TestReLogin(t *testing.T) {
	s := testserver.NewIdentityServer(false)
	defer s.Close()

	c := service.NewIdentityService()
	defer c.Close()

	ready := make(chan struct{})
	msgs := make(chan []model.IdentityStatus, 1)
	go func() {
		sC := c.ListenToIdentity()
		var results []model.IdentityStatus
		count := 0
		for status := range sC {
			count++
			if count == 1 {
				close(ready)
				continue
			}
			results = append(results, *status)
			if count == 3 {
				break
			}
		}
		msgs <- results
	}()
	<-ready
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
