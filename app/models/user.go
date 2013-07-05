package models

import "fmt"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "errors"
import "crypto/rand"
import "crypto/rsa"
import "crypto/x509"
import "encoding/pem"
import "regexp"

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
    return self.Slug
  }
}

func (self *User) Validate(c redis.Conn, v *revel.Validation) {

  var emailPattern = regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")

  // check email sanity
  v.Check(self.Email, 
    revel.Required{},
    revel.MaxSize{100},
    revel.MinSize{5},
    revel.Email{revel.Match{emailPattern}},
    EmailDoesNotExist{c},
  ).Key("user.Email")

  // check slug sanity
  v.Check(self.Slug, 
    revel.Required{},
    revel.MaxSize{64},
    revel.MinSize{2},
    SlugDoesNotExist{c},
  ).Key("user.Slug")

}

func (self *User) Id() string {
  return fmt.Sprintf("user:%s", self.Slug)
}

func (self User) Insert(c redis.Conn) bool {
  // TODO
  // add email regex
  // wrap in redis multi

  pk, _ := rsa.GenerateKey(rand.Reader, 2048)
  self.RSAKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}))
  suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")
  self.AccountIdentifier = self.Slug + "@" + suffix
  self.Guid = RandomString(15)
  self.DisplayName = self.Slug

  _, erra := c.Do("SET", redis.Args{}.Add("email:"+self.Email).Add(self.Id())...)
  if erra != nil {
    panic("data access problem")
  }
  _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(&self)...)
  if errb != nil {
    panic("data access problem")
  }

  return true
}
