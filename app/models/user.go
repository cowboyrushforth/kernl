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
import "strings"

type User struct {
  DisplayName  string
  Email        string
  Slug         string
  PwdHash     []byte
  RSAKey      string
  Guid        string
  AccountIdentifier string
}

func UserFromUid(c redis.Conn, uid string) (*User, error) {
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

func UserFromSlug(c redis.Conn, slug string) (*User, error) {
    return UserFromUid(c, "user:"+slug)
}

func UserFromGuid(c redis.Conn, guid string) (*User, error) {
    // lookup account identifier from guid
    result, err := c.Do("GET", redis.Args{}.Add("guid:"+guid)...)
    if err != nil {
      panic(err)
    }
    slug, _ := redis.String(result, nil)
    if slug == "" {
      return nil, errors.New("user not found")
    }
    return UserFromSlug(c, strings.Replace(slug, "user:", "", 1))
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

func (self *User) ConnectionsKey() string {
  return fmt.Sprintf("connections:%s", self.Id())
}

func (self *User) Insert(c redis.Conn) bool {
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
  _, errb := c.Do("SET", redis.Args{}.Add("guid:"+self.Guid).Add(self.Id())...)
  if errb != nil {
    panic("data access problem")
  }
  _, errc := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
  if errc != nil {
    panic("data access problem")
  }

  return true
}

// returns users public key in der
// encoded format
func (self *User) RSAPubKey() string {
  // decode private key
  p, _ := pem.Decode([]byte(self.RSAKey))
  if p == nil {
    panic("could not parse private key")
  }
  // parse private key 
  key, _ := x509.ParsePKCS1PrivateKey(p.Bytes)
  // marshal public portion
  bytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
  // encode into pem format
  return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bytes}))
}

// marks in redis that we desire to share with 
// said person
func (self *User) AddConnection(c redis.Conn, person *Person) {
  // XXX: score could be alpha ranking
  //      score could be timestamp of adding
  revel.INFO.Println("Adding Connection", person.AccountIdentifier, "to", self.AccountIdentifier)
  score := 1
  _, err := c.Do("ZADD", redis.Args{}.Add(self.ConnectionsKey()).Add(score).Add(person.AccountIdentifier)...)
  if err != nil {
    panic(err)
  }
}

type ConnectionEntry struct {
  AccountIdentifier string
  Person *Person
}

type ConnectionList struct {
  Connections []ConnectionEntry
}

func (self *User) ListConnections(c redis.Conn) ConnectionList {
    list := ConnectionList{}
    result, err := c.Do("ZRANGE", redis.Args{}.Add(self.ConnectionsKey()).Add(0).Add(-1)...)
    if err != nil {
      panic(err)
    }
    identifiers, _ := redis.Strings(result, nil)
    for _,element := range identifiers {
      person, err := PersonFromUid(c, element)
      if err == nil {
        ce := ConnectionEntry{AccountIdentifier: element,
        Person: person}
        list.Connections = append(list.Connections, ce)
      }
    }
    return list
}

func (self *User) HasConnection(c redis.Conn, q string) bool {
    result, err := c.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsKey()).Add(q)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Bool(result, nil)
    return b
}

