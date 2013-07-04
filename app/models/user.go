package models

import "fmt"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "errors"

type User struct {
  DisplayName  string
  Email        string
  Slug         string
  PwdHash     []byte
  RSAKey      string
  Guid        string
  AccountIdentifier string
}

func FetchUid(c redis.Conn, uid string) (*User, error) {
  user := User{}
  v, errb := redis.Values(c.Do("HGETALL", uid))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &user)
  if errc != nil {
    return nil, errc
  }
  if len(user.Email) == 0 {
    return nil, errors.New("user not found")
  }

  return &user, nil
}

func (self *User) String() string {
  if len(self.DisplayName) > 0 {
    return self.DisplayName
  } else {
    return self.Email
  }
}

func (self *User) Validate(c redis.Conn, v *revel.Validation) {

  // check email sanity
  v.Check(self.Email, 
    revel.Required{},
    revel.MaxSize{100},
    revel.MinSize{5},
  ).Key("user.Email")

  // if above validations check run the email uniqueness check
  if(v.HasErrors() == false) {
    v.Required(self.EmailDoesNotExist(c)).Key("user.Email").Message("Email Already Exists")
  }
}

func (self *User) Id() string {
  return fmt.Sprintf("user:%s", self.Email)
}

func (self *User) EmailDoesNotExist(c redis.Conn) bool {
  exists, err := redis.Bool(c.Do("EXISTS", self.Id()))
  if err != nil {
    panic("data access problem")
  }
  if exists {
    return false
  }
  return true
}

func (self User) Insert(c redis.Conn) bool {
  // TODO
  // add email regex

  _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(&self)...)
  if errb != nil {
    panic("data access problem")
  }

  return true
}
