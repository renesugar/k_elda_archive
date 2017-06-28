package connection

import (
	"net"
	"time"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
)

// The timeout for clients to connect to servers.
const connectTimeout = 5 * time.Second

// Client creates a grpc client connected to `addr`.
func Client(proto, addr string) (*grpc.ClientConn, error) {
	dialer := func(dialAddr string, t time.Duration) (net.Conn, error) {
		return net.DialTimeout(proto, dialAddr, t)
	}
	return grpc.Dial(addr, grpc.WithDialer(dialer), grpc.WithInsecure(),
		grpc.WithBlock(), grpc.WithTimeout(connectTimeout))
}

// Server creates a socket listening on `addr` and a grpc server. If it fails
// to open the socket, it tries again until it succeeds.
func Server(proto, addr string) (net.Listener, *grpc.Server) {
	for {
		sock, err := net.Listen(proto, addr)
		if err == nil {
			return sock, grpc.NewServer()
		}

		log.WithError(err).Error("Failed to open socket.")
		time.Sleep(30 * time.Second)
	}
}
