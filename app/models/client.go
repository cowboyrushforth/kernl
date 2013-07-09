package models

import "kernl/app/lib/redishandle"
import "github.com/garyburd/redigo/redis"
import "errors"
import "time"

type Client struct {
  ConsumerKey string
  Secret string
  Title string
  Description string
  Host string
  Webfinger string
  Contacts string
  LogoUrl string
  RedirectUris string
  XType string
  CreatedAt int64
  UpdatedAt int64
}

func ClientFromConsumerKey(consumer_key string) (*Client, error) {
  client := Client{}
  v, errb := redis.Values(redishandle.Do("HGETALL", "client:"+consumer_key))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &client)
  if errc != nil {
    return nil, errc
  }
  if len(client.ConsumerKey) == 0 {
    return nil, errors.New("client not found")
  }

  return &client, nil
}

func (self *Client) Id() string {
  return "client:" + self.ConsumerKey
}

func (self *Client) Insert() bool {
  if self.ConsumerKey == "" {
    self.ConsumerKey = RandomString(16)
    self.Secret = RandomString(32)
    self.CreatedAt = time.Now().Unix()
  }
  self.UpdatedAt = time.Now().Unix()
  _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
  if errb != nil {
    panic(errb)
  }
  return true
}

