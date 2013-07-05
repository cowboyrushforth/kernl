package models

import "github.com/garyburd/redigo/redis"
import "fmt"

type EmailDoesNotExist struct {
   Redis redis.Conn
}

func (self EmailDoesNotExist) IsSatisfied(i interface{}) bool {
  email := i.(string) 
  exists, err := redis.Bool(self.Redis.Do("EXISTS", "email:"+email))
  if err != nil {
    panic("data access problem")
  }
  if exists {
    return false
  }
  return true
}

func (v EmailDoesNotExist) DefaultMessage() string {
  return "Email already exists"
}

type SlugDoesNotExist struct {
   Redis redis.Conn
}

func (self SlugDoesNotExist) IsSatisfied(i interface{}) bool {
  slugkey := fmt.Sprintf("user:%s", i.(string))
  exists, err := redis.Bool(self.Redis.Do("EXISTS", slugkey))
  if err != nil {
    panic("data access problem")
  }
  if exists {
    return false
  }
  return true
}

func (v SlugDoesNotExist) DefaultMessage() string {
  return "Slug already exists"
}

