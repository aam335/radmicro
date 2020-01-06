package main

import "time"

// Config ...
type Config struct {
	Server struct {
		ServiceName     string   `env:"SERVICE_NAME"  env-description:"Service name for Nats & cache"`
		MaxConnections  int      `env:"MAX_CONNECTIONS" env-description:"Maximum of running workers"`
		RestartInterval duration `env:"RESTART_INTERVAL" env-description:"Minimum worker restart interval"`
	}
	SQL struct {
		Driver string `env:"SQL_DRIVER" env-description:"Sql driver name"`
		URI    string `env:"SQL_URI" env-description:"Sql URI"`
		Query  map[string]Query
	}
	Nats struct {
		URI       string `env:"NATS_URI" env-description:"Nats URI"`
		QueueName string `env:"QUEUE_NAME" env-description:"Nats Queue name"`
	}
	Redis struct {
		URI       string   `env:"REDIS_URI" env-description:"Redis URI"`
		LockTTL   duration `env:"LOCK_TTL" env-description:"Time, that cache skips querying SQL backend on the same key min 1s"`
		MaxIdle   int      `env:"REDIS_MAXIDLE" env-description:"Redis Pool.Maxidle"`
		MaxActive int      `env:"REDIS_MAXACTIVE" env-description:"Redis Pool.MaxActive"`
	}
}

// Query defines query details
type Query struct {
	Prepare   string
	Arguments []string
	Cacheable bool
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
