package main

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func RunNatsInstance() (*server.Server, string) {
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           server.RANDOM_PORT,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 2048,
	}
	s, err := server.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(5 * time.Second) {
		log.Fatal("Unable to start NATS Server in Go Routine")
	}
	uri := s.Addr().String()
	return s, uri
}

func TestPubSub_Fire(t *testing.T) {
	s, serverURI := RunNatsInstance()
	defer s.Shutdown()

	prefix := "test.prefix"
	bgCh := make(chan *nats.Msg)

	ns, err := nats.Connect(serverURI)
	require.NoError(t, err)
	_, err = ns.ChanSubscribe(">", bgCh)
	require.NoError(t, err)

	testPairs := map[string]string{"a": "sa", "b": "sb"}

	tests := []struct {
		topic string
		pairs map[string]string
		subj  string
	}{
		{topic: "test", pairs: testPairs, subj: prefix + "." + "test"},
		{topic: "test", subj: prefix + "." + "test"},
		{subj: prefix + "."},
	}

	ps, err := NewPubSub(prefix, serverURI)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, test := range tests {
		ps.Fire(test.topic, test.pairs)
		select {
		case <-ctx.Done():
			t.Fatal("timeout")
		case ret := <-bgCh:
			require.Equal(t, test.subj, ret.Subject)
			data := map[string]string{}
			require.NoError(t, json.Unmarshal(ret.Data, &data))
			require.Equal(t, data, test.pairs)
		}
	}
}

//
func replyBackend(serverURI string) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		var (
			nc  *nats.Conn
			err error
		)
		if nc, err = nats.Connect(serverURI); err == nil {
			_, err = nc.Subscribe(">", func(m *nats.Msg) {
				m.Respond(m.Data)
			})
		}
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestPubSub_Req(t *testing.T) {
	s, serverURI := RunNatsInstance()
	defer s.Shutdown()
	// serverURI := "localhost:4222"
	ns, err := nats.Connect(serverURI)
	require.NoError(t, err)

	replyBackend(serverURI)

	type fields struct {
		ns     *nats.Conn
		prefix string
	}
	type args struct {
		ctx   context.Context
		topic string
		key   string
		pairs map[string]string
	}
	testPairs := map[string]string{"a": "sa", "b": "sb"}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "no server", fields: fields{prefix: "testprefix."}, args: args{ctx: context.Background(), topic: "test", key: "key", pairs: testPairs}, wantErr: true},
		{name: "1st", fields: fields{prefix: "testprefix.", ns: ns}, args: args{ctx: nil, topic: "test", key: "key", pairs: testPairs}, wantRet: testPairs, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &pubsub{
				ns:     tt.fields.ns,
				prefix: tt.fields.prefix,
			}
			if tt.args.ctx == nil {
				var cancel context.CancelFunc
				tt.args.ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()

			}
			gotRet, err := ps.Req(tt.args.ctx, tt.args.topic, tt.args.key, tt.args.pairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("PubSub.Req() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("PubSub.Req() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
