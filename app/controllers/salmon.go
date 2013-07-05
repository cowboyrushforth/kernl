package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "net/url"

type Salmon struct {
  Kernl
}

func (c Salmon) Receive() revel.Result {

  guid := c.Params.Get("guid")
  raw_xml := c.Params.Get("xml")
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()

  user, err := models.UserFromGuid(rc, guid)
  if err != nil {
    return c.NotFound("User Not Found")
  }
  revel.INFO.Println("received salmon slap for guid", guid)
  sane_xml, _ := url.QueryUnescape(raw_xml)

  verified_payload, err := models.ParseVerifiedSalmonPayload(rc, user, sane_xml)

  if err != nil {
    c.Response.Status = 400
    return c.RenderText("")
  }

  revel.INFO.Println("verified payload:")
  revel.INFO.Println(verified_payload)

  c.Response.Status = 200
  return c.RenderText("")
}
