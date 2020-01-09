package main

import (
	"context"
	"time"

	"github.com/aam335/go-radius"
	"github.com/micro/go-micro/util/log"
)

func newHandler(userAuthFilter *radius.AttrFilter, accFilters [MaxKnownValue + 1]*radius.AttrFilter, ps PubSub, maxAuthTime time.Duration) radius.HandlerFunc {
	handler := func(w radius.ResponseWriter, p *radius.Packet) {
		switch p.Code {
		case radius.CodeAccessRequest: // auth
			accept := false
			key, attrs := userAuthFilter.FilterStrings(p)
			ctx, cancel := context.WithTimeout(context.Background(), maxAuthTime)
			resp, err := ps.Req(ctx, "Auth", key, attrs)
			cancel()
			if err != nil {
				// deadline exceeded, DHCP client will stops waiting THIS reply and
				// resends DHCP query. No reply to timeouted query.
				if ctx.Err() != nil {
					log.Info("Timeout:", key)
				} else {
					log.Error(err)
				}
				return
			}
			if resp["Auth-Type"] == "Accept" {
				accept = true
			}
			delete(resp, "Auth-Type")
			replyAttrs, err := p.Dictionary.StrsToAttrs(resp)
			if accept {
				w.AccessAccept(replyAttrs...)
			} else {
				w.AccessReject(replyAttrs...)
			}
		case radius.CodeAccountingRequest: //accounting
			acctStatusType := p.Attr("Acct-Status-Type")
			if acctStatusType != nil {
				if sType, ok := acctStatusType.Value.(uint32); ok {
					if sType <= MaxKnownValue && accFilters[sType] != nil {
						_, attrs := accFilters[sType].FilterStrings(p)
						topic := AcctStatusType(sType)
						if err := ps.Fire(topic, attrs); err != nil {
							log.Error(topic, ":", err)
						}
					}
				}
			}
			w.AccountingACK()
		}
		// log.Println("packet", w.RemoteAddr(), w.LocalAddr())
	}
	return handler
}
