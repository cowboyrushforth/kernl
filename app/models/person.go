package models

import "kernl/app/lib/redishandle"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "github.com/cowboyrushforth/go-webfinger"
import "github.com/cowboyrushforth/go-webfinger/jrd"
import "encoding/base64"
import "strings"
import "errors"

const DIASPORA_ACCOUNT_TYPE = 2
const PUMPIO_ACCOUNT_TYPE = 1

type Person struct {
  AccountType  int
  RemoteGuid   string
  DisplayName  string
  ImageSrc     string
  PodUrl       string
  ProfileUrl   string
  DialbackUrl  string
  InboxUrl     string
  OutboxUrl    string
  FollowersUrl string
  FollowingUrl string
  FavoritesUrl string
  ListsUrl     string
  Email        string
  RSAPubKey    string
  AccountIdentifier string
}

func PersonFromUid(uid string) (*Person, error) {
  person := Person{}
  v, errb := redis.Values(redishandle.Do("HGETALL", uid))
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

func PersonFromGuid(guid string) (*Person, error) {
    // lookup account identifier from guid
    result, err := redishandle.Do("GET", redis.Args{}.Add("guid:"+guid)...)
    if err != nil {
      panic(err)
    }
    uid, _ := redis.String(result, nil)
    if uid == "" {
      return nil, errors.New("person not found")
    }
    return PersonFromUid(uid)
}

func (self *Person) Id() string {
   return "person:" + self.AccountIdentifier
}

func (self *Person) PostsKey() string {
  return "posts:"+self.AccountIdentifier
}

func (self *Person) ConnectionsKey() string {
  return "connections:" + self.AccountIdentifier
}

func (self *Person) ConnectionsFollowingKey() string {
  return "connections:following:" + self.AccountIdentifier
}

func (self *Person) ConnectionsFollowersKey() string {
  return "connections:followers:" + self.AccountIdentifier
}

func (self *Person) ConnectionsBlockedKey() string {
  return "connections:blocked:" + self.AccountIdentifier
}

func (self *Person) ConnectionsMutualKey() string {
  return "connections:mutual:" + self.AccountIdentifier
}

func (self *Person) Validate(v *revel.Validation) {

  v.Check(self.DisplayName, 
    revel.Required{},
  ).Key("person.DisplayName")
  
  v.Check(self.AccountType,
    revel.Required{},
  ).Key("person.AccountType")


  if self.AccountType == DIASPORA_ACCOUNT_TYPE {
      v.Check(self.RemoteGuid, 
      revel.Required{},
    ).Key("person.RemoteGuid")

    v.Check(self.PodUrl, 
      revel.Required{},
    ).Key("person.PodUrl")

    v.Check(self.RSAPubKey, 
     revel.Required{},
    ).Key("person.RSAPubKey")
  }

  v.Check(self.ProfileUrl, 
    revel.Required{},
  ).Key("person.ProfileUrl")


  v.Check(self.AccountIdentifier, 
    revel.Required{},
  ).Key("person.AccountIdentifier")

}

// Connect initiates and synchronously
// performs a sharing notification (outbound connection)
// and datastore write
func (self *Person) Connect(user *User) (error) {
  pem_key, _ := base64.StdEncoding.DecodeString(self.RSAPubKey)
  self.RSAPubKey = string(pem_key)

  if self.AccountType == DIASPORA_ACCOUNT_TYPE {
    result, err := Diaspora_SendSharingNotification(user, self)
    if err != nil {
      panic(err)
    }
    if result.StatusCode != 200 && result.StatusCode != 202 {
      panic("received: "+result.Status)
    }
  } else if self.AccountType == PUMPIO_ACCOUNT_TYPE {
    // XXX todo
  }

  if self.Insert() {
    user.AddConnection(self, false, true)
    return nil
  } else {
    panic("could not save")
  }

  return nil
}

func (self *Person) AddFollower(user *User) {
  user.AddConnection(self, false, true)
}

func (self *Person) Insert() bool {
  self.DisplayName = strings.Replace(self.DisplayName, "acct:", "", 1)
  self.AccountIdentifier = strings.Replace(self.AccountIdentifier, "acct:", "", 1)
  // if we have not unpacked this yet do so now
  if (len(self.RSAPubKey) > 0) && (strings.Contains(self.RSAPubKey, "BEGIN PUBLIC KEY") == false) {
    pem_key, _ := base64.StdEncoding.DecodeString(self.RSAPubKey)
    self.RSAPubKey = string(pem_key)
  }

  // sanity check so we only insert or upsert ourselves
  result, err := redishandle.Do("GET", redis.Args{}.Add("guid:"+self.RemoteGuid)...)
  if err != nil {
    panic(err)
  }
  identifier, _ := redis.String(result, nil)
  if identifier == "" {
    _, erra := redishandle.Do("SET", redis.Args{}.Add("guid:"+self.RemoteGuid).Add(self.Id())...)
    if erra != nil {
      panic("data access problem")
    }
  }

  // if its a new user,
  // or if it matches a person
  if (identifier == "") || 
      (identifier == self.Id()) {

    _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
    if errb != nil {
      panic(errb)
    }
    return true

  } else {
    // this guid is already one of our users
    // so lets update/create their person row
    if(identifier[:5] == "user:") {
      revel.INFO.Println("ident", identifier)
      user, errc := UserFromUid(identifier)
      if errc != nil {
        panic(errc)
      }
      if user.AccountIdentifier == self.AccountIdentifier {
        revel.INFO.Println(user.AccountIdentifier, self.AccountIdentifier)
        _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
        if errb != nil {
          panic(errb)
        }
        return true
      }
    }
  }

  return false
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

// returns a local profile page for the user
// even if they are remote
func (self *Person) LocalUrl() string {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  if self.IsLocal() {
    return self.ProfileUrl
  } 
  return host_prefix + "/r/"+ self.RemoteGuid
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

  if len(resource.Subject) > 0 {
    person.AccountIdentifier = resource.Subject
    person.DisplayName = resource.Subject
  } else {
    person.AccountIdentifier = q
    person.DisplayName = q
  }

  profile_link := resource.GetLinkByRel("http://webfinger.net/rel/profile-page")
  if profile_link != nil {
    person.ProfileUrl = profile_link.Href
  }
  dialback_link := resource.GetLinkByRel("dialback")
  if dialback_link != nil {
    person.DialbackUrl = dialback_link.Href
  }
  inbox_link := resource.GetLinkByRel("activity-inbox")
  if inbox_link != nil {
    person.InboxUrl = inbox_link.Href
  }
  outbox_link := resource.GetLinkByRel("activity-outbox")
  if outbox_link != nil {
    person.OutboxUrl = outbox_link.Href
  }
  followers_link := resource.GetLinkByRel("followers")
  if followers_link != nil {
    person.FollowersUrl = followers_link.Href
  }
  following_link := resource.GetLinkByRel("following")
  if following_link != nil {
    person.FollowingUrl = following_link.Href
  }
  favorites_link := resource.GetLinkByRel("favorites")
  if favorites_link != nil {
    person.FavoritesUrl = favorites_link.Href
  }
  lists_link := resource.GetLinkByRel("lists")
  if lists_link != nil {
    person.ListsUrl = lists_link.Href
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

  if (len(person.DialbackUrl) > 0 && len(person.InboxUrl) > 0) {
    person.AccountType = PUMPIO_ACCOUNT_TYPE
    if len(person.RemoteGuid) == 0 {
      person.RemoteGuid = RandomString(32)
    }
  } else if (len(person.RemoteGuid) > 0 && len(person.PodUrl) > 0) {
    person.AccountType = DIASPORA_ACCOUNT_TYPE
  }

  return &person, nil
}
