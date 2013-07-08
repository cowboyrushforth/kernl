package redishandle

import "github.com/robfig/revel"

func init() {
  revel.OnAppStart(func() {
    redis_host := revel.Config.StringDefault("redis.host", "127.0.0.1")
    redis_port := revel.Config.StringDefault("redis.port", "6379")
    Instance = NewRedisHandle(redis_host, redis_port)
  })
}
