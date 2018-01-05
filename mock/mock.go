package mock

import (
	"context"
	"net"
)

type Dialer struct {
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.DialContextFunc(ctx, network, address)
}
