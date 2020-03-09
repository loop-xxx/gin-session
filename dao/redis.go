package dao

import (
	"github.com/go-redis/redis"
	"time"
)

type Redis struct {
	*redis.Client
}

const magicField = string("magic")

func DefaultRedis(addr,password string, db int) (r Redis, err error) {
	client := redis.NewClient(redis.Options{
		Addr:               addr,
		Password:           password,
		DB:                 db,
	})
	if _, err = client.Ping().Result(); err == nil {
		r = Redis{
			Client: client,
		}
	}

	return
}

func (r Redis) Done() (err error) {
	err = r.Close()
	return
}

func (r Redis) Check(key string, magic string, expiration time.Duration) (status Status, err error) {
	var exist bool
	var redisMagic string

	if exist, err = r.Expire(key, expiration).Result(); err == nil {
		if exist {
			if redisMagic, err = r.HGet(key, magicField).Result(); err == nil {
				if redisMagic == magic {
					status = GinSessionOk
				} else {
					status = GinSessionOld
				}
			}
		} else {
			status = GinSessionTimeout
		}
	}
	return
}

func (r Redis) Peek(key, magic string) (different bool, err error) {
	var redisMagic string
	if redisMagic, err = r.HGet(key, magicField).Result(); err == nil {
		if redisMagic != magic {
			different = true
		}
	}
	return
}

// session数据在redis中的形式
// HMSet redisKey magicField magicValue field1 fieldValue1 field2 fieldValue2 ...

func (r Redis) Pull(key string) (magic string, dataMap map[string]string, err error) {
	dataMap, err = r.HGetAll(key).Result()
	if err == nil {
		magic = dataMap[magicField]
	}
	return
}

func (r Redis) Push(key string, magic string, dataMap map[string]string, expiration time.Duration) (err error) {

	args := make([]interface{}, 0,len(dataMap)*2 + 2 + 2)
	args = append(args, "HMSET", key)
	args = append(args, magicField, magic)

	for key, value := range dataMap {
		args = append(args, key, value)
	}
	if _, err = r.Do(args...).Result(); err == nil {
		_, err = r.Expire(key, expiration).Result()
	}
	return
}
