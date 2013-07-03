package models

import "fmt"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "errors"

type User struct {
        DisplayName  string
        Email        string
        PwdHash     []byte
}

func FetchUid(uid string) (*User, error) {
  c, err := redis.Dial("tcp", ":6379")
  if err != nil {
    return nil, err
  }
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

func (u *User) String() string {
        return fmt.Sprintf("User(%s)", u.Email)
}

func (self *User) Validate(v *revel.Validation) {

        // check email sanity
        v.Check(self.Email, 
          revel.Required{},
          revel.MaxSize{100},
          revel.MinSize{5},
        ).Key("user.Email")

        // if above validations check run the email uniqueness check
        if(v.HasErrors() == false) {
          v.Required(self.EmailDoesNotExist()).Key("user.Email").Message("Email Already Exists")
        }
}

func (self *User) Id() string {
  return fmt.Sprintf("user:%s", self.Email)
}

func (self *User) EmailDoesNotExist() bool {
  c, err := redis.Dial("tcp", ":6379")
  if err != nil {
    panic("data access problem")
  }
  exists, err := redis.Bool(c.Do("EXISTS", self.Id()))
  if err != nil {
    panic("data access problem")
  }
  if exists {
     return false
  }
  return true
}

func (self User) Insert() bool {
// TODO
// add email regex
// add redis to config somewhere

  c, err := redis.Dial("tcp", ":6379")
  if err != nil {
    panic("data access problem")
  }
  _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(&self)...)
  if errb != nil {
    panic("data access problem")
  }

  return true
}
