package config

import (
	"strconv"
	"time"
)

var (
	RedisPassword          string
	RedisProtocol          string
	RedisAddress           string
	EnableRedisStore       bool
	RedisReadTimeout       time.Duration
	RedisWriteTimeout      time.Duration
	RedisConnectionTimeout time.Duration
)

func InitRedisConfig() {
	conf := NestedRevelConfig

	readDuration := func(s string) time.Duration {
		t := conf.IntDefault(s, 0)
		return time.Duration(t) * time.Second
	}

	EnableRedisStore, _ = conf.Bool("redis.enable")
	if EnableRedisStore {
		RedisProtocol, _ = conf.String("redis.protocol")
		RedisPassword = conf.StringDefault("redis.password", "")
		RedisConnectionTimeout = readDuration("redis.connect_timeout")
		RedisReadTimeout = readDuration("redis.read_timeout")
		RedisWriteTimeout = readDuration("redis.write_timeout")
		if IsMaster {
			RedisAddress = "" // localhost
		} else {
			RedisAddress = MasterIP
		}
		port, _ := conf.Int("redis.port")
		RedisAddress += ":" + strconv.Itoa(port)
	}
}
