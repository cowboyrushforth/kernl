package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "net/url"

type Salmon struct {
  Kernl
}

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
  verified_payload, err := models.ParseVerifiedSalmonPayload(rc, user, sane_xml)
  if err != nil {
    c.Response.Status = 400
    return c.RenderText("")
  }

  revel.INFO.Println("verified payload:")
  revel.INFO.Println(verified_payload)
  // TODO: process the payload, take needed actions

  // TODO: return 202 if we had this item already
  c.Response.Status = 200
  return c.RenderText("")
}
