package models

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "github.com/cowboyrushforth/go-webfinger"
import "github.com/cowboyrushforth/go-webfinger/jrd"
import "encoding/base64"
import "strings"
import "fmt"
import "errors"

type Person struct {
  RemoteGuid   string
  DisplayName  string
  ImageSrc     string
  PodUrl       string
  ProfileUrl   string
  Email        string
  RSAPubKey    string
  AccountIdentifier string
}

func PersonFromUid(c redis.Conn, uid string) (*Person, error) {
  person := Person{}
  v, errb := redis.Values(c.Do("HGETALL", "person:"+uid))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &person)
  if errc != nil {
    return nil, errc
  }
  if len(person.AccountIdentifier) == 0 {
    return nil, errors.New("person not found")
  }

  return &person, nil
}

func (self *Person) Id() string {
   return fmt.Sprintf("person:%s", self.AccountIdentifier)
}

func (self *Person) Validate(c redis.Conn, v *revel.Validation) {

  v.Check(self.RemoteGuid, 
    revel.Required{},
  ).Key("person.RemoteGuid")

  v.Check(self.DisplayName, 
    revel.Required{},
  ).Key("person.DisplayName")

  v.Check(self.PodUrl, 
    revel.Required{},
  ).Key("person.PodUrl")

  v.Check(self.ProfileUrl, 
    revel.Required{},
  ).Key("person.ProfileUrl")

  v.Check(self.RSAPubKey, 
    revel.Required{},
  ).Key("person.RSAPubKey")

  v.Check(self.AccountIdentifier, 
    revel.Required{},
  ).Key("person.AccountIdentifier")

}

// Connect initiates and synchronously
// performs a sharing notification (outbound connection)
// and datastore write
func (self *Person) Connect(c redis.Conn, user *User) (error) {
  result, err := SendSharingNotification(user, self)
  if err != nil {
      panic(err)
  }
  if result.StatusCode == 200 || result.StatusCode == 202 {
     if self.Insert(c) {
        user.AddConnection(c,self, false, true)
        return nil
     } else {
        panic("could not save")
     }
  } else {
      panic("received: "+result.Status)
  }
  return nil
}

func (self *Person) Insert(c redis.Conn) bool {
  pem_key, _ := base64.StdEncoding.DecodeString(self.RSAPubKey)
  self.RSAPubKey = string(pem_key)
  self.DisplayName = strings.Replace(self.DisplayName, "acct:", "", 1)
  self.AccountIdentifier = strings.Replace(self.AccountIdentifier, "acct:", "", 1)

  _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
  if errb != nil {
    panic(errb)
  }

  return true
}

func (self *Person) ThumbnailUrl() string {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  if len(self.ImageSrc) > 0 {
    return self.ImageSrc
  } else {
    return host_prefix + "/static/images/default_profile_32.jpg"
  }
}

func (self *Person) IsLocal() bool {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  if self.PodUrl == host_prefix {
    return true
  }
  return false
}

func PersonFromWebFinger(q string) (*Person, error) {
  revel.INFO.Println("about to finger:", q)

  person := Person{}
  client := webfinger.NewClient(nil)

  insecure_fingering := revel.Config.BoolDefault("auth.allow_unsecure_fingering", false)

  err := error(nil)
  resource := &jrd.JRD{}
  if insecure_fingering {
    resource, err = client.LookupInsecure(q, []string{})
  } else {
    resource, err = client.Lookup(q, []string{})
  }
  if err != nil {
    return nil, err
  }

  person.AccountIdentifier = resource.Subject
  person.DisplayName = resource.Subject
  profile_link := resource.GetLinkByRel("http://webfinger.net/rel/profile-page")
  if profile_link != nil {
    person.ProfileUrl = profile_link.Href
  }
  seed_link := resource.GetLinkByRel("http://joindiaspora.com/seed_location")
  if seed_link != nil {
    person.PodUrl = seed_link.Href
  }
  guid_link := resource.GetLinkByRel("http://joindiaspora.com/guid")
  if guid_link != nil {
    person.RemoteGuid = guid_link.Href
  }
  pubkey_link := resource.GetLinkByRel("diaspora-public-key")
  if pubkey_link != nil {
    person.RSAPubKey = pubkey_link.Href
  }

  return &person, nil
}
