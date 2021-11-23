# 用法

## go-cache
```go
package main

import (
	"fmt"
	gocache "github.com/patrickmn/go-cache"
	"github.com/zhiting-tech/smartassistant/pkg/cache"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"time"
)

func main() {
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := store.NewGoCache(gocacheClient, nil)

	cacheManager := cache.New(gocacheStore)
	err := cacheManager.Set("my-key", "my-value", 0)
	if err != nil {
		panic(err)
	}

	value, err := cacheManager.Get("my-key")
	if err == store.ErrValueNotFound {
		fmt.Println("value not found")
	} else {
		fmt.Println(value)
	}

}

```

## redis
```go
package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhiting-tech/smartassistant/pkg/cache"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"time"
)

func main() {
	redisStore := store.NewRedis(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}), nil)

	cacheManager := cache.New(redisStore)
	err := cacheManager.Set("my-key", "my-value", 15*time.Second)
	if err != nil {
		panic(err)
	}

	value, err := cacheManager.Get("my-key")
	if err == store.ErrValueNotFound {
		fmt.Println("value not found")
	} else {
		fmt.Println(value)
	}

}

```

## 通过包名方式
```go
package main

import (
	"fmt"
	"github.com/zhiting-tech/smartassistant/pkg/cache"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
)

func main() {
	//如果想使用自定义存储，可以使用一下方式初始化，默认使用的是go-cache作为存储。
	//redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	//redisStore := store.NewRedis(redisClient, nil)
	//cache.InitCache(redisStore)
	
	err := cache.Set("my-key", "my-value", 0)
	if err != nil {
		panic(err)
	}

	value, err := cache.Get("my-key")
	if err == store.ErrValueNotFound {
		fmt.Println("value not found")
	} else {
		fmt.Println("value:", value)
	}

	value, ttl, err := cache.GetWithTTL("my-key")
	if err == store.ErrValueNotFound {
		fmt.Println("value not found")
	} else {
		fmt.Println("value:", value, " ttl:", ttl)
	}

	err = cache.Delete("my-key")
	if err != nil {
		panic(err)
	}

	value, err = cache.Get("my-key")
	if err == store.ErrValueNotFound {
		fmt.Println("value not found")
	} else {
		fmt.Println("value:", value)
	}

}

```