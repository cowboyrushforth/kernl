package models

import "kernl/app/lib/redishandle"
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

func UserFromUid(uid string) (*User, error) {
  user := User{}
  v, errb := redis.Values(redishandle.Do("HGETALL", uid))
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

func UserFromSlug(slug string) (*User, error) {
    return UserFromUid("user:"+slug)
}

func UserFromGuid(guid string) (*User, error) {
    // lookup account identifier from guid
    result, err := redishandle.Do("GET", redis.Args{}.Add("guid:"+guid)...)
    if err != nil {
      panic(err)
    }
    slug, _ := redis.String(result, nil)
    if slug == "" {
      return nil, errors.New("user not found")
    }
    return UserFromSlug(strings.Replace(slug, "user:", "", 1))
}

func (self *User) String() string {
  if len(self.DisplayName) > 0 {
    return self.DisplayName
  } else {
    return self.Slug
  }
}

func (self *User) Validate(v *revel.Validation) {

  var emailPattern = regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")

  // check email sanity
  v.Check(self.Email, 
    revel.Required{},
    revel.MaxSize{100},
    revel.MinSize{5},
    revel.Email{revel.Match{emailPattern}},
    EmailDoesNotExist{},
  ).Key("user.Email")

  // check slug sanity
  v.Check(self.Slug, 
    revel.Required{},
    revel.MaxSize{64},
    revel.MinSize{2},
    SlugDoesNotExist{},
  ).Key("user.Slug")

}

func (self *User) Id() string {
  return "user:" + self.Slug
}

func (self *User) ConnectionsKey() string {
  return "connections:" + self.AccountIdentifier
}

func (self *User) ConnectionsInboundKey() string {
  return "connections:inbound:" + self.AccountIdentifier
}

func (self *User) ConnectionsOutboundKey() string {
  return "connections:outbound:" + self.AccountIdentifier
}

func (self *User) ConnectionsBlockedKey() string {
  return "connections:blocked:" + self.AccountIdentifier
}

func (self *User) ConnectionsMutualKey() string {
  return "connections:mutual:" + self.AccountIdentifier
}

func (self *User) NotificationsKey() string {
  return "notifications:" + self.Id()
}

func (self *User) PostsKey() string {
    return "posts:"+self.AccountIdentifier
}

func (self *User) Insert() bool {
  // TODO
  // add email regex
  // wrap in redis multi

  pk, _ := rsa.GenerateKey(rand.Reader, 2048)
  self.RSAKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}))
  suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")
  self.AccountIdentifier = self.Slug + "@" + suffix
  self.Guid = RandomString(15)
  self.DisplayName = self.Slug

  _, erra := redishandle.Do("SET", redis.Args{}.Add("email:"+self.Email).Add(self.Id())...)
  if erra != nil {
    panic(erra)
  }
  _, errb := redishandle.Do("SET", redis.Args{}.Add("guid:"+self.Guid).Add(self.Id())...)
  if errb != nil {
    panic("data access problem")
  }
  _, errc := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
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
func (self *User) AddConnection(person *Person, inbound bool, outbound bool) {
  // XXX: score could be alpha ranking
  //      score could be timestamp of adding
  // XXX: wrap in multi
  score := WordScore(person.DisplayName) 
  revel.INFO.Println("Adding Connection", person.AccountIdentifier, "to", self.AccountIdentifier,
                     "inbound", inbound, "outbound", outbound, "score", score)
  if inbound {
    _, erra := redishandle.Do("ZADD", redis.Args{}.Add(self.ConnectionsInboundKey()).Add(score).Add(person.AccountIdentifier)...)
    if erra != nil {
      panic(erra)
    }
    _, errb := redishandle.Do("ZADD", redis.Args{}.Add(person.ConnectionsOutboundKey()).Add(score).Add(self.AccountIdentifier)...)
    if errb != nil {
      panic(errb)
    }
  }
  if outbound {
    _, errc := redishandle.Do("ZADD", redis.Args{}.Add(self.ConnectionsOutboundKey()).Add(score).Add(person.AccountIdentifier)...)
    if errc != nil {
      panic(errc)
    }
    _, errd := redishandle.Do("ZADD", redis.Args{}.Add(person.ConnectionsInboundKey()).Add(score).Add(self.AccountIdentifier)...)
    if errd != nil {
      panic(errd)
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

func (self *User) AggregateConnections() {
  // TODO: optimize to not do this over frequently
  _, erra := redishandle.Do("ZUNIONSTORE", redis.Args{}.Add(self.ConnectionsKey()).Add("2").
  Add(self.ConnectionsInboundKey()).Add(self.ConnectionsOutboundKey()).Add("AGGREGATE").Add("MIN")...)
  if erra != nil {
    panic(erra)
  }
}

func (self *User) IntersectConnections() {
  // TODO: optimize to not do this over frequently
  _, erra := redishandle.Do("ZINTERSTORE", redis.Args{}.Add(self.ConnectionsMutualKey()).Add("2").
  Add(self.ConnectionsInboundKey()).Add(self.ConnectionsOutboundKey()).Add("AGGREGATE").Add("MIN")...)
  if erra != nil {
    panic(erra)
  }
}

func (self *User) ListConnections(inbound bool, outbound bool) ConnectionList {
    key := ""
    if inbound == true && outbound == true {
      self.AggregateConnections()
      key = self.ConnectionsKey()
    } else if inbound == true {
      key = self.ConnectionsInboundKey()
    } else if outbound == true {
      key = self.ConnectionsOutboundKey()
    }

    result, errb := redishandle.Do("ZRANGE", redis.Args{}.Add(key).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(identifiers)
}

func (self *User) ListBlockedConnections() ConnectionList {
    result, errb := redishandle.Do("ZRANGE", redis.Args{}.Add(self.ConnectionsBlockedKey()).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(identifiers)
}

func (self *User) ListMutualConnections() ConnectionList {
    self.IntersectConnections()
    result, errb := redishandle.Do("ZRANGE", redis.Args{}.Add(self.ConnectionsMutualKey()).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return self.materializeConnectionList(identifiers)
}

func (self *User) materializeConnectionList(identifiers []string) ConnectionList { 
    list := ConnectionList{}
    for _,element := range identifiers {
      person, err := PersonFromUid("person:"+element)
      if err == nil {
        ce := ConnectionEntry{AccountIdentifier: element,
        Person: person}
        list.Connections = append(list.Connections, ce)
      }
    }
    return list
}

func (self *User) HasConnection(q string) bool {
    self.AggregateConnections()
    result, err := redishandle.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsKey()).Add(q)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    }
    return false
}

func (self *User) HasOutboundConnection(q string) bool {
    result, err := redishandle.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsOutboundKey()).Add(q)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    }
    return false
}

func (self *User) SharesWithUser(account_identifier string) bool {
    result, err := redishandle.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsOutboundKey()).Add(account_identifier)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    } 
    return false
}

func (self *User) IsSharedWithByUser(account_identifier string) bool {
    result, err := redishandle.Do("ZSCORE", redis.Args{}.Add(self.ConnectionsInboundKey()).Add(account_identifier)...)
    if err != nil {
      panic(err)
    }
    b, _ := redis.Int(result, nil)
    if b > 0 {
      return true
    } 
    return false
}

// find or create a Person for this user
func (self *User) Person() (*Person) {
  person, err := PersonFromUid("person:"+self.AccountIdentifier)
  if err != nil {
     // we appear to not have this person.
     // try to finger them.
     err = nil
     person, err = PersonFromWebFinger(self.AccountIdentifier)
     if err != nil {
       panic("can not locate person")
     }
     person.Insert()
   }

   return person
}

