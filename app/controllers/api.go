package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "github.com/cowboyrushforth/goa1"
import "net/http"
import "encoding/json"
import "io"
import "bytes"
import "time"

type Api struct {
  Kernl
}

type FeedPost struct {
   Verb string `json:"verb"`
   ActivityObject models.ActivityObject `json:"object"`
}

func (c Api) IsVerifiedRequest(http *http.Request) bool {
  req, erra := goa1.ParseRequest(http) 
  if erra != nil {
    panic(erra)
  }

  client,errb := models.ClientFromConsumerKey(req.ConsumerKey)
  if errb != nil {
    panic(errb)
  }

  access_token, errc := models.AccessTokenFromToken(req.Token)
  ok, errc := goa1.Validate(req, client.Secret, access_token.TokenSecret)
  if errb != nil {
    panic(errc)
  }

  if ok  {
    c.RenderArgs["access_token"] = access_token
    return true
  }

  return false
}

func (c Api) Whoami() revel.Result {
  ok := c.IsVerifiedRequest(c.Request.Request)

  if ok {
    host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
    revel.INFO.Println("OK")
    return c.Redirect(host_prefix+"/api/user/"+c.RenderArgs["access_token"].(*models.AccessToken).Slug+"/profile")
  }

  revel.INFO.Println("NOT OK")
  c.Response.Status = 400
  return c.RenderText("FAIL")
}

func (c Api) Profile(slug string) revel.Result {
  if c.current_user() == nil {
    ok := c.IsVerifiedRequest(c.Request.Request)
    if ok == false {
      c.Response.Status = 400
      return c.RenderText("")
    }
  }
  user, err := models.UserFromSlug(slug)
  if err != nil {
    panic(err)
  }
  return c.RenderJson(user.ActivityObject())
}

func (c Api) FeedPost() revel.Result {
  var b bytes.Buffer
  var dest io.Writer = &b
  _,_ = io.Copy(dest, c.Request.Body)
  var payload FeedPost
  err := json.Unmarshal(b.Bytes(), &payload)
  if err != nil {
    panic(err)
  }
  revel.INFO.Println(payload)

  if len(payload.Verb) == 0 {
    payload.Verb = "post"
  }

  switch(payload.Verb) {
  case "follow":
    return c.handleFollow(&payload.ActivityObject)
  default:
    panic("verb unsupported")
  }

/*
//{follow {[] <nil>  rushforth rushforth [] acct:rushforth@lmac.com <nil> person    [] http://lmac.com/rushforth 0xc20026eb40 0xc200288540 0xc2002885d0 0xc2002883f0 0xc2002884b0 false {false false}}}
// if no actor, set actor based on login
// if the actor does not match login bail
// set default verb to post if no verb
// ensure recipients
// populate activity 
// save activity
// execute activity 
// add to inbox
// add to outbox
// render activity json
*/

  c.Response.Status = 400
  return c.RenderText("")
}

func (c Api) handleFollow(actor *models.ActivityObject) revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  activity_id := models.RandomString(32)
  act := models.Activity{
    Actor: actor,
    Id: activity_id,
    Object: actor,
    Published: time.Now().Format("2006-01-02T15:04:05.00-07:00"),
    Title: "foo followed foo",
    UpdatedAt: time.Now().Format("2006-01-02T15:04:05.00-07:00"),
    Url: host_prefix+"/activity"+activity_id}
  c.Response.Status = 200

  return c.RenderJson(act)
}
