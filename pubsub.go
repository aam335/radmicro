package main

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

// PubSub ...
type PubSub interface {
	Fire(topic string, pairs map[string]string) error
	Req(ctx context.Context, topic string, key string, pairs map[string]string) (ret map[string]string, err error)
}

// pubsub sends topics with attributes to backend
// and receives replies, if needed
type pubsub struct {
	ns     *nats.Conn
	prefix string
}

// NewPubSub creates new PubSub
func NewPubSub(prefix string, natsURI string) (PubSub, error) {
	ns, err := nats.Connect(natsURI)
	if err != nil {
		return nil, err
	}
	ps := pubsub{ns: ns, prefix: prefix + "."}
	return ps, nil
}

// Fire Pubs json encoded pairs to backend
func (ps pubsub) Fire(topic string, pairs map[string]string) error {
	data, err := json.Marshal(pairs)
	if err != nil {
		return err
	}
	toTopic := ps.prefix + topic
	if err = ps.ns.Publish(toTopic, data); err != nil {
		return err
	}
	return nil
}

// Req requests data from backend with context
// context - timeout context
// topic - "auth" etc..
// key - key value, that identifies requester (uses for cache)
func (ps pubsub) Req(ctx context.Context, topic string, key string, pairs map[string]string) (ret map[string]string, err error) {
	pubTo := ps.prefix + "req." + topic
	subsTo := ps.prefix + "rep." + topic + "." + key
	var (
		data []byte
		sub  *nats.Subscription
		msg  *nats.Msg
	)
	if data, err = json.Marshal(pairs); err != nil {
		return
	}
	if sub, err = ps.ns.SubscribeSync(subsTo); err != nil {
		return
	}
	nm := nats.Msg{Subject: pubTo, Reply: subsTo, Data: data}
	if err = ps.ns.PublishMsg(&nm); err != nil {
		return
	}
	if msg, err = sub.NextMsgWithContext(ctx); err != nil {
		return
	}
	err = json.Unmarshal(msg.Data, &ret)
	return
}
