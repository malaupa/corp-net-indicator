package main

import testserver "de.telekom-mms.corp-net-indicator/internal/testserver"

func main() {
	iS := testserver.NewIdentityServer(true)
	defer iS.Close()
	vS := testserver.NewVPNServer(true)
	defer vS.Close()

	select {}
}
