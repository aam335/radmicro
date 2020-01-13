package main

import (
	"fmt"
	"reflect"

	radius "github.com/aam335/go-radius"
)

// code represents code of radius packet
type code struct {
	radius.Code
	// Code byte
}

var codes = map[string]radius.Code{
	"AccessRequest": radius.CodeAccessRequest,
	"AccessAccept":  radius.CodeAccessAccept,
	"AccessReject":  radius.CodeAccessReject,

	"AccountingRequest":  radius.CodeAccountingRequest,
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
}

// UnmarshalText ...
func (c code) UnmarshalText(s string) error {
	if _, ok := codes[s]; !ok {
		return fmt.Errorf("Wrong Code value '%v', assepted:%v", s, reflect.ValueOf(codes).MapKeys())
	}
	c.Code = codes[s]
	return nil
}
