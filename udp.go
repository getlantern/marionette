package marionette

import "net"

type UDPClient struct {
	conn *net.UDPConn
}

type UDPServer struct {
	conn *net.UDPConn
}
