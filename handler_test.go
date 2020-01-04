package main

import (
	"testing"
	"time"

	mock_pubsub "github.com/aam335/radmicro/mocks/pubsub"

	"github.com/aam335/go-radius"
	mock_go_radius "github.com/aam335/radmicro/mocks/radius"
	"github.com/golang/mock/gomock"
)

func defConf() (*Config, *radius.AttrFilter, [MaxKnownValue + 1]*radius.AttrFilter) {
	var c = Config{Filter: map[string][]string{
		"Auth":          []string{"User-Name"},
		"Start":         []string{"User-Name"},
		"Stop":          []string{"User-Name"},
		"InterimUpdate": []string{"User-Name"},

		"On":  []string{"User-Name"},
		"Off": []string{"User-Name"},
	},
	}
	userAuthFilter := c.MustNewFilter(radius.Builtin, "Auth")
	userAuthFilter.SetKeys([]radius.OneKey{{Name: "User-Name"}})
	accFilters := [MaxKnownValue + 1]*radius.AttrFilter{
		AccountingOn:  c.MustNewFilter(radius.Builtin, "On"),
		AccountingOff: c.MustNewFilter(radius.Builtin, "Off"),
		Start:         c.MustNewFilter(radius.Builtin, "Start"),
		InterimUpdate: c.MustNewFilter(radius.Builtin, "InterimUpdate"),
		Stop:          c.MustNewFilter(radius.Builtin, "Stop"),
	}

	return &c, userAuthFilter, accFilters
}

func Test_newHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	_, userAuthFilter, accFilters := defConf()
	defer ctrl.Finish()

	mrw := mock_go_radius.NewMockResponseWriter(ctrl)
	mp := mock_pubsub.NewMockPubSub(ctrl)

	h := newHandler(userAuthFilter, accFilters, mp, time.Second)

	p := radius.Packet{Code: radius.CodeAccessRequest, Dictionary: radius.Builtin}
	p.Add("User-Name", "username")

	// Access accept
	mp.
		EXPECT().
		Req(gomock.Any(), gomock.Eq("Auth"), gomock.Eq("username"), gomock.Eq(map[string]string{"User-Name": "username"})).
		Return(map[string]string{"Auth-Type": "Accept"}, nil)
	mrw.EXPECT().AccessAccept()
	h(mrw, &p)

	// Access reject
	mp.
		EXPECT().
		Req(gomock.Any(), gomock.Eq("Auth"), gomock.Eq("username"), gomock.Eq(map[string]string{"User-Name": "username"})).
		Return(map[string]string{"Auth-Type": "Reject"}, nil)
	mrw.EXPECT().AccessReject()
	h(mrw, &p)

	// Accounting
	for i, f := range accFilters {
		if f == nil {
			continue
		}
		p = radius.Packet{Code: radius.CodeAccountingRequest, Dictionary: radius.Builtin}
		p.Add("Acct-Status-Type", uint32(i))
		p.Add("User-Name", "username")

		mp.
			EXPECT().
			Fire(gomock.Eq(AcctStatusType(uint32(i))), gomock.Eq(map[string]string{"User-Name": "username"})).
			Return(nil)
		mrw.EXPECT().AccountingACK()
		h(mrw, &p)
	}
}
