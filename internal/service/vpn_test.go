package service

// func TestGetVPNStatus(t *testing.T) {
// 	s := testserver.NewVPNServer(false)
// 	defer s.Close()

// 	c := NewVPNService()
// 	defer c.Close()

// 	status, err := c.Query()
// 	assert.Nil(t, err)
// 	assert.Equal(t, &model.VPNStatus{
// 		TrustedNetwork:  test.Pointer(model.NotTrusted),
// 		ConnectionState: test.Pointer(model.ConnectUnknown),
// 		IP:              test.Pointer("127.0.0.1"),
// 		Device:          test.Pointer("vpn-tun0"),
// 		ConnectedAt:     test.Pointer(int64(0)),
// 		CertExpiresAt:   test.Pointer(int64(60 * 60 * 24 * 365)),
// 	}, status)
// }

// func TestGetVPNStatusError(t *testing.T) {
// 	s := testserver.NewVPNServer(false)
// 	c := NewVPNService()
// 	defer c.Close()

// 	s.Close()

// 	status, err := c.GetStatus()
// 	assert.EqualError(t, err, "The name com.telekom_mms.oc_daemon.Daemon was not provided by any .service files")
// 	assert.Nil(t, status)
// }

// type testClient struct {
// 	oc.DBusClient
// 	connectCalled    bool
// 	disconnectCalled bool
// }

// func (c *testClient) GetConfig() *oc.Config {
// 	return &oc.Config{}
// }

// func (c *testClient) Authenticate() error {
// 	return nil
// }

// func (c *testClient) GetLogin() *logininfo.LoginInfo {
// 	return &logininfo.LoginInfo{Cookie: "cookie", Host: "host", ConnectURL: "connectURL", Fingerprint: "fingerprint", Resolve: "resolve"}
// }

// func (c *testClient) Query() (*vpnstatus.Status, error) {
// 	return &vpnstatus.Status{
// 		TrustedNetwork:  vpnstatus.TrustedNetworkNotTrusted,
// 		ConnectionState: vpnstatus.ConnectionStateDisconnected,
// 		OCRunning:       vpnstatus.OCRunningNotRunning}, nil
// }

// func (c *testClient) Connect() error {
// 	c.connectCalled = true
// 	return nil
// }

// func (c *testClient) Disconnect() error {
// 	c.disconnectCalled = true
// 	return nil
// }

// func TestConnectAndDisconnect(t *testing.T) {
// 	assert := assert.New(t)
// 	s := testserver.NewVPNServer(false)
// 	defer s.Close()

// 	c := NewVPNService()
// 	c.ocClient = &testClient{}
// 	defer c.Close()

// 	ready := make(chan struct{})
// 	msgs := make(chan []model.VPNStatus, 1)
// 	connected := make(chan struct{})
// 	go func() {
// 		sC := c.ListenToVPN()
// 		var results []model.VPNStatus
// 		count := 0
// 		for status := range sC {
// 			count++
// 			if count == 1 {
// 				close(ready)
// 				continue
// 			}
// 			results = append(results, *status)
// 			if count == 5 {
// 				break
// 			}
// 			if count == 3 {
// 				close(connected)
// 			}
// 		}
// 		msgs <- results
// 	}()

// 	<-ready
// 	assert.Nil(c.Connect("pass", "server"))
// 	go func() {
// 		<-connected
// 		assert.Nil(c.Disconnect())
// 	}()

// 	results := <-msgs
// 	assert.Equal(4, len(results))
// 	assert.Equal(model.VPNStatus{
// 		ConnectionState: test.Pointer(model.Connecting),
// 	}, results[0])
// 	assert.Equal(model.VPNStatus{
// 		ConnectionState: test.Pointer(model.Connected),
// 		ConnectedAt:     test.Pointer(int64(0)),
// 	}, results[1])
// 	assert.Equal(model.VPNStatus{
// 		ConnectionState: test.Pointer(model.Disconnecting),
// 	}, results[2])
// 	assert.Equal(model.VPNStatus{
// 		ConnectionState: test.Pointer(model.Disconnected),
// 		ConnectedAt:     test.Pointer(int64(0)),
// 	}, results[3])
// }
