package main

import (
	"github.com/aam335/go-radius"
	"github.com/aam335/go-radius/vendor"
	"log"
)

// Acct-Status-Type attribute
const (
	Start         = 1
	Stop          = 2
	InterimUpdate = 3
	AccountingOn  = 7
	AccountingOff = 8
	MaxKnownValue = 8
)

func main() {
	c, err := ConfigLoad()
	if err != nil {
		log.Fatal(err)
	}

	radius.Builtin.MustRegisterDC(vendor.Redback)
	// load tilters
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

	handler := func(w radius.ResponseWriter, p *radius.Packet) {
		switch p.Code {
		case radius.CodeAccessRequest: // auth
			accept := true
			if attrs, err := userAuthFilter.Filter(p); err == nil {
				_ = attrs
			} else {
				// debug
			}
			if accept {
				w.AccessAccept()
			} else {
				w.AccessReject()
			}
		case radius.CodeAccountingRequest: //accounting
			acctStatusType := p.Attr("Acct-Status-Type")
			if acctStatusType != nil {
				if sType, ok := acctStatusType.Value.(uint32); ok {
					if sType <= MaxKnownValue && accFilters[sType] != nil {
						if attrs, err := userAuthFilter.Filter(p); err == nil {
							_ = attrs
						} else {
							//debug
						}
					}
				}
			}
			w.AccountingACK()
		}
		log.Println("packet", w.RemoteAddr(), w.LocalAddr())
	}

	server := c.NewServer(radius.Builtin, radius.HandlerFunc(handler))
	_ = server
}
