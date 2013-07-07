package models

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "encoding/xml"
import "errors"

type XFlavor struct {
  XMLName xml.Name
  /* request */
  SenderHandle string `xml:"sender_handle"`
  RecipientHandle string `xml:"recipient_handle"`

  /* profile */
  DiasporaHandle string `xml:"diaspora_handle"`
  Searchable string `xml:"searchable"`
  ImageUrl string `xml:"image_url"`
  Nsfw bool `xml:"nsfw"`
  TagString string `xml:"tag_string"`

}
type XPost struct {
  Flavor XFlavor `xml:",any"`
}
type XPackage struct {
  Post XPost `xml:"post"`
}

func ParseAndProcessVerifiedPayload(c redis.Conn, user *User, sender *Person, payload string) error {
  v := XPackage{}
  err := xml.Unmarshal([]byte(payload), &v )
  if err != nil {
    panic(err)
  }
  flavor := v.Post.Flavor.XMLName.Local
  switch flavor {
  case "request":
    return HandleInboundRequest(c, user, sender, v)
  case "profile":
    return HandleInboundProfile(c, user, sender, v)
  case "status_message":
    return HandleInboundStatusMessage(c, sender, v)
  case "participation":
    return HandleInboundParticipation(c, sender, v)
  }
  return errors.New("flavor not understood")
}

// handle a request stanza, for when someone notifies
// they wish to start sharing with us
func HandleInboundRequest(c redis.Conn, user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundRequest, adding sender", xpkg.Post.Flavor.SenderHandle, "to", xpkg.Post.Flavor.RecipientHandle)
  user.AddConnection(c, sender, true, false)
  SendNotification(user, c, "share_started", sender.RemoteGuid)
  return nil
}

// handle a profile stanza, unclear
// how important this is at the moment
func HandleInboundProfile(c redis.Conn, user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundProfile, diaspora profile", xpkg.Post.Flavor.DiasporaHandle)
  _, person_err := PersonFromUid(c, "person:"+xpkg.Post.Flavor.DiasporaHandle)
  if person_err != nil {
    // we appear to not have this person.
    // try to finger them.
    person, person_err := PersonFromWebFinger(xpkg.Post.Flavor.DiasporaHandle)
    if person_err != nil {
      panic("can not locate user")
    }
    person.Insert(c)
  }
  return nil
}

func HandleInboundStatusMessage(c redis.Conn, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundStatusMessage")
  return nil
}

func HandleInboundParticipation(c redis.Conn, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundParticipation") 
  return nil
}
