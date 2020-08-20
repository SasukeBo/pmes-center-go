package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/log"
	"github.com/go-redis/redis/v8"
	"reflect"
	"time"
)

var (
	redisClient *redis.Client
	expiredTime = configer.GetInt("cache_expired_time")
)

func init() {
	connectRedis()
}

func connectRedis() {
	redisHost := configer.GetString("redis_host")
	redisPort := configer.GetString("redis_port")
	redisClient = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", redisHost, redisPort),
		DialTimeout: time.Minute,
	})
	if redisClient != nil {
		log.Info("redis server connected")
	}
}

func Pipelined(fn func(pip redis.Pipeliner) error) ([]redis.Cmder, error) {
	return redisClient.Pipelined(redisClient.Context(), fn)
}

func SetWithPip(pip redis.Pipeliner, key string, value interface{}, opts ...interface{}) error {
	value = prepareSet(value)
	expiredDuration := prepareExpire(opts)
	return pip.Set(Ctx(), key, value, expiredDuration).Err()
}

func prepareSet(value interface{}) interface{} {
	dt := reflect.TypeOf(value)
	switch dt.Kind() {
	case reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		if marshaled, err := json.Marshal(value); err != nil {
			value = fmt.Sprint(value)
		} else {
			value = string(marshaled)
		}
	default:
		value = fmt.Sprint(value)
	}
	return value
}

func prepareExpire(opts ...interface{}) time.Duration {
	var expireDuration = time.Duration(expiredTime) * time.Second
	if len(opts) > 0 {
		if v, ok := opts[0].(time.Duration); ok {
			expireDuration = v
		}
	}

	return expireDuration
}

func Set(key string, value interface{}, opts ...interface{}) error {
	value = prepareSet(value)
	var expireDuration = prepareExpire(opts)
	return redisClient.Set(redisClient.Context(), key, value, expireDuration).Err()
}

func Get(key string) (string, error) {
	get := getRedis(key)
	if err := get.Err(); err != nil {
		return "", err
	}
	return get.Val(), nil
}

func GetBool(key string) (bool, error) {
	get := getRedis(key)
	if err := get.Err(); err != nil {
		return false, err
	}

	v, err := get.Int()
	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func GetFloat(key string) (float64, error) {
	get := getRedis(key)
	if err := get.Err(); err != nil {
		return 0, err
	}

	v, err := get.Float64()
	if err != nil {
		return 0, err
	}

	return v, nil
}

func GetInt(key string) (int, error) {
	get := getRedis(key)
	if err := get.Err(); err != nil {
		return 0, err
	}

	v, err := get.Int()
	if err != nil {
		return 0, err
	}

	return v, nil
}

func Scan(key string, out interface{}) error {
	get := getRedis(key)
	if err := get.Err(); err != nil {
		return err
	}

	if get.Val() == "" {
		return errors.New("key not found")
	}
	if err := json.Unmarshal([]byte(get.Val()), out); err != nil {
		return err
	}

	return nil
}

func getRedis(key string) *redis.StringCmd {
	return redisClient.Get(redisClient.Context(), key)
}

func Del(key string) error {
	del := redisClient.Del(redisClient.Context(), key)
	return del.Err()
}

func Ctx() context.Context {
	return redisClient.Context()
}
