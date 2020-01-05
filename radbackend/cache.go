package main

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

const (
	// DefaultQueryLockTTL secs. Cache locks for DB queueing
	DefaultQueryLockTTL = 5
)

// Cache on redis
type Cache struct {
	prefix  string
	lockTTL int //seconds
	pool    *redis.Pool
}

// Send PING command to Redis
func ping(c redis.Conn) error {
	pong, err := c.Do("PING")
	if err != nil {
		return err
	}

	_, err = redis.String(pong, err)
	if err != nil {
		return err
	}
	return nil
}

// NewCache creates new cache object
// keys in this cache has prefix
func NewCache(redisAddr string, prefix string) (*Cache, error) {
	ca := Cache{prefix: prefix, lockTTL: DefaultQueryLockTTL}
	ca.pool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 1200,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddr)
			// if err != nil {
			// 	panic(err.Error())
			// }
			return c, err
		},
	}
	conn := ca.pool.Get()
	defer conn.Close()
	if err := ping(conn); err != nil {
		return nil, err
	}
	return &ca, nil
}

// GetPefix returns prefix for this cache
func (ca *Cache) GetPefix() string {
	return ca.prefix
}

// GetCache get value from cache, returns nil,nil on none exists.
// If cached value not exists and getFromSlowSource is not null, GetCache calls it to fill the cache
func (ca *Cache) GetCache(key string, getFromSlowSource func(key string) (int, []byte, error)) ([]byte, error) {
	// get prefix:key value
	// got value? return
	pfx := ca.prefix + ":" + key
	c := ca.pool.Get()
	defer c.Close()

	b, err := redis.Bytes(c.Do("GET", pfx))
	if err == redis.ErrNil {
		if getFromSlowSource == nil {
			return nil, nil
		}
		// not found, lock for blocking cyclic queries
		lockPfx := pfx + ":lock"
		lock, err := c.Do("set", lockPfx, time.Now().String(), "EX", ca.lockTTL, "NX")
		if err != nil {
			return nil, err
		}
		if lock == nil { // lock alredy set
			return nil, nil
		}
		ttl, data, err := getFromSlowSource(key)
		if err != nil {
			return nil, err
		}
		if data != nil {
			_, err = c.Do("set", pfx, data, "ex", ttl)
		}
		return data, err
	}
	return b, err
}
