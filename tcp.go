package marionette

type TCPClient struct {
	conn *net.TCPConn
}

type TCPServer struct {
	ln *net.TCPListener
}
