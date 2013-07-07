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
  NotificationCount int
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

func (self *User) ConnectionsInboundKey() string {
  return fmt.Sprintf("connections:inbound:%s", self.Id())
}

func (self *User) ConnectionsOutboundKey() string {
  return fmt.Sprintf("connections:outbound:%s", self.Id())
}

func (self *User) ConnectionsBlockedKey() string {
  return fmt.Sprintf("connections:blocked:%s", self.Id())
}

func (self *User) ConnectionsMutualKey() string {
  return fmt.Sprintf("connections:mutual:%s", self.Id())
}

func (self *User) NotificationsKey() string {
  return fmt.Sprintf("notifications:%s", self.Id())
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
func (self *User) AddConnection(c redis.Conn, person *Person, inbound bool, outbound bool) {
  // XXX: score could be alpha ranking
  //      score could be timestamp of adding
  score := WordScore(person.DisplayName) 
  revel.INFO.Println("Adding Connection", person.AccountIdentifier, "to", self.AccountIdentifier,
                     "inbound", inbound, "outbound", outbound, "score", score)
  if inbound {
    _, err := c.Do("ZADD", redis.Args{}.Add(self.ConnectionsInboundKey()).Add(score).Add(person.AccountIdentifier)...)
    if err != nil {
      panic(err)
    }
  }
  if outbound {
    _, errb := c.Do("ZADD", redis.Args{}.Add(self.ConnectionsOutboundKey()).Add(score).Add(person.AccountIdentifier)...)
    if errb != nil {
      panic(errb)
    }
  }
}

type ConnectionEntry struct {
  AccountIdentifier string
  Person *Person
}

type ConnectionList struct {
  Connections []ConnectionEntry
}

func (self *User) AggregateConnections(c redis.Conn) {
  // TODO: optimize to not do this over frequently
  _, erra := c.Do("ZUNIONSTORE", redis.Args{}.Add(self.ConnectionsKey()).Add("2").
  Add(self.ConnectionsInboundKey()).Add(self.ConnectionsOutboundKey()).Add("AGGREGATE").Add("MIN")...)
  if erra != nil {
    panic(erra)
  }
}

func (self *User) IntersectConnections(c redis.Conn) {
  // TODO: optimize to not do this over frequently
  _, erra := c.Do("ZINTERSTORE", redis.Args{}.Add(self.ConnectionsMutualKey()).Add("2").
  Add(self.ConnectionsInboundKey()).Add(self.ConnectionsOutboundKey()).Add("AGGREGATE").Add("MIN")...)
  if erra != nil {
    panic(erra)
  }
}

func (self *User) ListConnections(c redis.Conn, inbound bool, outbound bool) ConnectionList {
    key := ""
    if inbound == true && outbound == true {
      self.AggregateConnections(c)
      key = self.ConnectionsKey()
    } else if inbound == true {
      key = self.ConnectionsInboundKey()
    } else if outbound == true {
      key = self.ConnectionsOutboundKey()
    }

    result, errb := c.Do("ZRANGE", redis.Args{}.Add(key).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(c, identifiers)
}

func (self *User) ListBlockedConnections(c redis.Conn) ConnectionList {
    result, errb := c.Do("ZRANGE", redis.Args{}.Add(self.ConnectionsBlockedKey()).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(c, identifiers)
}

func (self *User) ListMutualConnections(c redis.Conn) ConnectionList {
    self.IntersectConnections(c)
    result, errb := c.Do("ZRANGE", redis.Args{}.Add(self.ConnectionsMutualKey()).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(c, identifiers)
}

func (self *User) materializeConnectionList(c redis.Conn, identifiers []string) ConnectionList { 
    list := ConnectionList{}
    for _,element := range identifiers {
      person, err := PersonFromUid(c, "person:"+element)
      if err == nil {
        ce := ConnectionEntry{AccountIdentifier: element,
        Person: person}
        list.Connections = append(list.Connections, ce)
      }
    }
    return list
}

func (self *User) HasConnection(c redis.Conn, q string) bool {
    self.AggregateConnections(c)
    result, err := c.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsKey()).Add(q)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    }
    return false
}

func (self *User) SharesWithUser(c redis.Conn, account_identifier string) bool {
    result, err := c.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsOutboundKey()).Add(account_identifier)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    } 
    return false
}

func (self *User) IsSharedWithByUser(c redis.Conn, account_identifier string) bool {
    result, err := c.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsInboundKey()).Add(account_identifier)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    } 
    return false
}

