package models

import "github.com/robfig/revel"
import "encoding/xml"
import "errors"
import "time"

type XFlavor struct {
  XMLName xml.Name
  /* request */
  SenderHandle string `xml:"sender_handle"`
  RecipientHandle string `xml:"recipient_handle"`

  /* profile */
  Searchable string `xml:"searchable"`
  ImageUrl string `xml:"image_url"`
  Nsfw bool `xml:"nsfw"`
  TagString string `xml:"tag_string"`

  /* post */
  RawMessage string `xml:"raw_message"`
  Public bool `xml:"public"`
  CreatedAt string `xml:"created_at"`

  /* shared */
  Guid string `xml:"guid"`
  DiasporaHandle string `xml:"diaspora_handle"`
  
  /* comment */
  ParentGuid string `xml:"parent_guid"`
  AuthorSignature string `xml:"author_signature"`
  Text string `xml:"text"`
}

type XPost struct {
  Flavor XFlavor `xml:",any"`
}

type XPackage struct {
  Post XPost `xml:"post"`
}

func ParseAndProcessVerifiedPayload(user *User, sender *Person, payload string) error {
  v := XPackage{}
  err := xml.Unmarshal([]byte(payload), &v )
  if err != nil {
    panic(err)
  }
  flavor := v.Post.Flavor.XMLName.Local
  switch flavor {
  case "request":
    return HandleInboundRequest(user, sender, v)
  case "profile":
    return HandleInboundProfile(user, sender, v)
  case "status_message":
    return HandleInboundStatusMessage(sender, v)
  case "participation":
    return HandleInboundParticipation(sender, v)
  case "comment":
    return HandleInboundComment(user, sender, v)
  }
  return errors.New("flavor not understood")
}

// handle a request stanza, for when someone notifies
// they wish to start sharing with us
func HandleInboundRequest(user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundRequest, adding sender", xpkg.Post.Flavor.SenderHandle, "to", xpkg.Post.Flavor.RecipientHandle)
  user.AddConnection(sender, true, false)
  SendNotification(user, "share_started", sender.RemoteGuid)
  return nil
}

// handle a profile stanza, unclear
// how important this is at the moment
func HandleInboundProfile(user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundProfile, diaspora profile", xpkg.Post.Flavor.DiasporaHandle)
  _, person_err := PersonFromUid("person:"+xpkg.Post.Flavor.DiasporaHandle)
  if person_err != nil {
    // we appear to not have this person.
    // try to finger them.
    person, person_err := PersonFromWebFinger(xpkg.Post.Flavor.DiasporaHandle)
    if person_err != nil {
      panic("can not locate user")
    }
    person.Insert()
  }
  return nil
}

func HandleInboundStatusMessage(sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundStatusMessage")
  ts,_ := time.Parse("2006-01-02 15:04:05 MST", xpkg.Post.Flavor.CreatedAt)
  post := Post{
    DisplayName: sender.DisplayName,
    Message: xpkg.Post.Flavor.RawMessage,
    Guid: xpkg.Post.Flavor.Guid,
    AccountIdentifier: xpkg.Post.Flavor.DiasporaHandle,
    Public: xpkg.Post.Flavor.Public,
    CreatedAt: ts.Unix(),
  }
  post.Insert(sender)
  return nil
}

func HandleInboundParticipation(sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundParticipation") 
  return nil
}

func HandleInboundComment(user *User, sender *Person, xpkg XPackage) error {
  revel.INFO.Println("HandleInboundComment") 
  comment := Comment{
        DisplayName: sender.DisplayName,
        Text: xpkg.Post.Flavor.Text,
        Guid: xpkg.Post.Flavor.Guid,
        ParentGuid: xpkg.Post.Flavor.ParentGuid,
        AccountIdentifier: xpkg.Post.Flavor.DiasporaHandle,
        CreatedAt: time.Now().Unix(),
  }
  comment.Insert(sender)
  return nil
}
