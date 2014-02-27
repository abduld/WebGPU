package stats

import (
	"strconv"
	"time"
	. "wb/app/config"

	"github.com/garyburd/redigo/redis"
	"github.com/robfig/revel"
)

type RPool struct {
	p *redis.Pool
}

type RedisConnection struct {
	c redis.Conn
}

var (
	RedisPool *RPool
)

func (pool *RPool) ActiveCount() int {
	return pool.p.ActiveCount()
}

func (pool *RPool) Close() {
	pool.p.Close()
}

func (pool *RPool) Get() *RedisConnection {
	c := pool.p.Get()
	if c == nil {
		revel.WARN.Println("Cannot get redis connection from pool")
	}
	conn := &RedisConnection{c}
	return conn
}

func (pool *RPool) Set(key, val string) {
	conn := pool.Get()
	conn.Set(key, val)
	//conn.Close()
}

func (pool *RPool) Append(key, val string) {
	conn := pool.Get()
	conn.Append(key, val)
	//conn.Close()
}

func (pool *RPool) Incr(key string) {
	conn := pool.Get()
	conn.Incr(key)
	//conn.Close()
}

func (pool *RPool) Decr(key string) {
	conn := pool.Get()
	conn.Decr(key)
	//conn.Close()
}

func (conn *RedisConnection) Set(key, val string) {
	if conn != nil {
		t := strconv.Itoa(time.Now().Nanosecond())
		conn.c.Send("SET", key+"::"+t, val)
		conn.c.Send("PUBLISH", key, val)
		conn.c.Flush()
	}
}

func (conn *RedisConnection) Incr(key string) {
	if conn != nil {
		conn.c.Send("INCR", key)
		conn.c.Send("PUBLISH", key)
		conn.c.Flush()
	}
}

func (conn *RedisConnection) Decr(key string) {
	if conn != nil {
		conn.c.Send("DECR", key)
		conn.c.Send("PUBLISH", key)
		conn.c.Flush()
	}
}

func (conn *RedisConnection) Append(key, val string) {
	if conn != nil {
		conn.Set(key, val)
	}
}

func RedisSubscribe(ch string,
	onMessage func(redis.Message, string, redis.PubSubConn) error,
	onPMessage func(redis.PMessage, string, redis.PubSubConn) error,
	onSubscription func(redis.Subscription, string, redis.PubSubConn) error,
	onError func(error, string, redis.PubSubConn) error) {
	conn := RedisPool.Get()
	if conn != nil {
		psc := redis.PubSubConn{conn.c}
		psc.PSubscribe(ch)
		var err error
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				err = onMessage(v, ch, psc)
			case redis.PMessage:
				err = onPMessage(v, ch, psc)
			case redis.Subscription:
				err = onSubscription(v, ch, psc)
			case error:
				err = onError(v, ch, psc)
			}
			if err != nil {
				psc.PUnsubscribe()
				psc.Close()
				return
			}
		}
	}
}

func (conn *RedisConnection) Close() {
	if conn != nil {
		conn.Close()
	}
}

func CreateRedisConnection() (redis.Conn, error) {
	c, err := redis.Dial(RedisProtocol, RedisAddress)
	if err != nil {
		return nil, err
	}
	if RedisPassword != "" {
		if _, err := c.Do("AUTH", RedisPassword); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func InitRedisStats() {
	RedisPool = &RPool{
		p: redis.NewPool(CreateRedisConnection, 40),
	}
}
