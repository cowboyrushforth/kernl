package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "net/url"

type Salmon struct {
  Kernl
}

// receive an encrypted salmon message
func (c Salmon) Receive(guid string, xml string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()

  user, err := models.UserFromGuid(rc, guid)
  if err != nil {
    return c.NotFound("User Not Found")
  }
  revel.INFO.Println("received salmon slap for guid", guid)

  // urldecode xml var
  sane_xml, _ := url.QueryUnescape(xml)

  // parse the xml, and verify it
  // against the senders pubkey
  // and the users privkey
  sender, verified_payload, err := models.ParseVerifiedSalmonPayload(rc, user, sane_xml)
  if err != nil {
    c.Response.Status = 400
    return c.RenderText("")
  }

  revel.INFO.Println("verified payload:")
  revel.INFO.Println(verified_payload)

  perr := models.ParseAndProcessVerifiedPayload(rc, user, sender, verified_payload)
  if perr != nil {
    revel.INFO.Println("could not process payload")
    c.Response.Status = 400
    return c.RenderText("")
  }

  // TODO: return 202 if we had this item already
  c.Response.Status = 200
  return c.RenderText("")
}

// receive a public salmon message
// same as above but user may not exist yet
func (c Salmon) ReceivePublic(xml string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  // urldecode xml var
  sane_xml, _ := url.QueryUnescape(xml)

  // parse the xml, and verify it
  // against the senders pubkey
  // and the users privkey
  sender, verified_payload, err := models.ParsePublicVerifiedSalmonPayload(rc, sane_xml)
  if err != nil {
    c.Response.Status = 400
    return c.RenderText("")
  }
  revel.INFO.Println("sender", sender.AccountIdentifier)
  revel.INFO.Println("verified payload", verified_payload)

  perr := models.ParseAndProcessVerifiedPayload(rc, nil, sender, verified_payload)
  if perr != nil {
    revel.INFO.Println("could not process payload")
    c.Response.Status = 400
    return c.RenderText("")
  }

  // TODO: return 202 if we had this item already
  c.Response.Status = 200
  return c.RenderText("")
}
