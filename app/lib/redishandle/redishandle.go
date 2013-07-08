package redishandle

import "github.com/garyburd/redigo/redis"
import "time"

type RedisHandle struct {
  *redis.Pool
}


func NewRedisHandle(host string, port string) RedisHandle {
  redis :=  redis.NewPool(func() (redis.Conn, error) {
    c, err := redis.Dial("tcp", host+":"+port)
    if err != nil {
      panic(err)
    }
    return c, err
  }, 3)
  return RedisHandle{redis}
}

func (self RedisHandle) GetRedisConn() redis.Conn {
  rc := self.Pool.Get()
  for i := 0; i < 5; i++ {
    if err := rc.Err(); err != nil {
      time.Sleep(10 * time.Millisecond)
      rc = self.Pool.Get()
    } else {
      break
    }
  }
  return rc
}

func (self RedisHandle) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
  rc := self.GetRedisConn()
  defer rc.Close() 
  return rc.Do(commandName, args...) 
}

type RedisAccess interface {
  Do(commandName string, args ...interface{}) (reply interface{}, err error) 
}

var Instance RedisAccess

// The package implements the Cache interface (as sugar).
func Do(commandName string, args ...interface{}) (reply interface{}, err error) {
  return Instance.Do(commandName, args...) 
}

