package main

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ilyakaznacheev/cleanenv"
)

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

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// Rc Redis Cache
var Rc *Cache

func main() {
	var c Config
	var err error
	if err = cleanenv.ReadConfig("config.toml", &c); err != nil {
		log.Fatal(err)
	}
	// init Redis cache pool
	lockTTL := c.Redis.LockTTL.Seconds()
	if lockTTL < 1 {
		lockTTL = 1
	}
	Rc = &Cache{prefix: c.Server.ServiceName,
		lockTTL: int(lockTTL),
		Pool: &redis.Pool{
			MaxIdle:   c.Redis.MaxIdle,
			MaxActive: c.Redis.MaxActive,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", c.Redis.URI)
				if err != nil {
					log.Print("Redis pool:", err)
				}
				return c, err
			},
		},
	}

}
