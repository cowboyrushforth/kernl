package models

import "kernl/app/lib/redishandle"
import "github.com/garyburd/redigo/redis"
import "errors"
import "time"

type RequestToken struct {
    Token string
    ConsumerKey string
    Callback string
    Used bool
    TokenSecret string
    Verifier string
    Authenticated bool
    Username string
    AccessToken string
    CreatedAt int64
    UpdatedAt int64
}

func RequestTokenFromToken(token string) (*RequestToken, error) {
  request_token := RequestToken{}
  v, errb := redis.Values(redishandle.Do("HGETALL", "reqtoken:"+token))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &request_token)
  if errc != nil {
    return nil, errc
  }
  if len(request_token.ConsumerKey) == 0 {
    return nil, errors.New("token not found")
  }

  return &request_token, nil
}

func (self *RequestToken) Id() string {
  return "reqtoken:" + self.Token
}

func (self *RequestToken) Insert() bool {
  if self.Token == "" {
    self.Token = RandomString(16)
    self.TokenSecret = RandomString(32)
    self.Verifier = RandomString(16)
    self.CreatedAt = time.Now().Unix()
    self.Used = false
    self.Authenticated = false
  }
  self.UpdatedAt = time.Now().Unix()
  _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
  if errb != nil {
    panic(errb)
  }
  return true
}

