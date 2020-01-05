package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis"
)

func TestNew(t *testing.T) {

	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	type args struct {
		redisAddr string
		prefix    string
	}
	tests := []struct {
		name    string
		args    args
		want    *Cache
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "fail addr", args: args{redisAddr: "unreach!", prefix: "prefix"}, want: nil, wantErr: true},
		{name: "proper addr", args: args{redisAddr: s.Addr(), prefix: "prefix"}, want: &Cache{prefix: "prefix", lockTTL: DefaultQueryLockTTL}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCache(tt.args.redisAddr, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && got == nil {
				return
			}
			if !(got.prefix == tt.want.prefix && got.lockTTL == tt.want.lockTTL) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_GetCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	ca, err := NewCache(s.Addr(), "test")
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		prefix  string
		lockTTL int
	}
	type args struct {
		key               string
		getFromSlowSource func(string) (int, []byte, error)
	}

	getFunc := func(key string) (int, []byte, error) {
		if string(key) == "error" {
			return 10, nil, errors.New("error")
		}
		if string(key) == "nils" {
			return 0, nil, nil
		}

		return 3, []byte("test"), nil
	}

	f := fields{prefix: "pfx", lockTTL: 5}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "not in cache", fields: f, args: args{key: "aaaaa", getFromSlowSource: getFunc}, want: []byte("test"), wantErr: false},
		{name: "in cache", fields: f, args: args{key: "aaaaa", getFromSlowSource: getFunc}, want: []byte("test"), wantErr: false},
		{name: "in cache no callback", fields: f, args: args{key: "aaaaa", getFromSlowSource: nil}, want: []byte("test"), wantErr: false},
		{name: "not in cache no callback", fields: f, args: args{key: "aaaaabbb", getFromSlowSource: nil}, want: nil, wantErr: false},
		{name: "callback error", fields: f, args: args{key: "error", getFromSlowSource: getFunc}, want: nil, wantErr: true},
		{name: "nils", fields: f, args: args{key: "nils", getFromSlowSource: getFunc}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &Cache{
				prefix:  tt.fields.prefix,
				lockTTL: tt.fields.lockTTL,
				pool:    ca.pool,
			}
			got, err := ca.GetCache(tt.args.key, tt.args.getFromSlowSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cache.GetCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cache.GetCache() = %v, want %v", got, tt.want)
			}
		})
	}
}
