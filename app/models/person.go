package models

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "encoding/base64"
import "strings"
import "fmt"

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
// performs a sharing notification and
// datastore write
func (self *Person) Connect(c redis.Conn, user *User) (error) {
  result, err := SendSharingNotification(user, self)
  if err != nil {
      panic(err)
  }
  if result.StatusCode == 200 || result.StatusCode == 202 {
     if self.Insert(c) {
        user.AddConnection(self)
        return nil
     } else {
        panic("could not save")
     }
  } else {
      panic("received: "+result.Status)
  }
  return nil
}

func (self Person) Insert(c redis.Conn) bool {
  pem_key, _ := base64.StdEncoding.DecodeString(self.RSAPubKey)
  self.RSAPubKey = string(pem_key)
  self.DisplayName = strings.Replace(self.DisplayName, "acct:", "", 1)
  self.AccountIdentifier = strings.Replace(self.DisplayName, "acct:", "", 1)

  // TODO
  // wrap in redis multi
  _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(&self)...)
  if errb != nil {
    panic(errb)
  }

  return true
}
