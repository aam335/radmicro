package main

import (
	"context"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/ilyakaznacheev/cleanenv"
)

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
	rc := &Cache{prefix: c.Server.ServiceName,
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
	c.runWorkers(context.Background(), rc, worker)
}
