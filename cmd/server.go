package main

import testserver "com.telekom-mms.corp-net-indicator/internal/schema"

func main() {
	iS := testserver.NewIdentityServer(true)
	defer iS.Close()
	vS := testserver.NewVPNServer(true)
	defer vS.Close()

	select {}
}
