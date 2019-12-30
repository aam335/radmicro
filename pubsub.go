package main

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

// PubSub sends topics with attributes to backend
// and receives replies, if needed
type PubSub struct {
	ns     *nats.Conn
	prefix string
}

// NewPubSub creates new PubSub
func NewPubSub(prefix string, natsURI string) (*PubSub, error) {
	ns, err := nats.Connect(natsURI)
	if err != nil {
		return nil, err
	}
	ps := PubSub{ns: ns, prefix: prefix + "."}
	return &ps, nil
}

// Fire Pubs json encoded pairs to backend
func (ps *PubSub) Fire(topic string, pairs map[string]string) error {
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
func (ps *PubSub) Req(ctx context.Context, topic string, key string, pairs map[string]string) (ret map[string]string, err error) {
	subsTo := ps.prefix + "subs." + topic + key
	pubTo := ps.prefix + key
	data, err := json.Marshal(pairs)
	if err != nil {
		return nil, err
	}
	sub, _ := ps.ns.SubscribeSync(subsTo)
	nm := nats.Msg{Subject: pubTo, Reply: subsTo, Data: data}
	ps.ns.PublishMsg(&nm)
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(msg.Data, &ret)
	return
}
