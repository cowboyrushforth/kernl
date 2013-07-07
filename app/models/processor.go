package models

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "encoding/xml"
import "errors"

type XFlavor struct {
  XMLName xml.Name
  SenderHandle string `xml:"sender_handle"`
  RecipientHandle string `xml:"recipient_handle"`
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
  }
  return errors.New("flavor not understood")
}

// handle a request stanza, for when someone notifies
// they wish to start sharing with us
func HandleInboundRequest(c redis.Conn, user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundRequest, adding sender", xpkg.Post.Flavor.SenderHandle, "to", xpkg.Post.Flavor.RecipientHandle)
  user.AddConnection(c, sender, true, false)
  return errors.New("asdf")
}
