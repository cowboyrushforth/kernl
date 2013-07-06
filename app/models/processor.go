package models

import "github.com/robfig/revel"
import "encoding/xml"

type XFlavor struct {
  XMLName xml.Name
}
type XPost struct {
  Flavor XFlavor `xml:",any"`
}
type XPackage struct {
  Post XPost `xml:"post"`
}

func ParseAndProcessVerifiedPayload(user *User, sender *Person, payload string) {
  v := XPackage{}
  err := xml.Unmarshal([]byte(payload), &v )
  if err != nil {
    panic(err)
    return
  }
  flavor := v.Post.Flavor.XMLName.Local
  revel.INFO.Println("message flavor", flavor)

  // TODO - now that we have a flavor
}
