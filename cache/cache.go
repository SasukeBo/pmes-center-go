package cache

import (
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/allegro/bigcache"
	"time"
)

var (
	globalcache *bigcache.BigCache
)

func init() {
	var err error
	globalcache, err = bigcache.NewBigCache(bigcache.DefaultConfig(30 * time.Minute))
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize cache: %v\n", err))
	}
}

// Set cache
func Set(key , value string) {
	globalcache.Set(key, []byte(value))
}

// Get cache
func Get(key string) (string, error) {
	v, err := globalcache.Get(key)
	if err != nil {
		return "", err
	}

	return string(v), nil
}

// Delete a key value from global cache
func Delete(key string) {
	err := globalcache.Delete(key)
	if err != nil {
		log.Warnln(fmt.Sprintf("delete cache key=%s failed: %v\n", key, err))
	}
}
