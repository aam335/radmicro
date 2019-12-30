package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aam335/go-radius"
	"github.com/jinzhu/configor"
	"golang.org/x/time/rate"
)

// Config ...
type Config struct {
	Server struct {
		BindTo           string   `env:"BIND_TO"  default:"0.0.0.0:1812"`
		BufferSize       int      `env:"BUFFER_SIZE"`
		ReplicateTo      []string `env:"REPLICATE_TO" `
		ReplicateReplies bool     `env:"REPLICATE_REPLIES" `
		Secret           string   `env:"SECRET" `
		ClientsSecrets   []string `env:"HOST_SECRETS" `
		ServiceName      string   `env:"SERVICE_NAME" `
	} ``
	RateLimit struct {
		MaxPendingReq    int `env:"MAX_PENDING_REQ" `
		RequestPerSecond int `env:"REQUEST_PER_SECOND" `
		Burst            int `env:"BURST" `
	} ``
	Key    []radius.OneKey     `env:"KEY"`
	Filter map[string][]string `env:"FILTER" `
}

// GetClientSecrets ...
func (c *Config) GetClientSecrets() map[string]string {
	clientSecrets := make(map[string]string)
	for _, hostSecret := range c.Server.ClientsSecrets {
		fields := strings.Split(hostSecret, ":")
		if len(fields) != 2 {
			log.Panicf("%v not is in IP:secret format", hostSecret)
		}
		clientSecrets[fields[0]] = fields[1]
	}
	return clientSecrets
}

// ConfigLoad Load config from file and os.Env
// for syntax and logic details look https://github.com/jinzhu/configor
func ConfigLoad() (*Config, error) {
	var mainConfig Config
	gen := flag.Bool("gen", false, "generate toml config")
	config := flag.String("c", "config.toml", "config file name")
	flag.Parse()
	if err := configor.Load(&mainConfig, *config); err != nil {
		return nil, err
	}
	if *gen {
		var buff bytes.Buffer
		e := toml.NewEncoder(&buff)
		if err := e.Encode(mainConfig); err == nil {
			fmt.Print(string(buff.Bytes()))
			os.Exit(0)
		} else {
			return nil, err
		}
	}
	return &mainConfig, nil
}

// NewServer ..
func (c *Config) NewServer(d *radius.Dictionary, h radius.HandlerFunc) *radius.Server {
	r := radius.Server{
		Addr:               c.Server.BindTo,
		Secret:             []byte(c.Server.Secret),
		ClientsSecrets:     c.GetClientSecrets(),
		ReplicateTo:        c.Server.ReplicateTo,
		ReplicateReplies:   c.Server.ReplicateReplies,
		MaxPendingRequests: uint32(c.RateLimit.MaxPendingReq),
		BufferSize:         c.Server.BufferSize,
		Dictionary:         d,
		Handler:            h,
	}
	if c.RateLimit.RequestPerSecond > 0 {
		rl := time.Duration(uint64(time.Second) / uint64(c.RateLimit.RequestPerSecond))
		lim := rate.Every(rl)
		if c.RateLimit.Burst == 0 {
			c.RateLimit.Burst = 100 // Maximum one-time burst for queries
		}
		r.RateLimiter = rate.NewLimiter(lim, c.RateLimit.Burst)
		r.RateLimiterCtx = context.Background()
	}
	return &r
}

// NewFilter makes new filter based on dictionary & loaded configuration
func (c *Config) NewFilter(d *radius.Dictionary, filterID string) (*radius.AttrFilter, error) {
	if _, ok := c.Filter[filterID]; ok {
		return d.NewAttrFilter(c.Filter[filterID])
	}
	return nil, fmt.Errorf("Attribute filter for %v not known", filterID)
}

// MustNewFilter forced NewFilter
func (c *Config) MustNewFilter(d *radius.Dictionary, filterID string) *radius.AttrFilter {
	filter, err := c.NewFilter(d, filterID)
	if err != nil {
		log.Fatal(err)
	}
	return filter
}
