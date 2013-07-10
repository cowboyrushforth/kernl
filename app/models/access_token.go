package models

import "kernl/app/lib/redishandle"
import "github.com/garyburd/redigo/redis"
import "errors"
import "time"

type AccessToken struct {
    Token string
    TokenSecret string
    ConsumerKey string
    RequestToken string
    Slug string
    CreatedAt int64
    UpdatedAt int64
}

func AccessTokenFromToken(token string) (*AccessToken, error) {
  access_token := AccessToken{}
  v, errb := redis.Values(redishandle.Do("HGETALL", "acctoken:"+token))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &access_token)
  if errc != nil {
    return nil, errc
  }
  if len(access_token.ConsumerKey) == 0 {
    return nil, errors.New("token not found")
  }
  return &access_token, nil
}

func (self *AccessToken) Id() string {
  return "acctoken:" + self.Token
}

func (self *AccessToken) Insert() bool {
  if self.Slug == "" ||
     self.RequestToken == "" ||
     self.ConsumerKey == "" {
       panic("missing fields")
  }
  if self.Token == "" {
    self.Token = RandomString(16)
    self.TokenSecret = RandomString(32)
    self.CreatedAt = time.Now().Unix()
  }
  self.UpdatedAt = time.Now().Unix()
  _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
  if errb != nil {
    panic(errb)
  }
  return true
}

