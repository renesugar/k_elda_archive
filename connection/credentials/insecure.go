package credentials

import (
	"google.golang.org/grpc"
)

// Insecure defines connections with no authentication.
type Insecure struct{}

// ClientOpts returns the `DialOption`s necessary to setup an insecure client.
func (insecure Insecure) ClientOpts() []grpc.DialOption {
	return []grpc.DialOption{grpc.WithInsecure()}
}

// ServerOpts returns the `ServerOption`s necessary to setup an insecure server.
func (insecure Insecure) ServerOpts() []grpc.ServerOption {
	return nil
}
