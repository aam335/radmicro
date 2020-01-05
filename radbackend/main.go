package main

import (
	"log"
	"time"

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
		URI    string `env:"SQL_CONN_URI" env-description:"Sql URI"`
	}
	Nats struct {
		URI       string `env:"NATS_URI" env-description:"Nats URI"`
		QueueName string `env:"QUEUE_NAME" env-description:"Nats Queue name"`
	}
	Redis struct {
		URI string `env:"REDIS_URI" env-description:"Redis URI"`
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

func main() {
	var c Config
	if err := cleanenv.ReadConfig("config.toml", &c); err != nil {
		log.Fatal(err)
	}

}
