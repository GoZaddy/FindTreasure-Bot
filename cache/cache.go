package cache

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gozaddy/findtreasure/customerrors"
	"time"
)

var (
	pool *redis.Pool
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func () (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}



func init(){
	pool = newPool(":6379")
}

func Get(key string) (interface{}, error){
	c := pool.Get()
	res, err := c.Do("GET", key)
	if res == nil {
		return nil, customerrors.ErrNilRedisValue
	}
	c.Close()
	return res, err
}

func Set(key string, value interface{}, time string) error{
	c := pool.Get()
	_, err := c.Do("SETEX", key, time, value)
	c.Close()
	return err
}

func Delete(key string) error{
	c := pool.Get()
	_, err := c.Do("DEL", key)
	c.Close()
	return err
}
