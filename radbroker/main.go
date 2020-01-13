package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/aam335/go-radius"
	"github.com/aam335/go-radius/vendor"
)

func main() {
	c, err := ConfigLoad()
	if err != nil {
		log.Fatal(err)
	}

	radius.Builtin.MustRegisterDC(vendor.Redback)
	// load filters
	userAuthFilter := c.MustNewFilter(radius.Builtin, "Auth")

	accFilters := [MaxKnownValue + 1]*radius.AttrFilter{
		AccountingOn:  c.MustNewFilter(radius.Builtin, "On"),
		AccountingOff: c.MustNewFilter(radius.Builtin, "Off"),
		Start:         c.MustNewFilter(radius.Builtin, "Start"),
		InterimUpdate: c.MustNewFilter(radius.Builtin, "InterimUpdate"),
		Stop:          c.MustNewFilter(radius.Builtin, "Stop"),
	}
	// load keys into key helper
	if err := userAuthFilter.SetKeys(c.Key); err != nil {
		log.Fatal(err)
	}
	ps, err := NewPubSub(c.Server.ServiceName, c.Server.NatsURI)
	if err != nil {
		log.Fatal(err)
	}

	handler := newHandler(userAuthFilter, accFilters, ps, c.Server.MaxAuthDuration.Duration)
	log.Infof("Starting server on nats:%v/%v", c.Server.NatsURI, c.Server.ServiceName)
	server := c.NewServer(radius.Builtin, radius.HandlerFunc(handler))
	log.Printf("exit status: %v", server.ListenAndServe())
}
