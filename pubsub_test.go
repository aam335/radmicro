package main

import (
	"testing"

	server "github.com/nats-io/nats-server/server"
	natsserver "github.com/nats-io/nats-server/test"
)

const TEST_PORT = 8369

func RunServerOnPort(port int) *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(&opts)
}

func RunServerWithOptions(opts *server.Options) *server.Server {
	return natsserver.RunServer(opts)
}

func TestPubSub_Fire(t *testing.T) {
	s := RunServerOnPort(TEST_PORT)
	defer s.Shutdown()

	// ps, err := NewPubSub()

}
