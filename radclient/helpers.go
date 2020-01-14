package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	radius "github.com/aam335/go-radius"
)

// radCode represents radCode of radius packet
type radCode struct {
	radius.Code
}

// radius.Code

// RadField ...
type RadField struct {
	Code  radCode
	Attrs map[string]string
}

// RadFields describes packet text representation
type RadFields struct {
	Send, Recv RadField
	WantErr    bool
}

var codes = map[string]radius.Code{
	"AccessRequest":      radius.CodeAccessRequest,
	"AccessAccept":       radius.CodeAccessAccept,
	"AccessReject":       radius.CodeAccessReject,
	"AccountingResponse": radius.CodeAccountingResponse,

	"AccessChallenge": radius.CodeAccessChallenge,

	"StatusServer": radius.CodeStatusServer,
	"StatusClient": radius.CodeStatusClient,

	"DisconnectRequest": radius.CodeDisconnectRequest,
	"DisconnectACK":     radius.CodeDisconnectACK,
	"DisconnectNAK":     radius.CodeDisconnectNAK,

	"CoARequest": radius.CodeCoARequest,
	"CoAACK":     radius.CodeCoAACK,
	"CoANAK":     radius.CodeCoANAK,
	"Any":        0,
}

// UnmarshalText ...
func (c *radCode) UnmarshalText(text []byte) error {
	return c.SetValue(string(text))
}

// SetValue ..
func (c *radCode) SetValue(s string) error {
	if _, ok := codes[s]; !ok {
		return fmt.Errorf("Wrong Code value '%v', assepted:%v", s, reflect.ValueOf(codes).MapKeys())
	}
	c.Code = codes[s]
	return nil
}

// MarshalText ...
func (c *radCode) MarshalText() ([]byte, error) {
	for k, v := range codes {
		if v == c.Code {
			return []byte(k), nil
		}
	}
	return nil, fmt.Errorf("Wrong Code value '%v'", c)
}

func (c radCode) String() string {
	bs, err := c.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

func (c radCode) UnmarshalJSON(text []byte) error {
	return c.SetValue(string(text))
}

func (c radCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}
